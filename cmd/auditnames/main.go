// Command auditnames audits Promptful Go naming components and object bases.
//
// It parses every Git-visible Go source file with go/ast (so grouped
// const/var entries, struct fields, and interface methods are all seen),
// records each codebase-prefixed declared name, counts its semantic
// components, and confirms its object base is registered in
// docs-internal/ListObject.md. Registered XInstance chains from
// docs-internal/ListXInstance.md are exempt from the component limit.
//
// Names with three components are review findings; names with more than
// three are violations; a name whose base is not registered is a violation.
// Two reports are written under docs-work/audit. No source files are modified.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// LAuditRecord is one audited declaration.
type LAuditRecord struct {
	LAuditRecordPath       string
	LAuditRecordLine       int
	LAuditRecordName       string
	LAuditRecordBase       string
	LAuditRecordComponents int
	LAuditRecordStatus     string
}

func main() {
	root := "."
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-root" && i+1 < len(os.Args) {
			root = os.Args[i+1]
			i++
		}
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		LAuditFail("cannot resolve root %q: %v", root, err)
	}

	version, err := LVersionRead(abs)
	if err != nil {
		LAuditFail("%v", err)
	}
	bases, err := LAuditBaseRead(filepath.Join(abs, "docs-internal", "ListObject.md"))
	if err != nil {
		LAuditFail("%v", err)
	}
	if len(bases) == 0 {
		LAuditFail("no registered bases found in docs-internal/ListObject.md")
	}
	xchains := LAuditChainRead(filepath.Join(abs, "docs-internal", "ListXInstance.md"))

	files, err := LSourceFilesList(abs)
	if err != nil {
		LAuditFail("%v", err)
	}

	var records []LAuditRecord
	for _, rel := range files {
		recs, err := LFileScan(abs, rel, bases, xchains)
		if err != nil {
			// Report the unreadable file rather than silently skipping it.
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", rel, err)
			continue
		}
		records = append(records, recs...)
	}

	sort.SliceStable(records, func(a, b int) bool {
		if records[a].LAuditRecordPath != records[b].LAuditRecordPath {
			return records[a].LAuditRecordPath < records[b].LAuditRecordPath
		}
		return records[a].LAuditRecordLine < records[b].LAuditRecordLine
	})

	if err := LReportWrite(abs, version, records); err != nil {
		LAuditFail("%v", err)
	}
}

// LAuditFail prints a diagnostic and exits non-zero.
func LAuditFail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "auditnames: "+format+"\n", args...)
	os.Exit(1)
}

// LVersionRead returns the current-version field from version.json.
func LVersionRead(root string) (string, error) {
	path := filepath.Join(root, "version.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("version.json was not found: %s", path)
	}
	re := regexp.MustCompile(`"current-version"\s*:\s*"([^"]+)"`)
	m := re.FindSubmatch(data)
	if m == nil {
		return "", fmt.Errorf("version.json must contain a 'current-version' property")
	}
	version := string(m[1])
	if !regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`).MatchString(version) {
		return "", fmt.Errorf("invalid current-version in version.json: %q", version)
	}
	return version, nil
}

// LAuditBaseRead reads the authoritative base list from the fenced
// AUDIT-BASES block of ListObject.md.
func LAuditBaseRead(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read base registry: %s: %w", path, err)
	}
	lines := strings.Split(string(data), "\n")
	var bases []string
	inBlock, inFence := false, false
	word := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)
	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		switch {
		case strings.Contains(line, "AUDIT-BASES:BEGIN"):
			inBlock = true
		case strings.Contains(line, "AUDIT-BASES:END"):
			inBlock = false
		case inBlock && strings.HasPrefix(strings.TrimSpace(line), "```"):
			inFence = !inFence
		case inBlock && inFence:
			b := strings.TrimSpace(line)
			if word.MatchString(b) {
				bases = append(bases, b)
			}
		}
	}
	// Longest base first so word-boundary matching prefers the most specific.
	sort.SliceStable(bases, func(a, b int) bool { return len(bases[a]) > len(bases[b]) })
	return bases, nil
}

// LAuditChainRead loads registered XInstance chain names.
func LAuditChainRead(path string) map[string]bool {
	out := map[string]bool{}
	data, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	re := regexp.MustCompile("^##\\s+`?([PpLl](?:[Ss])?[A-Za-z0-9]+_[PpLl](?:[Ss])?[A-Za-z0-9]+)`?\\s*$")
	for _, raw := range strings.Split(string(data), "\n") {
		if m := re.FindStringSubmatch(strings.TrimRight(raw, "\r")); m != nil {
			out[strings.ToLower(m[1])] = true
		}
	}
	return out
}

// LSourceFilesList returns Git-visible Go files relative to root.
func LSourceFilesList(root string) ([]string, error) {
	cmd := exec.Command("git", "-C", root, "-c", "core.quotePath=false",
		"ls-files", "--cached", "--others", "--exclude-standard", "--", "*.go")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git could not enumerate source files: %w", err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// LFileScan parses one Go file and returns its audited declarations.
func LFileScan(root, rel string, bases []string, xchains map[string]bool) ([]LAuditRecord, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath.Join(root, rel), nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	var records []LAuditRecord
	record := func(name string, pos token.Pos) {
		if !LNamePrefixed(name) {
			return
		}
		rec := LNameEvaluate(name, bases, xchains)
		rec.LAuditRecordPath = rel
		rec.LAuditRecordLine = fset.Position(pos).Line
		records = append(records, rec)
	}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			record(d.Name.Name, d.Name.NamePos)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec: // var / const, including grouped blocks
					for _, id := range s.Names {
						record(id.Name, id.NamePos)
					}
				case *ast.TypeSpec:
					record(s.Name.Name, s.Name.NamePos)
					LAuditMemberList(s.Type, record)
				}
			}
		}
	}
	return records, nil
}

// LAuditMemberList records struct field and interface method names.
func LAuditMemberList(expr ast.Expr, record func(string, token.Pos)) {
	switch t := expr.(type) {
	case *ast.StructType:
		if t.Fields == nil {
			return
		}
		for _, field := range t.Fields.List {
			for _, id := range field.Names { // named (non-embedded) fields only
				record(id.Name, id.NamePos)
			}
		}
	case *ast.InterfaceType:
		if t.Methods == nil {
			return
		}
		for _, m := range t.Methods.List {
			for _, id := range m.Names {
				record(id.Name, id.NamePos)
			}
		}
	}
}

// LNamePrefixed reports whether a name carries an L/P/LS/PS codebase prefix.
func LNamePrefixed(name string) bool {
	n := strings.TrimLeft(name, "_")
	pfx, rest := LAuditPrefixSplit(n)
	return pfx != "" && rest != "" && unicode.IsUpper([]rune(rest)[0])
}

// LAuditPrefixSplit separates a codebase prefix from the remainder. The secondary
// marker S is only consumed when the following rune starts the base (upper).
func LAuditPrefixSplit(name string) (string, string) {
	r := []rune(name)
	if len(r) < 2 {
		return "", name
	}
	c0 := unicode.ToUpper(r[0])
	if c0 != 'L' && c0 != 'P' {
		return "", name
	}
	if unicode.ToUpper(r[1]) == 'S' && len(r) > 2 && unicode.IsUpper(r[2]) {
		return string(r[:2]), string(r[2:])
	}
	if unicode.IsUpper(r[1]) {
		return string(r[:1]), string(r[1:])
	}
	return "", name
}

// LNameEvaluate computes base, component count, and status for one name.
func LNameEvaluate(name string, bases []string, xchains map[string]bool) LAuditRecord {
	rec := LAuditRecord{LAuditRecordName: name}
	working := strings.TrimLeft(name, "_")

	// Registered XInstance chains are exempt from the component limit.
	if strings.Contains(working, "_") && xchains[strings.ToLower(working)] {
		rec.LAuditRecordBase = LAuditBaseMatch(working, bases)
		rec.LAuditRecordComponents = 0
		rec.LAuditRecordStatus = "Valid"
		return rec
	}

	base := LAuditBaseMatch(working, bases)
	rec.LAuditRecordBase = base
	_, rest := LAuditPrefixSplit(working)
	rec.LAuditRecordComponents = LAuditComponentGet(rest)

	switch {
	case base == "":
		rec.LAuditRecordBase = "UNKNOWN"
		rec.LAuditRecordStatus = "Violation"
	case rec.LAuditRecordComponents > 3:
		rec.LAuditRecordStatus = "Violation"
	case rec.LAuditRecordComponents == 3:
		rec.LAuditRecordStatus = "Review"
	default:
		rec.LAuditRecordStatus = "Valid"
	}
	return rec
}

// LAuditBaseMatch returns the longest registered base that the name's remainder
// starts with at a word boundary, or "" when none match.
func LAuditBaseMatch(name string, bases []string) string {
	_, rest := LAuditPrefixSplit(name)
	if rest == "" {
		return ""
	}
	r := []rune(rest)
	for _, b := range bases { // bases are longest-first
		if !strings.HasPrefix(rest, b) {
			continue
		}
		bl := len([]rune(b))
		if bl == len(r) || unicode.IsUpper(r[bl]) || unicode.IsDigit(r[bl]) {
			return b
		}
	}
	return ""
}

// LAuditComponentGet counts semantic components in a base-and-suffix string,
// splitting camelCase, acronym runs, and digit groups.
func LAuditComponentGet(s string) int {
	r := []rune(s)
	n := len(r)
	count := 0
	i := 0
	for i < n {
		switch {
		case unicode.IsDigit(r[i]):
			for i < n && unicode.IsDigit(r[i]) {
				i++
			}
			count++
		case unicode.IsUpper(r[i]):
			j := i + 1
			for j < n && unicode.IsUpper(r[j]) {
				j++
			}
			if j-i == 1 {
				for j < n && unicode.IsLower(r[j]) {
					j++
				}
				i = j
			} else if j < n && unicode.IsLower(r[j]) {
				i = j - 1 // trailing upper begins the next word
			} else {
				i = j
			}
			count++
		case unicode.IsLower(r[i]):
			for i < n && unicode.IsLower(r[i]) {
				i++
			}
			count++
		default:
			i++ // skip separators such as underscores
		}
	}
	return count
}

// LReportWrite writes the full and violation reports under docs-work/audit.
func LReportWrite(root, version string, records []LAuditRecord) error {
	dir := filepath.Join(root, "docs-work", "audit")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("cannot create audit directory: %w", err)
	}
	auditPath := filepath.Join(dir, "Names-"+version+".md")
	violationPath := filepath.Join(dir, "NamesViolated-"+version+".md")

	head := "| File | Line | Name | Base | Components | Status |\n|---|---:|---|---|---:|---|\n"
	row := func(rec LAuditRecord) string {
		return fmt.Sprintf("| %s | %d | %s | %s | %d | %s |\n",
			rec.LAuditRecordPath, rec.LAuditRecordLine, rec.LAuditRecordName,
			rec.LAuditRecordBase, rec.LAuditRecordComponents, rec.LAuditRecordStatus)
	}

	var all, violations strings.Builder
	fmt.Fprintf(&all, "# Promptful naming audit %s\n\n%s", version, head)
	fmt.Fprintf(&violations, "# Promptful naming violations %s\n\n%s", version, head)
	violationCount := 0
	for _, rec := range records {
		all.WriteString(row(rec))
		if rec.LAuditRecordStatus == "Violation" {
			violations.WriteString(row(rec))
			violationCount++
		}
	}
	if len(records) == 0 {
		all.WriteString("| _None_ | | | | | |\n")
	}
	if violationCount == 0 {
		violations.WriteString("| _None_ | | | | | |\n")
	}

	if err := LFileAtomicWrite(auditPath, all.String()); err != nil {
		return err
	}
	if err := LFileAtomicWrite(violationPath, violations.String()); err != nil {
		return err
	}

	fmt.Printf("Audit report:      %s\n", auditPath)
	fmt.Printf("Violations report: %s\n", violationPath)
	fmt.Printf("Names audited: %d\n", len(records))
	fmt.Printf("Violations: %d\n", violationCount)
	return nil
}

// LFileAtomicWrite writes UTF-8 without a BOM via a temporary file.
func LFileAtomicWrite(path, content string) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return fmt.Errorf("cannot write %s: %w", path, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("cannot finalize %s: %w", path, err)
	}
	return nil
}
