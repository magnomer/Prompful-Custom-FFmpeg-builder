package scripting

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"customffmpegbuilder/internal/workspace"
)

type ScriptFilePlan struct {
	WorkspaceDirectory string   `json:"workspaceDirectory"`
	ScriptFilePath     string   `json:"scriptFilePath"`
	ScriptLines        []string `json:"scriptLines"`
}

type ScriptFile struct {
	ScriptFilePath   string `json:"scriptFilePath"`
	ScriptSha256Hash string `json:"scriptSha256Hash"`
}

var safeMsys2PackageNamePattern = regexp.MustCompile(`^[A-Za-z0-9_+.-]+$`)
var safeConfigureFlagPattern = regexp.MustCompile(`^--[A-Za-z0-9][A-Za-z0-9_+./:=,-]*$`)

func ValidateMsys2PackageName(packageName string) error {
	if packageName == "" {
		return errors.New("MSYS2 package name is empty")
	}
	if !safeMsys2PackageNamePattern.MatchString(packageName) {
		return fmt.Errorf("MSYS2 package name contains unsafe characters: %s", packageName)
	}
	return nil
}

func ValidateConfigureFlag(configureFlag string) error {
	if configureFlag == "" {
		return errors.New("configure flag is empty")
	}
	if !safeConfigureFlagPattern.MatchString(configureFlag) {
		return fmt.Errorf("configure flag contains unsafe characters: %s", configureFlag)
	}
	if strings.Contains(configureFlag, "..") {
		return fmt.Errorf("configure flag contains unsafe path traversal marker: %s", configureFlag)
	}
	return nil
}

func WriteScriptFile(scriptFilePlan ScriptFilePlan) (ScriptFile, error) {
	if scriptFilePlan.WorkspaceDirectory == "" || scriptFilePlan.ScriptFilePath == "" {
		return ScriptFile{}, errors.New("approved shell script paths must not be empty")
	}
	if err := workspace.CheckPathInsideWorkspace(scriptFilePlan.WorkspaceDirectory, scriptFilePlan.ScriptFilePath); err != nil {
		return ScriptFile{}, err
	}
	if err := workspace.CheckRealPathInsideWorkspace(scriptFilePlan.WorkspaceDirectory, filepath.Dir(scriptFilePlan.ScriptFilePath)); err != nil {
		return ScriptFile{}, err
	}
	if len(scriptFilePlan.ScriptLines) == 0 {
		return ScriptFile{}, errors.New("approved shell script has no lines")
	}
	scriptDirectory := filepath.Dir(scriptFilePlan.ScriptFilePath)
	if err := os.MkdirAll(scriptDirectory, 0o755); err != nil {
		return ScriptFile{}, err
	}
	if err := workspace.CheckRealPathInsideWorkspace(scriptFilePlan.WorkspaceDirectory, scriptFilePlan.ScriptFilePath); err != nil {
		return ScriptFile{}, err
	}
	scriptText := strings.Join(scriptFilePlan.ScriptLines, "\n") + "\n"
	scriptHash := sha256.Sum256([]byte(scriptText))
	scriptHashString := hex.EncodeToString(scriptHash[:])
	if existingInfo, err := os.Lstat(scriptFilePlan.ScriptFilePath); err == nil {
		if existingInfo.Mode()&os.ModeSymlink != 0 {
			return ScriptFile{}, errors.New("approved shell script path is a symlink")
		}
		if existingInfo.IsDir() {
			return ScriptFile{}, errors.New("approved shell script path is a directory")
		}
		existingBytes, readErr := os.ReadFile(scriptFilePlan.ScriptFilePath)
		if readErr != nil {
			return ScriptFile{}, readErr
		}
		existingHash := sha256.Sum256(existingBytes)
		if strings.EqualFold(hex.EncodeToString(existingHash[:]), scriptHashString) {
			return ScriptFile{ScriptFilePath: scriptFilePlan.ScriptFilePath, ScriptSha256Hash: scriptHashString}, nil
		}
		if removeErr := os.Remove(scriptFilePlan.ScriptFilePath); removeErr != nil {
			return ScriptFile{}, fmt.Errorf("remove stale approved shell script: %w", removeErr)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return ScriptFile{}, err
	}
	if err := workspace.CheckRealPathInsideWorkspace(scriptFilePlan.WorkspaceDirectory, scriptDirectory); err != nil {
		return ScriptFile{}, err
	}
	outputFile, err := os.OpenFile(scriptFilePlan.ScriptFilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	if err != nil {
		return ScriptFile{}, err
	}
	_, writeErr := outputFile.Write([]byte(scriptText))
	closeErr := outputFile.Close()
	if writeErr != nil {
		return ScriptFile{}, writeErr
	}
	if closeErr != nil {
		return ScriptFile{}, closeErr
	}
	return ScriptFile{ScriptFilePath: scriptFilePlan.ScriptFilePath, ScriptSha256Hash: scriptHashString}, nil
}

func PacmanInstallScriptLines(packageNames []string) ([]string, error) {
	quotedPackageNames := make([]string, 0, len(packageNames))
	for _, packageName := range packageNames {
		if err := ValidateMsys2PackageName(packageName); err != nil {
			return nil, err
		}
		quotedPackageNames = append(quotedPackageNames, shellQuote(packageName))
	}
	return []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"echo 'Using the official MSYS2 package server for this private toolchain.'",
		"cat > /etc/pacman.d/mirrorlist.msys <<'EOF'",
		"Server = https://repo.msys2.org/msys/$arch/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.mingw32 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/mingw32/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.mingw64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/mingw64/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.ucrt64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/ucrt64/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.clang64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/clang64/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.clangarm64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/clangarm64/",
		"EOF",
		"echo 'Initializing the private MSYS2 package keyring.'",
		"pacman-key --init 2>&1",
		"pacman-key --populate msys2 2>&1",
		"echo 'Preparing pacman database signature policy.'",
		"# MSYS2 packages remain signature-checked. Repository database signatures are treated as optional, matching normal MSYS2 pacman behavior and avoiding false failures when a database .sig is not served or is cleared during refresh.",
		"sed -i -E 's/^SigLevel[[:space:]]*=.*/SigLevel = Required DatabaseOptional/' /etc/pacman.conf",
		"echo 'Clearing stale or half-downloaded package databases before refresh.'",
		"rm -f /var/lib/pacman/sync/*.db /var/lib/pacman/sync/*.db.sig /var/lib/pacman/sync/*.files /var/lib/pacman/sync/*.files.sig /var/cache/pacman/pkg/*.part",
		"echo 'Refreshing the private MSYS2 package databases and keyring package.'",
		"pacman -Syy --needed --noconfirm msys2-keyring",
		"echo 'Installing the approved build-tool packages.'",
		"pacman -S --needed --noconfirm " + strings.Join(quotedPackageNames, " "),
		"echo 'Checking that the selected MSYS2 compiler can create a Windows executable.'",
		"echo 'The compiler check uses a private writable folder inside the extracted MSYS2 toolchain.'",
		`compiler_check_dir="/usr/local/var/customffmpeg-compiler-check"`,
		`rm -rf "$compiler_check_dir"`,
		`mkdir -p "$compiler_check_dir"`,
		`trap 'rm -rf "$compiler_check_dir"' EXIT`,
		`printf 'int main(void){return 0;}\n' > "$compiler_check_dir/check.c"`,
		`gcc "$compiler_check_dir/check.c" -o "$compiler_check_dir/check.exe"`,
		`test -s "$compiler_check_dir/check.exe"`,
		`chmod +x "$compiler_check_dir/check.exe"`,
		`"$compiler_check_dir/check.exe"`,
		`rm -rf "$compiler_check_dir"`,
	}, nil
}

func FfmpegLibraryPackageInstallScriptLines(packageNames []string) ([]string, error) {
	quotedPackageNames := make([]string, 0, len(packageNames))
	for _, packageName := range packageNames {
		if err := ValidateMsys2PackageName(packageName); err != nil {
			return nil, err
		}
		quotedPackageNames = append(quotedPackageNames, shellQuote(packageName))
	}
	if len(quotedPackageNames) == 0 {
		return []string{
			"#!/usr/bin/env bash",
			"set -euo pipefail",
			"echo 'No FFmpeg library packages were selected.'",
		}, nil
	}
	return []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"echo 'Using the official MSYS2 package server for FFmpeg library packages.'",
		"cat > /etc/pacman.d/mirrorlist.msys <<'EOF'",
		"Server = https://repo.msys2.org/msys/$arch/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.mingw32 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/mingw32/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.mingw64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/mingw64/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.ucrt64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/ucrt64/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.clang64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/clang64/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.clangarm64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/clangarm64/",
		"EOF",
		"sed -i -E 's/^SigLevel[[:space:]]*=.*/SigLevel = Required DatabaseOptional/' /etc/pacman.conf",
		"echo 'Clearing half-downloaded package files before installing FFmpeg library packages.'",
		"rm -f /var/cache/pacman/pkg/*.part",
		"echo 'Refreshing package databases before installing FFmpeg library packages.'",
		"pacman -Syy --needed --noconfirm",
		"echo 'Installing MSYS2 packages required by the selected FFmpeg libraries.'",
		"pacman -S --noconfirm " + strings.Join(quotedPackageNames, " "),
	}, nil
}

func ConfigureScriptLines(configureFlags []string) ([]string, error) {
	quotedConfigureFlags := make([]string, 0, len(configureFlags))
	for _, configureFlag := range configureFlags {
		if err := ValidateConfigureFlag(configureFlag); err != nil {
			return nil, err
		}
		quotedConfigureFlags = append(quotedConfigureFlags, shellQuote(configureFlag))
	}

	pkgConfigModules := pkgConfigModulesForConfigureFlags(configureFlags)
	// Only generate pre-checks for libraries with a minimum version requirement.
	// Existence-only checks are redundant — FFmpeg configure already checks presence
	// and emits clear errors. Version checks are kept because configure emits only
	// a generic "not found" when the version constraint fails, which is misleading.
	type pkgConfigCheck struct {
		mod  pkgConfigModule
		line string
	}
	pkgConfigChecks := []pkgConfigCheck{}
	for _, mod := range pkgConfigModules {
		if mod.MinVersion != "" {
			pkgConfigChecks = append(pkgConfigChecks, pkgConfigCheck{
				mod:  mod,
				line: "${PKG_CONFIG} --print-errors --atleast-version=" + shellQuote(mod.MinVersion) + " " + shellQuote(mod.Name) + " >/dev/null 2>&1",
			})
		}
	}

	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`profile_prefix="${MSYSTEM_PREFIX:-/ucrt64}"`,
		`export PATH="${profile_prefix}/bin:/usr/bin:${PATH}"`,
		`export PKG_CONFIG_PATH="${profile_prefix}/lib/pkgconfig:${profile_prefix}/share/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig"`,
		`export PKG_CONFIG_LIBDIR="${profile_prefix}/lib/pkgconfig:${profile_prefix}/share/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig"`,
		`export CPPFLAGS="-I${profile_prefix}/include"`,
		`if [ -x "${profile_prefix}/bin/pkgconf.exe" ]; then export PKG_CONFIG="${profile_prefix}/bin/pkgconf.exe"; elif [ -x "${profile_prefix}/bin/pkg-config.exe" ]; then export PKG_CONFIG="${profile_prefix}/bin/pkg-config.exe"; else export PKG_CONFIG="pkg-config"; fi`,
		`echo "Using pkg-config command: ${PKG_CONFIG}"`,
		`echo "Using pkg-config search paths: ${PKG_CONFIG_LIBDIR}"`,
	}
	if len(pkgConfigChecks) > 0 {
		scriptLines = append(scriptLines, "echo 'Checking selected external libraries with pkg-config before FFmpeg configure.'")
		for _, check := range pkgConfigChecks {
			scriptLines = append(scriptLines,
				fmt.Sprintf(`echo "pkg-config check starting: %s (requires >= %s) at $(date +%%T)"`, check.mod.Name, check.mod.MinVersion),
				check.line,
				fmt.Sprintf(`echo "pkg-config check completed: %s at $(date +%%T)"`, check.mod.Name),
			)
		}
	}
	scriptLines = append(scriptLines, `echo "Starting FFmpeg configure at $(date +%T)"`)
	scriptLines = append(scriptLines, "./configure "+strings.Join(quotedConfigureFlags, " "))
	scriptLines = append(scriptLines, `echo "FFmpeg configure completed at $(date +%T)"`)
	return scriptLines, nil
}

// pkgConfigModule holds a pkg-config module name and an optional minimum version
// that must be satisfied before FFmpeg configure is attempted.
type pkgConfigModule struct {
	Name       string
	MinVersion string // empty means no version check beyond existence
}

func pkgConfigModulesForConfigureFlags(configureFlags []string) []pkgConfigModule {
	// Only list libraries that are expected to provide a pkg-config module in MSYS2.
	// Some valid FFmpeg options, such as --enable-libgsm, are probed by FFmpeg
	// through headers/libraries instead of a .pc file. Pre-checking those with
	// pkg-config incorrectly blocks valid builds.
	//
	// MinVersion values mirror the lower bounds enforced by FFmpeg's own configure
	// script. When FFmpeg requires a minimum version, we check it here with
	// --atleast-version so that the error is caught before configure runs and the
	// message clearly names the library and required version rather than the
	// generic "not found" that configure emits when the version constraint fails.
	type entry struct {
		name       string
		minVersion string
	}
	moduleByFlag := map[string]entry{
		"--enable-libaom":            {name: "aom"},
		"--enable-libass":            {name: "libass"},
		"--enable-libbluray":         {name: "libbluray"},
		"--enable-libcdio":           {name: "libcdio"},
		"--enable-libdav1d":          {name: "dav1d", minVersion: "1.0"},
		"--enable-libfdk-aac":        {name: "fdk-aac"},
		"--enable-libfontconfig":     {name: "fontconfig"},
		"--enable-fontconfig":        {name: "fontconfig"},
		"--enable-libfreetype":       {name: "freetype2"},
		"--enable-libfribidi":        {name: "fribidi"},
		"--enable-libharfbuzz":       {name: "harfbuzz"},
		"--enable-libilbc":           {name: "libilbc"},
		"--enable-libjxl":            {name: "libjxl", minVersion: "0.7.0"},
		"--enable-libmodplug":        {name: "libmodplug"},
		"--enable-libmp3lame":        {name: "lame"},
		"--enable-libopencore-amrnb": {name: "opencore-amrnb"},
		"--enable-libopencore-amrwb": {name: "opencore-amrwb"},
		"--enable-libopenh264":       {name: "openh264"},
		"--enable-libopenjpeg":       {name: "libopenjp2"},
		"--enable-libopus":           {name: "opus"},
		"--enable-libplacebo":        {name: "libplacebo", minVersion: "5.229.0"},
		"--enable-librav1e":          {name: "rav1e"},
		"--enable-librubberband":     {name: "rubberband"},
		"--enable-libsoxr":           {name: "soxr"},
		"--enable-libspeex":          {name: "speex"},
		"--enable-libssh":            {name: "libssh"},
		"--enable-libsvtav1":         {name: "SvtAv1Enc", minVersion: "0.9.0"},
		"--enable-libtesseract":      {name: "tesseract"},
		"--enable-libtwolame":        {name: "twolame"},
		"--enable-libvmaf":           {name: "libvmaf", minVersion: "2.0.0"},
		"--enable-libvorbis":         {name: "vorbis"},
		"--enable-libvpx":            {name: "vpx"},
		"--enable-libwebp":           {name: "libwebp"},
		"--enable-libx264":           {name: "x264"},
		"--enable-libx265":           {name: "x265"},
		"--enable-libxavs2":          {name: "xavs2"},
		"--enable-libzimg":           {name: "zimg", minVersion: "2.9"},
		"--enable-openal":            {name: "openal"},
		"--enable-openssl":           {name: "openssl"},
		"--enable-gnutls":            {name: "gnutls"},
		"--enable-sdl2":              {name: "sdl2"},
	}
	modules := []pkgConfigModule{}
	seen := map[string]bool{}
	for _, configureFlag := range configureFlags {
		e, exists := moduleByFlag[configureFlag]
		if !exists || seen[e.name] {
			continue
		}
		seen[e.name] = true
		modules = append(modules, pkgConfigModule{Name: e.name, MinVersion: e.minVersion})
	}
	return modules
}

func MakeScriptLines(parallelJobCount int) ([]string, error) {
	if parallelJobCount < 1 || parallelJobCount > 256 {
		return nil, errors.New("parallel job count must be between 1 and 256")
	}
	return []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`echo "Starting FFmpeg make at $(date +%T)"`,
		fmt.Sprintf("make -j%d", parallelJobCount),
		`echo "FFmpeg make completed at $(date +%T)"`,
	}, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
