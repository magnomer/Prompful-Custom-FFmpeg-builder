# PlanCLI — PromptfulX CLI Implementation Plan

Status: proposal / roadmap.
Scope: turn the single Wails GUI binary into two standalone executables
(`promptful.exe` GUI, `promptfulx.exe` CLI) sharing one backend, per
`CLIImplementation.txt`.

This document is the phased execution plan. It is grounded in the current code,
not the aspirational spec. The spec is treated as a v2/v3 wishlist; this plan
ships a working single-build CLI first, then layers batch on top.

---

## 1. Current state (what already exists)

The project is ~80% ready. Key facts, verified against the tree:

- **Backend build logic is already GUI-free.** Build workers live in
  wails-free packages: `internal/planning`, `internal/execution`,
  `internal/download`, `internal/extraction`, `internal/workspace`,
  `internal/catalogfacts`, `internal/scripting`, `internal/consent`,
  `internal/audit`, `internal/reviewsession`. None import wails.

- **Static data is already embedded.** `internal/catalogfacts/catalogfacts.go:13`
  uses `//go:embed catalogdata/libraries/*.json catalogdata/versions/*.json
  catalogdata/presets/*.json catalogdata/librarysources/*.json`. The spec's
  entire §4 "EmbeddedDataStore" requirement is effectively done. The root
  `versions/ libraries/ presets/` dirs are source, not runtime dependencies.

- **Progress is already an abstraction.** Every deep worker takes
  `emitProgress func(string, string)` as a parameter (see all funcs in
  `internal/program/ffmpeg_run.go`, `artifact.go`, `library_prep.go`,
  `toolchain_run.go`). Progress is NOT hardcoded to wails inside the workers.

- **The build-request model already exists.** `planning.LSettingsFFmpeg`
  (`internal/planning/types.go:99`) is the pure input struct. The planner
  `planning.LPlanFFmpegCreate(settings)` is pure and wails-free. This is the
  CLI's `BuildRequest` target — no new model needed.

- **Wails coupling is shallow and isolated.** All 12 wails call sites are in
  `internal/program` (only that package imports wails). Of those, the build
  path touches wails through exactly **3 seams** (below). The rest are window
  geometry and file/dir dialogs, which the CLI does not use.

### 1.1 The three wails seams on the build path

| Seam | Current implementation | Site | CLI replacement |
|---|---|---|---|
| status | `wailsRuntime.EventsEmit("approved-action-status")` | `program.go:520` `lStatusEmit` | print to stdout |
| log | `wailsRuntime.EventsEmit("security-log")` | `program.go:526` `lLogEmit` | print stdout/stderr |
| confirm | `wailsRuntime.MessageDialog` yes/no | `program.go:456` `lConsentNativeAsk` | `--yes` / stdin prompt |

The reusable worker `lFFmpegBuild` (`internal/program/ffmpeg_run.go:26`) touches
wails ONLY through these three funcs. Abstract them and the whole worker becomes
CLI-usable unchanged.

### 1.2 Build entry points (reuse targets)

- `LProgram.LPlanFFmpegRequest(settings) -> LReviewFFmpeg` — pure; creates plan +
  review session. Maps to CLI `plan`.
- `LProgram.LPlanFFmpegApprove(sessionId, approval) -> LResultAction` — validates,
  native-confirm (seam 3), then `go lFFmpegBuild(...)` async. Maps to CLI `build`.
- `lFFmpegBuild(...)` — the actual build worker. Reused as-is by the CLI once the
  seams are abstracted.

---

## 2. Naming-rule compliance

All new names obey `docs/internal/NamingRules.md`:
`Object + optional single Modifier + bare verb`. New L-objects MUST be
registered in `docs/internal/objects.json`, and any user-action chains in
`docs/internal/chains.json`.

New objects to register:

| Object | Role |
|---|---|
| `LReporter` | Build status + log output sink (GUI events / CLI stdout). |
| `LConfirmer` | Single backend-owned approval gate (native dialog / CLI prompt). |
| `LBuildFlow` | (Phase 1B) wails-free owner of the build worker. |
| `LArgument` | (v2) expands raw CLI argument sources. |
| `LFlagFile` | (v2) parses one `.flags` file. |
| `LBatch`, `LBatchPlan`, `LBatchReport` | (v2) batch coordination. |

New methods (all end in a bare verb already used in the codebase — `Emit`,
`Get`, `Run`, `Resolve`, `Parse`, `Create`):

- `LReporterStatusEmit(status string)`
- `LReporterLogEmit(level, message string)`
- `LConfirmerApprovalGet(actionName, planHash, message string) (bool, error)`
  (verb `Get` returns the yes/no decision; confirm `Ask`/`Get` choice against
  the verb whitelist before coding.)

---

## 3. Phased plan

### Ladder overview

```
Step 0   cmd/ split                            mechanical, ship first     DONE
Step 1A  Reporter + Confirmer interface shim   the only seam work         DONE
Step 1   CLI reporter/confirmer impls                                     DONE
Step 2   version + preset + flag resolvers -> backend (shared, tested)   DONE
Step 3   plan + list commands                  pure, cheap, high value    DONE
Step 4   build command (sync, exit codes)                                 DONE
-------- v1 SHIPS: standalone single-build CLI --------
Step 5   flag files @file  (spec §15)
Step 6   verify + explain commands
Step 7   Phase 1B: worker -> internal/buildflow (removes wails from CLI binary)
-------- v2 --------
Step 8   batch: folder/zip discovery, isolation, summary (spec §17-34)
Step 9   parallel, json-log, strict/dry-run
```

Steps 2 and 3 were swapped from an earlier draft: the `plan`/`list` commands
call the version/preset/flag resolvers, so the resolvers (now Step 2) must land
before the commands (now Step 3). Build order = number order.

Biggest early win: Steps 2-3 give a working `promptfulx plan` / `list` quickly —
pure functions, no seam work. The seam refactor (1A, done) only gates `build`.

---

### Step 0 — add the CLI binary (DONE)

Mechanical. No logic change.

```
main.go                 GUI, stays at repo root (wails build works)
cmd/
  promptfulx/main.go    new CLI entry
```

> Decision: the GUI main stays at the repo root, NOT under `cmd/promptful`.
> The wails CLI (`wails dev` / `wails build`, per README) requires the main
> package at the module root and has no config field to relocate it. Go's
> `//go:embed` also cannot reach `frontend/dist` from `cmd/promptful` (no `..`
> in embed paths). Moving the GUI would break the wails toolchain and its NSIS
> installer packaging for no gain — the CLI does not need wails at all. If the
> project later drops the wails CLI, the GUI can move to `cmd/promptful` then.

- Root `main.go` unchanged; `wails.json` and the frontend embed untouched.
- New `cmd/promptfulx/main.go`: CLI skeleton (arg dispatch + usage). Exit code 2
  on unknown command, per §38.
- Build: GUI via `wails build`; CLI via `go build ./cmd/promptfulx`.
- Verified: both compile; `promptfulx` prints usage (exit 0) and returns exit 2
  on an unknown command.

### Step 1A — Reporter + Confirmer interface shim

Keeps everything in `internal/program`. Proves the CLI can drive the worker.

**1A.1** New package `internal/reporting/reporting.go` (zero deps, no wails):

```go
package reporting

// LReporter receives build status + log output. The GUI impl emits wails
// events; the CLI impl prints to stdout/stderr.
type LReporter interface {
	LReporterStatusEmit(status string)
	LReporterLogEmit(level string, message string)
}

// LConfirmer answers the single backend-owned approval gate. The GUI impl
// shows a native dialog; the CLI impl honors --yes or prompts stdin.
type LConfirmer interface {
	LConfirmerApprovalGet(actionName string, planHash string, message string) (bool, error)
}
```

**1A.2** Add fields to `LProgram` (`internal/program/program.go:25`):

```go
LReporter  reporting.LReporter
LConfirmer reporting.LConfirmer
```

**1A.3** Gut the three seam funcs to delegate:

```go
func (program *LProgram) lStatusEmit(status string) {
	if program.LReporter != nil {
		program.LReporter.LReporterStatusEmit(status)
	}
}

func (program *LProgram) lLogEmit(level, message string) {
	if program.LReporter != nil {
		program.LReporter.LReporterLogEmit(level, message)
	}
}
```

In `lConsentNativeAsk` (`program.go:456`): keep building the localized message
exactly as now, then replace the `wailsRuntime.MessageDialog` call with
`program.LConfirmer.LConfirmerApprovalGet(actionName, planHash, message)`.

**1A.4** GUI implementations — keep in `internal/program` (it already imports
wails). LContext changes at startup, so read it lazily via the program pointer:

```go
type LReporterWails struct{ program *LProgram }

func (r LReporterWails) LReporterStatusEmit(status string) {
	if r.program.LContext != nil {
		wailsRuntime.EventsEmit(r.program.LContext, "approved-action-status",
			map[string]string{"status": status})
	}
}
// LReporterLogEmit mirrors the old lLogEmit body.
// LConfirmerWails wraps the old MessageDialog body.
```

Wire in `LProgramStart`:
`program.LReporter = LReporterWails{program}` (and the confirmer).

After 1A the GUI behaves identically, and `lFFmpegBuild` runs with any reporter.

### Step 1 — CLI reporter / confirmer

`cmd/promptfulx/reporter.go`:

```go
type LReporterConsole struct{ quiet bool }

func (r LReporterConsole) LReporterStatusEmit(s string) {
	fmt.Fprintf(os.Stdout, "[status] %s\n", s)
}

func (r LReporterConsole) LReporterLogEmit(level, msg string) {
	w := os.Stdout
	if level == "error" || level == "warn" {
		w = os.Stderr
	}
	fmt.Fprintf(w, "[%s] %s\n", level, msg)
}

type LConfirmerFlag struct{ assumeYes, noInput bool }

func (c LConfirmerFlag) LConfirmerApprovalGet(action, hash, msg string) (bool, error) {
	if c.assumeYes {
		return true, nil
	}
	if c.noInput {
		return false, errors.New("approval required but --no-input set")
	}
	fmt.Printf("%s\nProceed? [y/N] ", msg)
	var in string
	fmt.Scanln(&in)
	return in == "y" || in == "Y", nil
}
```

### Step 2 — version + preset + flag resolvers (shared backend)

`LSettingsFFmpeg` takes a source archive URL + a `SelectedLibraryIds` list, not a
version string or preset name. The planner (`LPlanFFmpegCreate`) is already
authoritative — given the URL and library IDs it resolves version, libraries,
compatibility, and the configure command itself. So the CLI only has to produce
those two inputs, and every source primitive already exists in the backend. No
frontend logic needs extraction (the earlier "extract PresetResolutionService"
concern was wrong: presets carry their library IDs directly).

| Need | Existing backend primitive |
|---|---|
| version -> archive/signature URL | `planning.LReleaseSupportedListGet()` -> `LReleaseChoice{Version, Codename, ArchiveUrl, SignatureUrl}` |
| preset -> library IDs (per version, `normal`/`extended` mode) | `LCatalogPresetSourceBuildResolved(url, profile)` -> `LPresetLibraryChoice{PresetId, LibraryIds, ExtendedLibraryIds}` |
| ffmpeg flag -> internal library ID | `LCatalogSourceGet(url, profile)` -> `LLibraryChoice{LibraryId, ConfigureFlags}` (scan ConfigureFlags) |
| shell profile default | `LCatalogDefaultWindowsShellProfileName = "ucrt64"` (empty normalizes to it) |

Facts that shape the CLI surface (verified against `catalogdata`):

- Real preset IDs: `ai, compatibility, default, editor, efficiency, full,
  maxtest, minimal, streaming`. (The spec's `extended-full` is fictional.)
- Presets are per-FFmpeg-version and have two modes: `normal` (`libraryIds`) and
  `extended` (`extendedLibraryIds`, a superset). CLI: `--preset P [--extended]`.
- Supported versions: 4.4.8, 5.1.9, 6.1.6, 7.0.3, 7.1.5, 8.0.3, 8.1.2.
- Library IDs differ from ffmpeg flag names (`x264` id vs `--enable-libx264`
  flag; `fdk-aac` vs `--enable-libfdk-aac`), so the flag surface needs the
  ConfigureFlags scan.

New resolvers in `internal/planning` (shared by GUI and CLI, unit-tested):

```text
LReleaseVersionResolve(version string) (LReleaseChoice, bool)
LPresetLibraryIdsResolve(url, profile, presetId string, extended bool) ([]string, bool)
LLibraryFlagResolve(url, profile, flagName string) (libraryId string, bool)
```

Locked decisions:

- Library enable/disable surface is FFmpeg-style: `--enable-libx264` /
  `--disable-libfdk-aac`, mapped to internal IDs via `LLibraryFlagResolve`
  (spec §9.3).
- `list libraries` requires `--ffmpeg-version` (fails exit 2 otherwise) because
  availability is genuinely per-version.
- `--extended` selects a preset's extended mode. Default is normal.

### Step 3 — `plan` and `list` commands (pure, no seams)

Consumes the Step 2 resolvers. `cmd/promptfulx/main.go`:

```go
func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if len(args) == 0 {
		return runGuided() // spec §5.1 guided mode — stub for v1
	}
	switch args[0] {
	case "plan":
		return cmdPlan(args[1:])
	case "build":
		return cmdBuild(args[1:])
	case "list":
		return cmdList(args[1:])
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", args[0])
		return 2 // invalid arguments (spec §38)
	}
}
```

`parseSettings` maps flags into `LSettingsFFmpeg` using the Step 2 resolvers
(version -> URL, preset -> IDs, enable/disable overrides). `cmdPlan` then calls
the pure planner:

```go
func cmdPlan(args []string) int {
	settings, err := parseSettings(args) // uses Step 2 resolvers
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	plan, err := planning.LPlanFFmpegCreate(settings)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 4 // unsupported version/library/preset combo
	}
	printPlan(plan) // resolved libs, configure command, warnings
	return 0
}
```

`list versions|presets|libraries` reads embedded catalog data via the Step 2
resolvers / `planning` — no seams. `list libraries` requires `--ffmpeg-version`.
Output is human text for v1; JSON is deferred.

### Step 4 — `build` command

Unlike the GUI's fire-and-forget `go lFFmpegBuild`, the CLI runs synchronously
and maps the final status to an exit code:

```
resolve settings -> LPlanFFmpegRequest -> print plan
  -> LConfirmerApprovalGet (respects --yes / --no-input)
  -> LPlanFFmpegApprove (runs worker, blocks to completion in CLI)
  -> map status -> exit code
```

Exit codes to wire (spec §38): 0 success, 1 general, 2 invalid args,
4 unsupported combo, 7 configure failure, 8 build failure, 9 verify failure,
10 user cancelled. Start with 0/1/2/4/7/8 for v1.

Resolved: `LPlanFFmpegApprove`'s tail was refactored into shared helpers
(`lFFmpegApproveValidate` + `lFFmpegBuildLaunch(plan, approval, runInline)`).
GUI keeps the async `go` path via `LPlanFFmpegApprove`; the CLI calls the new
`LPlanFFmpegApproveSync`, which runs the worker inline and returns on completion.
The CLI's `buildReporter` captures the final status ("completed" -> 0, else 8).

Implemented exit codes: 0 completed, 8 build did not complete, 4 blocked plan /
unsupported combo, 2 invalid args, 10 confirmation rejected, 1 other validation
failure (e.g. toolchain not prepared). Finer configure/verify codes (7/9) are
deferred.

Resolved v1 gaps:

- `setup` command implemented (`cmd/promptfulx/setup.go`). It mirrors `build`:
  `LPlanToolchainRequest` -> `LPlanToolchainApproveSync` (the toolchain approve
  tail was refactored the same way as FFmpeg: `lToolchainApproveValidate` +
  `lToolchainPrepareLaunch(plan, approval, runInline)`). Starts from
  `LSettingsBuildCreate()` defaults with `--msys2-*` overrides. Exit codes:
  0 done, 6 setup failure, 4 blocked, 2 bad args, 10 rejected.
- `build` now pre-checks `LToolchainBuildPreparedCheck` and prints a CLI-native
  message with the exact `promptfulx setup --workspace ... --yes` command
  (exit 6), instead of the GUI-oriented "Go to Build configuration" text.

Remaining v1 caveats:

- `cmd/promptfulx` imports `internal/program`, so `promptfulx.exe` links the
  Wails library as dead weight until Step 7 (`buildflow`) removes it.
- A real end-to-end `setup` + `build` has not been run here (needs MSYS2
  download + a full compile). Pre-build/pre-setup gating is verified; the actual
  download/compile path is exercised by the shared GUI code but untested from
  the CLI entry.
- Naming: `LPlanFFmpegApproveSync` / `LPlanToolchainApproveSync` end in a
  qualifier rather than a bare verb. This matches existing deviations in the
  codebase (e.g. `LWorkspaceLayoutResolveVersioned`) but is worth a naming
  review pass.

**v1 ships here:** standalone single-build CLI.

### Step 5 — flag files `@file` (spec §15)

New `LFlagFile` parser: one arg per line, quoted values, `#` comments (literal
inside quotes), blank lines, UTF-8, Windows paths. Expand `@file` before command
parsing. Later args win (spec §10 override order). Defer nested files + cycle
detection until needed.

### Step 6 — `verify` and `explain` commands

`verify` reuses `LVerificationBuildRun` (`internal/program/verify_build.go:53`).
`explain` reads embedded catalog data. Both low-risk.

### Step 7 — Phase 1B: relocate worker to `internal/buildflow`

Goal: zero wails imported into `promptfulx.exe`.

- New wails-free package `internal/buildflow`. Move in: `lFFmpegBuild` and its
  helpers from `ffmpeg_run.go`, `library_prep.go`, `artifact.go`,
  `verify_build.go`, `toolchain_run.go`. They already take
  `emitProgress`/consents as params — a mechanical move.
- Convert receivers `(program *LProgram)` -> `(flow *LBuildFlow)`, where
  `LBuildFlow` holds `LReporter` + `LConfirmer` + the review-session store.
- `internal/program` shrinks to a GUI adapter: wails window/dialog/events +
  thin delegation into `buildflow`.
- Both `cmd/promptful` and `cmd/promptfulx` import `buildflow`; only
  `cmd/promptful` (+`internal/program`) import wails.

1B is bigger (file moves + receiver swaps) but no logic change. Do it after v1
proves out, not before.

> Until Step 7, `promptfulx.exe` imports `internal/program` and therefore
> compiles the wails library in as dead weight (a few MB, no runtime dependency,
> LContext stays nil so no event is ever emitted). Standalone still holds. Step 7
> only removes the bloat.

---

## 4. v2 — batch (spec §17-34)

Batch is the CLI's real advantage over the GUI, but it is a layer on top of a
solid single build: loop the jobs + isolate workspaces + summarize. Sequence:

- Step 8: `@folder` / `@zip` discovery (`.flags` only, recursive by default,
  deterministic sort), workspace + output isolation with pre-flight conflict
  detection, batch plan, continue-on-error default, batch summary, non-zero exit
  if any job failed (exit 11).
- Step 9: `--parallel N` (max 8), `--json-log` events, `--strict`, `--dry-run`,
  `--stop-on-error`, shared `--cache`.

Zip input must reject unsafe paths (zip-slip: `..\`, absolute paths) and select
only `.flags` entries (spec §30).

---

## 5. Scope discipline

The spec (`CLIImplementation.txt`) is a full v3 design written as a v1 spec.
Ship the minimum that is genuinely useful, in this order of value:

1. `plan` + `list` (Steps 0-2) — days of work, pure functions, no seam risk.
2. `build` (Steps 1A, 3, 4) — the seam shim + resolvers gate this.
3. Everything else is additive and can wait.

Do not build batch, parallelism, JSON logs, or strict/dry-run before a single
`promptfulx build` works end-to-end.
