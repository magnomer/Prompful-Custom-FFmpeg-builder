<div>

# Promptful Custom FFmpeg Builder

**EN** · A Windows desktop program (Go + Wails) that builds FFmpeg from source for you, using a private MSYS2 toolchain — after you review and approve each step.

<br>

**KO** · 비공개 MSYS2 툴체인으로 FFmpeg 소스를 로컬에서 직접 빌드하는 Windows 데스크톱 프로그램(Go + Wails)입니다. 빌드의 모든 단계는 사용자의 검토/승인을 받은 후에 이루어집니다.

<br>

> Built with the help of Claude and ChatGPT (Vibe coding). This is a personal project, so bugs may still exist.

<br><br>

<a href="#english"><img alt="English" src="https://img.shields.io/badge/English-Read-555555?style=for-the-badge"></a>
&nbsp;&nbsp;
<a href="#korean"><img alt="한국어" src="https://img.shields.io/badge/한국어-읽기-555555?style=for-the-badge"></a>

</div>

---

<a id="english"></a>

## English

### Screenshot

<img width="1349" height="953" alt="Screenshot-EN-1" src="https://github.com/user-attachments/assets/18a3d24a-d032-4737-9716-2fab3223d563" />
<img width="1349" height="953" alt="Screenshot-EN-2" src="https://github.com/user-attachments/assets/ff49f6f5-9808-4bb3-83cc-35a3c9ac0df3" />
<img width="1349" height="953" alt="Screenshot-EN-3" src="https://github.com/user-attachments/assets/9dcd6223-9c08-44b3-adb3-5edf6e50b8f6" />
<img width="1349" height="953" alt="Screenshot-EN-4" src="https://github.com/user-attachments/assets/30de29b2-4cb7-4b32-9cd3-0f716ff82309" />

### What it is

Promptful Custom FFmpeg Builder is a Windows GUI that automates a local FFmpeg build. You pick the libraries and options you want, review the exact plan it generates, and approve it. The program then downloads the sources, sets up a private MSYS2 build environment inside your chosen workspace, runs `configure` and `make`, and copies the finished binaries into a result folder.

The program ships with **no FFmpeg binary, no source archive, and no bundled codec libraries**. Everything is downloaded fresh from official sources during a build you approve.

### How it works

Every build runs through a review-then-approve flow:

1. You select libraries and configure options in the UI.
2. The program generates a plan: the exact packages, configure flags, scripts, license effects, and warnings.
3. You review that plan, then approve it.
4. A native Windows confirmation dialog asks one more time before anything runs.
5. The build runs and streams its log live. You can cancel at any point.

Downloads are verified before use — FFmpeg sources against their PGP signature, the MSYS2 archive against its detached signature. All work stays inside the workspace folder you pick; nothing is installed system-wide and admin rights are not required.

### FFmpeg build steps

1. Download the FFmpeg source archive, its `.asc` signature, and the release signing key.
2. Verify the signature, then extract the source.
3. Install the MSYS2 packages required by your selected libraries.
4. Generate, hash, and run the `configure` script through the private MSYS2 Bash.
5. Generate, hash, and run `make` with the chosen parallel job count.
6. Copy `ffmpeg.exe`, `ffprobe.exe`, and the required DLLs into `workspace/FFmpeg/`.
7. Write `build-report-<runId>.json` with the selected libraries, flags, license profile, file sizes, and SHA-256 hashes.

### Libraries and options

The **Library** page splits FFmpeg's built-in components from external libraries. Built-in rows are checked and locked. External libraries are off by default and, when enabled, add packages, configure flags, license effects, and any relevant warnings.

The **Options** page exposes common `configure` choices as named rows. Manual advanced flags are available as an escape hatch and appear in the review before you confirm.

### Requirements

- Windows 10 or later (Windows-only; planning is blocked on other hosts)
- An internet connection for downloading sources and packages
- ~Several GB of free disk space in the workspace for the toolchain and build

The toolchain uses a private MSYS2 archive inside your workspace:

- Default shell profile: `ucrt64` (also supports `mingw64`, `clang64`)
- Accepts MSYS2 `.tar.zst` (preferred) and `.tar.xz` archives — never the `.exe` installer

### Download

Grab the latest build from the [Releases page](https://github.com/magnomer/Prompful-Custom-FFmpeg-builder/releases/latest).

### Build from source

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
cd frontend
npm install
cd ..
wails dev
```

### License

The program builds FFmpeg from official sources; the resulting binary's license depends on the libraries and options you select. The Library page shows the derived license profile (e.g. GPL effects) before you build.

---

<a id="korean"></a>

## 한국어

### 스크린샷

<img width="1349" height="953" alt="Screenshot-KO-1" src="https://github.com/user-attachments/assets/401ef4b5-23ef-4d1e-8d29-57750dab97ae" />
<img width="1349" height="953" alt="Screenshot-KO-2" src="https://github.com/user-attachments/assets/c487b728-73d0-49cb-9cfe-45168a697a34" />
<img width="1349" height="953" alt="Screenshot-KO-3" src="https://github.com/user-attachments/assets/9531c660-ce9a-40bc-9841-3aec962280ee" />
<img width="1349" height="953" alt="Screenshot-KO-4" src="https://github.com/user-attachments/assets/2e4f3d67-8ad7-45bc-af3e-6a54d2e465fa" />

### 이 프로그램은

Promptful Custom FFmpeg Builder는 FFmpeg 로컬 빌드를 자동화하는 Windows GUI 프로그램입니다. 라이브러리와 옵션을 고르고, 빌드 계획을 검토하여 커스텀 빌드를 작성할 수 있도록 도와주는 것을 목표로 합니다. 빌드 계획에 이상이 없을 경우 소스 파일을 다운로드하여 작업 공간 내에 비공개 MSYS2 빌드 환경을 구성한 다음, ffmpeg를 빌드합니다. 최종적으로 완성된 바이너리는 결과 폴더로 복사됩니다.

해당 프로그램은 **FFmpeg 바이너리, 소스, 혹은 코덱 라이브러리를 제공하지 않습니다.** 제공 받은 공식 출처 링크를 통해 다운로드 후 로컬 환경에서 빌드를 구성합니다.

### 작동 방식

모든 빌드는 계획 검토 후 사용자의 승인이 이루어졌을 때만 이행됩니다.

1. 라이브러리와 configure 옵션을 선택합니다.
2. 프로그램이 빌드 계획을 생성합니다. 이 계획에는 패키지 종류, configure 플래그, 스크립트, 라이선스 영향, 경고 사항 등이 담겨 있습니다.
3. 계획을 검토한 뒤 사용자가 승인합니다.
4. 실행 직전, Windows 기본 확인 대화상자가 한 번 확인합니다.
5. 빌드가 실행 후에는 로그가 실시간으로 제공되며 언제든 중도에 취소할 수 있습니다.

다운로드한 파일은 빌드 개시 전에 검증 절차를 거칩니다. FFmpeg 소스는 PGP 서명으로, MSYS2 아카이브는 분리된(detached) 서명으로 확인합니다. 모든 작업은 작업 공간 폴더 안에서만 이루어지며, 시스템 전역 설치나 관리자 권한이 필요하지 않습니다.

### FFmpeg 빌드 단계

1. FFmpeg 소스 아카이브, `.asc` 서명, 릴리스 서명 키를 다운로드합니다.
2. 서명을 검증한 뒤 소스를 압축 해제합니다.
3. 선택한 라이브러리에 필요한 MSYS2 패키지를 설치합니다.
4. `configure` 스크립트를 생성·해시하고 비공개 MSYS2 Bash로 실행합니다.
5. `make` 스크립트를 생성·해시하고 선택한 병렬 작업 수로 실행합니다.
6. `ffmpeg.exe`, `ffprobe.exe`, 필요한 DLL을 `workspace/FFmpeg/`로 복사합니다.
7. 선택한 라이브러리, 플래그, 라이선스 프로필, 파일 크기, SHA-256 해시를 담은 `build-report-<runId>.json`을 작성합니다.

### 라이브러리 및 옵션

**Library** 페이지는 FFmpeg의 내장 구성 요소와 외부 라이브러리가 별도로 구분하여 표시되어 있습니다. 내장 항목은 이미 체크되어 있으며 수정이 불가능합니다. 외부 라이브러리는 기본적으로는 체크되어 있지 않으나, 선택 시 패키지, configure 플래그, 라이선스 영향, 관련 경고 등이 추가될 수 있습니다.

**Options** 페이지는 `configure` 옵션을 보여줍니다. 그리고 검토 화면에 다시 표시됩니다.

### 실행 환경

- Windows 10 이상 (Windows 전용, 다른 호스트에서는 계획 단계가 차단됨)
- 소스 및 패키지 다운로드를 위한 인터넷 연결
- 툴체인과 빌드를 위한 작업 공간 여유 디스크 공간 수 GB

툴체인은 작업 공간 안의 비공개 MSYS2 아카이브를 사용합니다:

- 기본 셸 프로필: `ucrt64` (`mingw64`, `clang64`도 지원)
- MSYS2 `.tar.zst`(권장) 및 `.tar.xz` 아카이브를 받음 — `.exe` 설치 프로그램은 사용하지 않음

### 다운로드

[최신 빌드](https://github.com/magnomer/Prompful-Custom-FFmpeg-builder/releases/latest)

### 소스에서 빌드

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
cd frontend
npm install
cd ..
wails dev
```

### 라이선스

프로그램은 공식 출처에서 FFmpeg를 다운로드 하여 빌드하며, 빌드 된 바이너리의 라이선스는 선택한 라이브러리와 옵션에 따라 달라집니다. Library 페이지에서 빌드 전에 파생된 라이선스 프로필(예: GPL)을 확인할 수 있습니다.
