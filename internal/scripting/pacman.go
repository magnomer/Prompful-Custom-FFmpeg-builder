package scripting

import (
	"fmt"
	"strings"
)

// LRepoMingwNames are the MSYS2 mingw-family repositories that the default
// pacman.conf enables. A build targets exactly one of them (the selected shell
// profile); the rest are disabled before syncing so an unrelated repo's transient
// mirror failure cannot affect ??or appear in ??the build.
var LRepoMingwNames = []string{"mingw32", "mingw64", "ucrt64", "clang64", "clangarm64"}

// LMSYSMirrorCatalog are the upstream package servers pacman tries in order. The first two
// are MSYS2's own origin and CDN; the rest are long-lived independent mirrors. Listing
// several lets a per-file 404 fall through to the next server when one mirror has not
// yet propagated a just-published package (mirror/CDN skew) instead of failing the
// build. The list is fixed and ordered so the generated, hash-pinned script is stable.
var LMSYSMirrorCatalog = []string{
	"https://repo.msys2.org",
	"https://mirror.msys2.org",
	"https://mirrors.tuna.tsinghua.edu.cn/msys2",
	"https://download.nus.edu.sg/mirror/msys2",
}

// LRepoMirrorlistMsys pairs each pacman mirrorlist file with the repository path
// appended to every mirror base. Ordered for a deterministic script. The "mingw"
// entry writes mirrorlist.mingw, the file stock pacman.conf actually Includes for
// every mingw-family repo (ucrt64/mingw64/clang64/etc.); $repo is a pacman variable
// so one file serves them all and the ordered LMSYSMirrorCatalog finally takes effect.
// The per-profile files are kept (harmless, deterministic) for any logic/tests that
// reference their names.
var LRepoMirrorlistMsys = []struct {
	mirrorlistName string
	repoPath       string
}{
	{"msys", "/msys/$arch/"},
	{"mingw", "/mingw/$repo/"},
	{"mingw32", "/mingw/mingw32/"},
	{"mingw64", "/mingw/mingw64/"},
	{"ucrt64", "/mingw/ucrt64/"},
	{"clang64", "/mingw/clang64/"},
	{"clangarm64", "/mingw/clangarm64/"},
}

// LMirrorScriptCreate writes every pacman mirrorlist with the full ordered mirror set
// so pacman retries the next server on a per-file 404. Shared by the toolchain install
// and the FFmpeg library/build-dependency package install so both gain the same fallback.
func LMirrorScriptCreate() []string {
	lines := []string{}
	for _, repo := range LRepoMirrorlistMsys {
		lines = append(lines, "cat > /etc/pacman.d/mirrorlist."+repo.mirrorlistName+" <<'EOF'")
		for _, mirror := range LMSYSMirrorCatalog {
			lines = append(lines, "Server = "+mirror+repo.repoPath)
		}
		lines = append(lines, "EOF")
	}
	return lines
}

// LXferCommandSet writes a pacman XferCommand into pacman.conf so a stalled mirror
// fails over fast instead of hanging. curl aborts a transfer that stays under
// 2048 B/s for 30s (--speed-limit/--speed-time), whereupon pacman advances to the
// next Server in the mirrorlist; -C - resumes a partially cached file and --retry 3
// tolerates transient errors. curl already ships in MSYS2. Shared by the toolchain
// install and the FFmpeg library/build-dependency install so both gain the failover.
func LXferCommandSet() []string {
	return []string{
		"echo 'Setting a pacman transfer command that fails a stalled mirror over to the next server.'",
		`printf '%s\n' 'XferCommand = /usr/bin/curl -fL -C - --retry 3 --speed-limit 2048 --speed-time 30 -o %o %u' >> /etc/pacman.conf`,
	}
}

func LRepositoryNameResolve(windowsShellProfileName string) string {
	for _, repoName := range LRepoMingwNames {
		if repoName == windowsShellProfileName {
			return repoName
		}
	}
	return "ucrt64"
}

func LScriptPacmanBuild(packageNames []string, windowsShellProfileName string) ([]string, error) {
	quotedPackageNames := make([]string, 0, len(packageNames))
	for _, packageName := range packageNames {
		if err := LPackageMsysValidate(packageName); err != nil {
			return nil, err
		}
		quotedPackageNames = append(quotedPackageNames, LShellTextQuote(packageName))
	}

	// Disable every mingw-family repo except the selected profile's, so `pacman -Syy`
	// only refreshes the databases this build actually uses (the profile repo plus
	// the always-needed `msys` repo). This keeps a build that chose mingw64 from
	// touching clang64/ucrt64/etc. databases.
	selectedRepoName := LRepositoryNameResolve(windowsShellProfileName)
	disableUnusedRepoLines := []string{"echo 'Limiting package databases to the selected profile and the msys repository.'"}
	for _, repoName := range LRepoMingwNames {
		if repoName == selectedRepoName {
			continue
		}
		disableUnusedRepoLines = append(disableUnusedRepoLines,
			fmt.Sprintf("sed -i -E '/^\\[%s\\]/,/^Include/ s/^/#/' /etc/pacman.conf", repoName))
	}

	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"echo 'Using the official MSYS2 package servers (with fallback mirrors) for this private toolchain.'",
	}
	scriptLines = append(scriptLines, LMirrorScriptCreate()...)
	scriptLines = append(scriptLines,
		"echo 'Initializing the private MSYS2 package keyring.'",
		"pacman-key --init 2>&1",
		"pacman-key --populate msys2 2>&1",
		"echo 'Preparing pacman database signature policy.'",
		"# MSYS2 packages remain signature-checked. Repository database signatures are treated as optional, matching normal MSYS2 pacman behavior and avoiding false failures when a database .sig is not served or is cleared during refresh.",
		"sed -i -E 's/^SigLevel[[:space:]]*=.*/SigLevel = Required DatabaseOptional/' /etc/pacman.conf",
	)
	scriptLines = append(scriptLines, LXferCommandSet()...)
	scriptLines = append(scriptLines, disableUnusedRepoLines...)
	scriptLines = append(scriptLines,
		"echo 'Clearing stale or half-downloaded package databases before refresh.'",
		"rm -f /var/lib/pacman/sync/*.db /var/lib/pacman/sync/*.db.sig /var/lib/pacman/sync/*.files /var/lib/pacman/sync/*.files.sig /var/cache/pacman/pkg/*.part",
		"echo 'Refreshing the private MSYS2 package databases and keyring package.'",
		"pacman -Syy --needed --noconfirm msys2-keyring",
		"echo 'Installing the approved build-tool packages.'",
		"pacman -S --needed --noconfirm "+strings.Join(quotedPackageNames, " "),
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
	)
	return scriptLines, nil
}

func LPackageScriptCreate(packageNames []string) ([]string, error) {
	quotedPackageNames := make([]string, 0, len(packageNames))
	for _, packageName := range packageNames {
		if err := LPackageMsysValidate(packageName); err != nil {
			return nil, err
		}
		quotedPackageNames = append(quotedPackageNames, LShellTextQuote(packageName))
	}
	if len(quotedPackageNames) == 0 {
		return []string{
			"#!/usr/bin/env bash",
			"set -euo pipefail",
			"echo 'No FFmpeg library packages were selected.'",
		}, nil
	}
	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"echo 'Using the official MSYS2 package servers (with fallback mirrors) for FFmpeg library packages.'",
	}
	scriptLines = append(scriptLines, LMirrorScriptCreate()...)
	scriptLines = append(scriptLines,
		"sed -i -E 's/^SigLevel[[:space:]]*=.*/SigLevel = Required DatabaseOptional/' /etc/pacman.conf",
	)
	scriptLines = append(scriptLines, LXferCommandSet()...)
	scriptLines = append(scriptLines,
		"echo 'Clearing half-downloaded package files before installing FFmpeg library packages.'",
		"rm -f /var/cache/pacman/pkg/*.part",
		// Full refresh + upgrade (-Syyu), not a bare -Syy, so the install is not a partial
		// upgrade: MSYS2 is rolling, and installing a new package against a stale set of
		// already-installed dependencies can pull a dependency version whose file the mirror
		// has already superseded (404). A full upgrade keeps every package coherent.
		//
		// msys2-runtime is held back with --ignore: upgrading the core runtime forces MSYS2 to
		// close every process including the build shell ("To complete this update all MSYS2
		// processes including this terminal will be closed"), which aborts the non-interactive
		// build with exit status 1. The FFmpeg libraries install into the mingw/ucrt repos,
		// which do not depend on a newer msys2-runtime, so holding it back keeps the install
		// coherent without killing the shell.
		"echo 'Refreshing databases and upgrading the environment before installing FFmpeg library packages.'",
		"pacman -Syyu --needed --noconfirm --ignore msys2-runtime --ignore msys2-runtime-devel",
		"echo 'Installing MSYS2 packages required by the selected FFmpeg libraries.'",
		// --overwrite '*' lets pacman take ownership of untracked files in the private,
		// rebuildable prefix. A prior library prep (e.g. AviSynth+) may have written headers
		// such as avisynth/avisynth_c.h that an MSYS2 package (libx264) also ships; without
		// --overwrite, pacman aborts with "conflicting files" on a dirty/re-run prefix.
		// Install ordering re-runs the prep afterward, so the prep-provided files win in the end.
		"pacman -S --noconfirm --overwrite '*' "+strings.Join(quotedPackageNames, " "),
	)
	scriptLines = append(scriptLines, LRabbitMQScriptClean()...)
	scriptLines = append(scriptLines, LOAPVPackageRepair()...)
	return scriptLines, nil
}

// LRabbitMQScriptClean repairs the MSYS2 librabbitmq.pc. The rabbitmq-c
// CMake build composes Libs.private by looping its socket libraries as "-l<lib>"; an
// empty list element leaves a stray, name-less "-l" token (the shipped 0.15 .pc reads
// "Libs.private: ... -lws2_32 -l -lssl -lcrypto"). With the build's default
// --pkg-config-flags=--static, FFmpeg's configure link probe passes that bare "-l" to
// gcc, the link fails, and configure reports "librabbitmq >= 0.7.1 not found".
// Stripping the empty "-l" namespec from the Libs lines makes the static link valid.
// Guarded by file existence, so it is a no-op when rabbitmq-c was not installed.
func LRabbitMQScriptClean() []string {
	return []string{
		`rabbitmq_pc="${MSYSTEM_PREFIX:-/ucrt64}/lib/pkgconfig/librabbitmq.pc"`,
		`if [ -f "${rabbitmq_pc}" ]; then`,
		`  echo "Repairing librabbitmq.pc: removing empty -l tokens that break the static link probe."`,
		`  sed -i -E '/^Libs/ s/ -l( |$)/\1/g' "${rabbitmq_pc}"`,
		`  echo "Patched librabbitmq.pc: $(grep -E '^Libs' "${rabbitmq_pc}")"`,
		`fi`,
	}
}

// LOAPVPackageRepair repairs the MSYS2 openapv pkg-config file.
// The current MSYS2 package installs the archives under lib/oapv/ (and the import
// library under lib/oapv/import/) while the pkg-config file can advertise the
// ordinary profile lib directory. FFmpeg's configure then sees the oapv module
// but the link probe cannot resolve -loapv, finally reporting
// "oapv >= 0.2.0.0 not found using pkg-config". Pointing Libs at the shipped
// archive directory makes the configure probe find the static archive. The same
// package leaves OAPV_STATIC_DEFINE only in Cflags.private, but FFmpeg's normal
// probe does not request private cflags; without that define the header declares
// dllimport symbols and the static link fails with __imp_oapve_encode unresolved.
// Guarded by file existence, so it is a no-op when openapv was not selected/installed.
func LOAPVPackageRepair() []string {
	return []string{
		`oapv_pc="${MSYSTEM_PREFIX:-/ucrt64}/lib/pkgconfig/oapv.pc"`,
		`if [ -f "${oapv_pc}" ]; then`,
		`  echo "Repairing oapv.pc: pointing Libs at MSYS2's openapv archive directory and exposing the static compile define."`,
		`  sed -i -E 's|^Libs:.*|Libs: -L${libdir}/oapv -loapv|' "${oapv_pc}"`,
		`  if grep -q '^Cflags:' "${oapv_pc}"; then`,
		`    sed -i -E '/^Cflags:/ { /-DOAPV_STATIC_DEFINE/! s|$| -DOAPV_STATIC_DEFINE|; }' "${oapv_pc}"`,
		`  else`,
		`    printf '%s\n' 'Cflags: -I${includedir} -DOAPV_STATIC_DEFINE' >> "${oapv_pc}"`,
		`  fi`,
		`  echo "Patched oapv.pc: $(grep -E '^(Libs|Cflags)' "${oapv_pc}")"`,
		`fi`,
	}
}
