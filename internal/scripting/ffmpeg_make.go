package scripting

import (
	"errors"
	"fmt"
)

func LMakeLinesCreate(parallelJobCount int) ([]string, error) {
	if parallelJobCount < 1 || parallelJobCount > 256 {
		return nil, errors.New("parallel job count must be between 1 and 256")
	}
	// Windows caps a native process command line at 32767 characters. A full shared FFmpeg
	// build links each shared library by passing every object file (~1240 for libavcodec,
	// about 30KB on their own) plus every enabled external library's -l/-L flags to gcc on a
	// single line, which overflows the cap and fails with "gcc: Argument list too long".
	// FFmpeg's config.mak sets CC=gcc and LD=gcc resolved through PATH, so a gcc shim placed
	// first on PATH intercepts both compilation and linking; when an invocation approaches the
	// cap it is forwarded through a gcc @response-file, which sidesteps the length limit (gcc
	// then also uses response files for its own collect2/ld sub-invocation).
	//
	// The shim is a small native program compiled once at the start of make. The common
	// short-command path is a single fast native exec of the real gcc with no extra process;
	// only the rare over-cap link invocation writes and forwards a response file. This avoids
	// re-spawning bash for every one of the ~2000+ compiler invocations (each MSYS2 process
	// spawn is a costly fork emulation), which added minutes of pure overhead.
	//
	// The shim is written against headers the ucrt64 mingw toolchain actually ships:
	// <process.h> (_spawnv/_P_WAIT) and _tempnam, NOT POSIX <spawn.h>/posix_spawn/mkstemp
	// (which ucrt64 does not provide -- a shim using them fails with "spawn.h: No such file
	// or directory"). The real gcc path is passed in at compile time via -DREAL_GCC using a
	// cygpath -m (mixed, forward-slash) Windows path, because the native shim resolves it with
	// no MSYS2 path conversion.
	//
	// If the shim fails to compile for any reason, make falls back to an equivalent bash gcc
	// wrapper (slower, but needs no compilation) so the build still proceeds.
	shimSource := []string{
		`#include <stdio.h>`,
		`#include <stdlib.h>`,
		`#include <string.h>`,
		`#include <stdint.h>`,
		`#include <process.h>`,
		`#ifndef REAL_GCC`,
		`#define REAL_GCC "gcc.exe"`,
		`#endif`,
		`int main(int argc, char **argv) {`,
		`    const char *gcc = REAL_GCC;`,
		`    size_t command_length = 0;`,
		`    for (int i = 1; i < argc; i++) command_length += strlen(argv[i]) + 1;`,
		`    if (command_length < 30000) {`,
		`        argv[0] = (char *)gcc;`,
		`        intptr_t rc = _spawnv(_P_WAIT, gcc, (const char *const *)argv);`,
		`        if (rc < 0) { perror("gcc shim: spawnv"); return 127; }`,
		`        return (int)rc;`,
		`    }`,
		`    char *response_path = _tempnam(NULL, "ffgcc");`,
		`    if (response_path == NULL) { perror("gcc shim: tempnam"); return 1; }`,
		`    FILE *rf = fopen(response_path, "w");`,
		`    if (rf == NULL) { perror("gcc shim: fopen"); free(response_path); return 1; }`,
		`    for (int i = 1; i < argc; i++) {`,
		`        for (const char *c = argv[i]; *c; c++) {`,
		`            if (*c == '\\' || *c == ' ' || *c == '"') fputc('\\', rf);`,
		`            fputc(*c, rf);`,
		`        }`,
		`        fputc('\n', rf);`,
		`    }`,
		`    if (fclose(rf) != 0) { perror("gcc shim: fclose"); remove(response_path); free(response_path); return 1; }`,
		`    char at_arg[4096];`,
		`    snprintf(at_arg, sizeof(at_arg), "@%s", response_path);`,
		`    const char *child_argv[] = { gcc, at_arg, NULL };`,
		`    intptr_t rc = _spawnv(_P_WAIT, gcc, child_argv);`,
		`    remove(response_path);`,
		`    free(response_path);`,
		`    if (rc < 0) { perror("gcc shim: spawnv"); return 1; }`,
		`    return (int)rc;`,
		`}`,
	}
	bashFallbackWrapper := []string{
		"#!/usr/bin/env bash",
		"set -uo pipefail",
		`real_gcc="${MSYSTEM_PREFIX:-/ucrt64}/bin/gcc.exe"`,
		`command_length=0`,
		`for argument in "$@"; do command_length=$(( command_length + ${#argument} + 1 )); done`,
		`if [ "${command_length}" -lt 30000 ]; then`,
		`  exec "${real_gcc}" "$@"`,
		`fi`,
		`response_file="$(mktemp)"`,
		`{`,
		`  for argument in "$@"; do`,
		`    escaped_argument=${argument//\\/\\\\}`,
		`    escaped_argument=${escaped_argument// /\\ }`,
		`    escaped_argument=${escaped_argument//\"/\\\"}`,
		`    printf '%s\n' "${escaped_argument}"`,
		`  done`,
		`} > "${response_file}"`,
		`"${real_gcc}" "@${response_file}"`,
		`gcc_status=$?`,
		`rm -f "${response_file}"`,
		`exit "${gcc_status}"`,
	}
	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`wrapper_dir="$(mktemp -d)"`,
		`trap 'rm -rf "${wrapper_dir}"' EXIT`,
		`real_gcc="${MSYSTEM_PREFIX:-/ucrt64}/bin/gcc.exe"`,
		`real_gcc_mixed="$(cygpath -m "${real_gcc}" 2>/dev/null || echo "${real_gcc}")"`,
		`cat > "${wrapper_dir}/gcc_shim.c" <<'FFMPEG_GCC_SHIM'`,
	}
	scriptLines = append(scriptLines, shimSource...)
	scriptLines = append(scriptLines, "FFMPEG_GCC_SHIM")
	scriptLines = append(scriptLines,
		`if "${real_gcc}" -O2 -DREAL_GCC="\"${real_gcc_mixed}\"" -o "${wrapper_dir}/gcc.exe" "${wrapper_dir}/gcc_shim.c" 2>"${wrapper_dir}/gcc_shim_build.log"; then`,
		`  echo "Compiled native gcc response-file shim at ${wrapper_dir} to bypass the Windows 32767-char command-line limit during the shared-library link (no per-call bash spawn)."`,
		`else`,
		`  echo "Native gcc shim did not compile; falling back to the bash gcc wrapper (slower but portable). Compiler output:"`,
		`  cat "${wrapper_dir}/gcc_shim_build.log" || true`,
		`  rm -f "${wrapper_dir}/gcc.exe"`,
		`  cat > "${wrapper_dir}/gcc" <<'FFMPEG_GCC_WRAPPER'`,
	)
	scriptLines = append(scriptLines, bashFallbackWrapper...)
	scriptLines = append(scriptLines,
		"FFMPEG_GCC_WRAPPER",
		`  chmod +x "${wrapper_dir}/gcc"`,
		`fi`,
		`export PATH="${wrapper_dir}:${PATH}"`,
		`echo "Starting FFmpeg make at $(date +%T)"`,
		fmt.Sprintf("make -j%d", parallelJobCount),
		`echo "FFmpeg make completed at $(date +%T)"`,
	)
	return scriptLines, nil
}
