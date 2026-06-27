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

Promptful Custom FFmpeg Builder is a Windows program that helps you build FFmpeg from source without setting up the build environment by hand.

It creates a private MSYS2 toolchain inside your selected workspace, prepares packages and build scripts, and runs the build only after you review and approve the generated plan.

### What it does not include

This repository does not provide ready-made files or bundled build dependencies.

- Prebuilt FFmpeg binaries.
- FFmpeg source archives.
- Codec libraries.
- Bundled MSYS2 installation.
- System-wide build environment.

### How it works

Every build starts with a plan, so you can check what will happen before anything runs.

- Choose a workspace, FFmpeg source, libraries, and options.
- Select the private MSYS2 build profile.
- Generate a build plan.
- Review packages, configure flags, scripts, warnings, and license effects.
- Approve the plan.
- Confirm once more in a native Windows dialog.
- Watch the live build log.

### Build environment

The build environment is created inside your selected workspace.

- Default private MSYS2 profile: `ucrt64`.
- Other supported private profiles: `mingw64`, `clang64`.
- Accepted MSYS2 archives: `.tar.zst`, `.tar.xz`.
- The MSYS2 `.exe` installer is not used.

### Build output

The finished FFmpeg files and build report are saved inside your workspace.

- `ffmpeg.exe` is copied to `workspace/FFmpeg/`.
- `ffprobe.exe` is copied when available.
- Required DLL files are copied with the executables.
- A JSON build report is saved.
- The report records libraries, flags, license profile, sizes, and hashes.

### Libraries

The program currently tracks 125 FFmpeg library choices, including always-included FFmpeg parts, selectable add-on libraries, and entries kept only to explain current limits.

- [Library model](docs/LIBRARY_MODEL_EN.md)
- [Library coverage report](docs/LIBRARY_COVERAGE_REPORT_EN.md)

### Options

FFmpeg build options are treated as part of the generated build plan, not as hidden command-line text.

Common configure choices are exposed as named options, while advanced manual flags remain available for unusual cases. Before a build runs, the final configure flags are shown again for review.

### Requirements

Users only need to prepare a normal Windows environment, internet access, and enough disk space.

- Windows 10 or later.
- Internet access.
- Several GB of free disk space.

### Build from source

These commands start the development version.

- `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- `cd frontend`
- `npm install`
- `cd ..`
- `wails dev`

### Download

- [Latest release](https://github.com/magnomer/Prompful-Custom-FFmpeg-builder/releases/latest)

### Documentation

- [Consent boundary](docs/CONSENT_BOUNDARY_EN.md)
- [Transparency model](docs/TRANSPARENCY_EN.md)

---

<a id="korean"></a>

## 한국어

### 스크린샷

<img width="1349" height="953" alt="Screenshot-KO-1" src="https://github.com/user-attachments/assets/401ef4b5-23ef-4d1e-8d29-57750dab97ae" />
<img width="1349" height="953" alt="Screenshot-KO-2" src="https://github.com/user-attachments/assets/c487b728-73d0-49cb-9cfe-45168a697a34" />
<img width="1349" height="953" alt="Screenshot-KO-3" src="https://github.com/user-attachments/assets/9531c660-ce9a-40bc-9841-3aec962280ee" />
<img width="1349" height="953" alt="Screenshot-KO-4" src="https://github.com/user-attachments/assets/2e4f3d67-8ad7-45bc-af3e-6a54d2e465fa" />

### 이 프로그램은?

Promptful Custom FFmpeg Builder는 윈도우에서 FFmpeg를 소스부터 직접 빌드할 수 있게 도와주는 데스크톱 프로그램입니다.

사용자가 선택한 작업 폴더 안에 전용 MSYS2 빌드 환경을 만들고, 필요한 패키지와 빌드 스크립트를 준비한 뒤, 사용자가 생성된 계획을 검토하고 사용자로부터 승인을 받은 경우 FFmpeg를 빌드합니다.

### 다음은 포함되어 있지 않습니다

이 프로그램은 완성된 실행 파일이나 다음과 같이 빌드에 필요한 외부 구성 요소가 기본적으로 포함되어 있지 않습니다.

- 미리 빌드된 FFmpeg 실행 파일
- FFmpeg 소스 압축 파일
- 코덱 라이브러리
- MSYS2 설치본

이 프로그램은 MSYS2나 FFmpeg 소스 파일, 라이브러리 등을 로컬에서 다운로드 한 후 FFmpeg를 빌드합니다.

### 작동 방식

먼저 빌드 계획을 만듭니다. 빌드 실행 전에 빌드 진행 내용을 검토할 수 있습니다.

- 작업 폴더, FFmpeg 소스, 라이브러리, 옵션을 선택합니다.
- 전용 MSYS2 빌드 프로필을 선택합니다.
- 빌드 계획을 생성합니다.
- 패키지, configure 플래그, 스크립트, 경고, 라이선스 영향을 확인합니다.
- 빌드 계획을 승인합니다.
- 윈도우 기본 확인 창에서 한 번 더 확인합니다.
- 실시간 빌드 로그를 보며 진행 상태를 확인합니다.

### 빌드 환경

빌드 환경은 사용자가 선택한 작업 폴더 안에 만들어집니다.

- 기본 전용 MSYS2 프로필: `ucrt64`.
- 추가 지원 프로필: `mingw64`, `clang64`.
- 지원하는 MSYS2 압축 파일: `.tar.zst`, `.tar.xz`.
- MSYS2 `.exe` 설치 파일은 사용하지 않습니다.

### 빌드 결과

완성된 FFmpeg 파일과 빌드 보고서는 작업 폴더 안에 저장됩니다.

- `ffmpeg.exe`는 `workspace/FFmpeg/`에 복사됩니다.
- `ffprobe.exe`는 생성된 경우 함께 복사됩니다.
- 필요한 DLL 파일도 실행 파일과 함께 복사됩니다.
- JSON 형식의 빌드 보고서가 저장됩니다.
- 보고서에는 라이브러리, 플래그, 라이선스 프로필, 크기, 해시가 기록됩니다.

### 라이브러리

이 프로그램은 FFmpeg 빌드에서 확인하거나 선택할 수 있는 라이브러리 항목 125개를 관리합니다. 기본으로 포함되는 항목, 사용자가 선택할 수 있는 추가 라이브러리, 현재 한계를 설명하기 위해 남겨 둔 항목 등이 포함됩니다.

- [라이브러리 모델](docs/LIBRARY_MODEL_KO.md)
- [라이브러리 지원 범위 보고서](docs/LIBRARY_COVERAGE_REPORT_KO.md)

### 옵션

FFmpeg 빌드 옵션은 숨겨진 명령줄이 아니라, 빌드 계획의 일부로 정리됩니다.

자주 쓰는 configure 선택지는 이름이 붙은 옵션으로 제공됩니다. 특수한 상황을 위한 수동 플래그도 사용할 수 있으며, 실제 빌드 전에 최종 configure 플래그가 다시 표시됩니다.

### 요구 사항

사용자가 미리 준비해야 하는 것은 일반적인 윈도우 환경, 인터넷 연결, 충분한 저장 공간입니다.

- Windows 10 이상.
- 인터넷 연결.
- 몇 GB 이상의 여유 디스크 공간.

### 소스에서 실행하기

다음 명령으로 개발 버전을 실행할 수 있습니다.

- `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- `cd frontend`
- `npm install`
- `cd ..`
- `wails dev`

### 다운로드

- [최신 릴리스](https://github.com/magnomer/Prompful-Custom-FFmpeg-builder/releases/latest)

### 문서

- [사용자 동의 관리](docs/CONSENT_BOUNDARY_KO.md)
- [투명성](docs/TRANSPARENCY_KO.md)
