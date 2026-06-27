# 라이브러리 지원 범위 보고서

이 문서는 이 프로그램이 어떤 FFmpeg 관련 라이브러리를 어떻게 다루는지 정리한 보고서입니다.

여기서 지원하는 항목은 단순히 FFmpeg configure 옵션 이름이 포함되어 있다는 것이 아닙니다. 이 프로그램에서는 라이브러리를 **기본 포함 항목**, **MSYS2 패키지로 설치 가능한 항목**, **프로그램이 직접 소스/아카이브를 준비하는 항목**, **현재는 의도적으로 막아 둔 항목**으로 나누어져 있습니다

## 요약

- 전체 카탈로그 행: **125**.
- 공식 FFmpeg 소스에 기본 포함되어 고정되는 행: **10**.
- MSYS2 패키지 경로를 쓰는 행: **96**.
- 프로그램 내부에서 소스/아카이브를 준비하는 행: **12**.
- 외부 SDK 또는 별도 가져오기 경로로 분류된 행: **7**.
- 고정 기본 항목을 포함해 일반 UI에서 정상 선택 가능한 행: **112**.
- 고정 기본 항목을 제외한 추가 라이브러리 중 정상 선택 가능한 행: **102**.
- UI에서 비활성화된 행: **5 — `lensfun`, `svtjpegxs`, `vapoursynth`, `tensorflow`, `onnxruntime`**.
- 일반 빌드/가져오기 절차가 없어 차단되는 행: **8 — `smbclient`, `openvino`, `torch`, `libmfx`, `pocketsphinx`, `dc1394`, `decklink`, `cuda-nvcc`**.


## 사용 용어

| 구분 | 이 프로그램에서의 의미 | 실제 동작 |
|---|---|---|
| 기본 포함 | FFmpeg 소스 자체에 포함되는 프로그램·라이브러리·기본 코덱/포맷입니다. | 사용자가 끌 수 없는 고정 항목으로 취급됩니다. 별도 패키지나 `--enable-lib...` 플래그를 붙이지 않습니다. |
| MSYS2 패키지 경로 | MSYS2의 mingw/ucrt/clang 계열 패키지와 FFmpeg configure 플래그가 연결되는 항목입니다. | 선택하면 빌드 전에 해당 프로필의 패키지를 설치하고, configure 단계에서 필요한 플래그를 추가합니다. |
| 내부 소스 준비 경로 | 적절한 사전 빌드 패키지에 기대지 않고 프로그램이 고정된 소스/아카이브를 받아 준비하는 항목입니다. | 다운로드, 해시 검증, 압축 해제, 빌드 또는 가져오기, 헤더/라이브러리 확인을 거친 뒤 FFmpeg configure에 넘깁니다. |
| 외부 SDK/가져오기 경로 | 독점 SDK, 별도 런타임, 또는 아직 안전한 준비 절차가 없는 항목입니다. | 정상 선택 대상으로 보지 않거나, 선택 시 빌드 계획을 차단합니다. |

이 모델 때문에 카탈로그에는 “지금 바로 빌드 가능한 것”과 “알고 있지만 일부러 막아 둔 것”이 함께 들어 있습니다. 이는 지원을 부풀리기 위한 장치가 아니라, 왜 어떤 항목이 빠졌거나 비활성화되어 있는지 추적하기 위한 장치입니다.

## 카테고리 일람

| 분야 | 행 수 | 해석 |
|---|---:|---|
| 기본 포함 항목(공식 FFmpeg 소스) | 10 | ffmpeg.exe, ffprobe.exe, libav* 계열처럼 FFmpeg 소스 자체에 들어 있는 고정 항목입니다. |
| 비디오 인코더 | 16 | H.264/H.265/AV1/EVC/VVC/AVS/APV 계열의 소프트웨어 인코딩 경로가 중심입니다. |
| 하드웨어 인코더 | 4 | GPU/드라이버가 실제로 지원해야 동작하는 헤더·API 활성화 항목입니다. |
| 비디오 디코더 | 7 | AV1, EVC, AVS2/AVS3, LCEVC, AviSynth 계열 디코딩·프레임 공급 경로입니다. |
| 이미지 코덱 | 8 | PNG, WebP, JPEG 2000, JPEG XL, SVG 렌더링, 색 관리 등 이미지 기반 작업을 다룹니다. |
| 필터와 처리 | 17 | 스케일링, 품질 측정, GPU 처리, 색 관리, 플러그인, QR, 스크립트 처리과 관련됩니다. |
| 오디오 | 22 | 오디오 코덱, 리샘플러, 음성·공간 오디오, 분석·식별 관련 항목을 포함합니다. |
| 자막과 텍스트 | 8 | 자막 렌더링, 글꼴 처리, 양방향 문자, 방송 자막 처리를 담당합니다. |
| 디스크 및 장치 입력 | 14 | 디스크, 장치, 출력 디바이스, 캡처 SDK, 레거시 모듈 입력을 묶습니다. |
| 네트워크 | 7 | 스트리밍·원격 접근·메시징 프로토콜입니다. TLS 백엔드는 별도 섹션에서 다룹니다. |
| 보안 네트워크(TLS) | 4 | HTTPS/TLS 계열 네트워크 접근을 위한 백엔드 후보입니다. |
| OCR | 1 | 영상·이미지 안의 글자를 읽는 OCR 경로입니다. |
| AI 지원 | 4 | FFmpeg의 AI/모델 추론 관련 연결 지점을 추적하지만 현재는 보수적으로 막은 항목이 많습니다. |
| 지원 라이브러리 | 3 | XML, QR, 방송 메타데이터처럼 다른 기능을 보조하는 라이브러리입니다. |

## 프리셋 일람

공개 프리셋은 단계적으로 넓어지는 구조입니다. `최소 구성`은 FFmpeg 기본 포함 항목만 둡니다. `기본`부터 실사용 빈도가 높은 외부 라이브러리를 붙이고, 이후 프리셋은 효율, 호환성, 편집/처리, 전체 기능 쪽으로 범위를 넓힙니다. 

| 프리셋 | 성격 | 추가되는 대표 항목 |
|---|---|---|
| 최소 구성 | 가장 작은 기준점입니다. FFmpeg 소스에 기본 포함되는 고정 항목만 남깁니다. | 없음 |
| 기본 | 일반 사용자가 처음 빌드할 때의 실용 기준입니다. 대표 비디오/오디오 코덱, 자막 구성, 기본 네트워크·처리 라이브러리를 포함합니다. | `x264`, `x265`, `svt-av1`, `dav1d`, `opus`, `mp3lame`, `ass`, `freetype`, `zimg`, `vmaf`, `srt`, `ssh` 등 |
| 최대 효율 | 기본에 품질·효율 중심 오디오 도구를 더합니다. | `fdk-aac`, `soxr` |
| 최대 호환성 | 오래된 포맷, 방송 자막, 추가 프로토콜, 보조 코덱까지 넓힙니다. | `rav1e`, `openh264`, `ilbc`, `twolame`, `xevd`, `shine`, `codec2`, `lc3`, `aribb24`, `aribcaption`, `rtmp` 등 |
| 오디오/비디오 편집 | 편집, 필터, 색 관리, 이미지 작업 흐름, OCR/전사 인접 기능을 강화합니다. | `libjxl`, `libplacebo`, `frei0r`, `mysofa`, `shaderc`, `opencv`, `opencolorio`, `ladspa`, `lv2`, `qrencode`, `whisper` 등 |
| 전체 | 디스크/장치, 추가 네트워크, TLS, OCR, 표시·계산 보조 기능까지 가장 넓게 잡습니다. | `kvazaar`, `bluray`, `dvdread`, `dvdnav`, `cdio`, `openssl`, `rist`, `rabbitmq`, `tesseract`, `opencl` 등 |

### 확장 옵션

확장 옵션은 소스 빌드 또는 사용자 지정 빌드가 필요한 라이브러리를 일부 프리셋에 덧붙이는 장치입니다. `최소 구성`과 `기본`은 의도적으로 영향을 받지 않습니다.

| 확장 옵션을 켠 기본 프리셋 | 추가되는 내부 준비 항목 |
|---|---|
| 최대 효율 | `vvenc`, `uavs3d`, `lcevc-dec` |
| 최대 호환성 | `davs2`, `uavs3d`, `lcevc-dec`, `avisynthplus`, `xavs2` |
| 오디오/비디오 편집 | `avisynthplus`, `lcevc-dec` |
| 전체 | `vvenc`, `xavs2`, `davs2`, `uavs3d`, `lcevc-dec`, `avisynthplus`, `mpeghdec` |

## 충돌 방지

일부 라이브러리는 FFmpeg에서 별도 연동 항목으로 보이지만, 실제 빌드 계획에서는 동시에 활성화 할 수 없는 조합입니다. 이 프로그램은 최종 configure 전에 다음과 같이 정리합니다.

- TLS 백엔드: `openssl`, `gnutls`, `mbedtls`, `libtls`는 하나만 선택합니다. 기본 우선순위는 OpenSSL, GnuTLS, mbedTLS, libtls 순입니다.
- 셰이더 컴파일러: `shaderc`와 `glslang`이 함께 잡히면 `shaderc`를 남깁니다.
- EVC 디코더 연동: `xevd`와 `xevdb`는 함께 켜지지 않습니다. 프리셋의 경우 `xevd`를 남깁니다.
- EVC 인코더 연동: `xeve`와 `xeveb`도 함께 켜지지 않습니다. 프리셋의 경우 `xeve`를 남깁니다.

## 소스 준비 또는 가져오기 절차가 구현된 항목

아래 항목은 단순히 configure 플래그만 추가하는 방식이 아닙니다. 프로그램이 고정된 버전의 소스 또는 아카이브를 받아 해시를 검증하고, 필요한 경우 빌드·설치·가져오기 확인을 거친 뒤 FFmpeg 빌드에 연결합니다.

| ID | 표시 이름 | 경로 | 고정 버전 | 준비 방식 | 빌드/가져오기 방식 | 일반 UI 상태 |
|---|---|---|---|---|---|---|
| `avisynthplus` | AviSynth+ / 스크립트 기반 비디오 처리 | 내부 소스 준비 경로 | 3.7.5 | 내부 소스 빌드 | CMake | 일반 선택 가능: 소스 준비/가져오기 |
| `davs2` | libdavs2 / AVS2 디코딩 | 내부 소스 준비 경로 | 1.7 | 내부 소스 빌드 | configure + make | 일반 선택 가능: 소스 준비/가져오기 |
| `xavs2` | xavs2 / AVS2 인코딩 | 내부 소스 준비 경로 | 1.4 | 내부 소스 빌드 | configure + make | 일반 선택 가능: 소스 준비/가져오기 |
| `uavs3d` | libuavs3d / AVS3 디코딩 | 내부 소스 준비 경로 | master | 내부 소스 빌드 | CMake | 일반 선택 가능: 소스 준비/가져오기 |
| `lcevc-dec` | liblcevc-dec / LCEVC 디코딩 | 내부 소스 준비 경로 | 4.2.0 | 내부 소스 빌드 | CMake | 일반 선택 가능: 소스 준비/가져오기 |
| `vvenc` | vvenc / VVC/H.266 인코딩 | 내부 소스 준비 경로 | 1.14.0 | 내부 소스 빌드 | CMake | 일반 선택 가능: 소스 준비/가져오기 |
| `mpeghdec` | libmpeghdec / MPEG-H 오디오 디코딩 | 내부 소스 준비 경로 | r3.0.3 | 내부 소스 빌드 | CMake | 일반 선택 가능: 소스 준비/가져오기 |
| `quirc` | libquirc / QR 코드 디코딩 | 내부 소스 준비 경로 | 1.2 | 내부 소스 빌드 | make | 일반 선택 가능: 소스 준비/가져오기 |
| `klvanc` | libklvanc / 방송 메타데이터 | 내부 소스 준비 경로 | vid.obe.1.6.0 | 내부 소스 빌드 | configure + make | 일반 선택 가능: 소스 준비/가져오기 |
| `libtls` | libtls / 보안 네트워크 접근 | 내부 소스 준비 경로 | 4.3.2 | 내부 소스 빌드 | CMake | 일반 선택 가능: 소스 준비/가져오기 |
| `tensorflow` | TensorFlow / AI 모델 추론 | 외부 SDK/가져오기 경로 | 2.16.1 | 외부 아카이브 가져오기 | 아카이브 가져오기 | UI 비활성화 |

## 의도적으로 비활성화되거나 차단된 항목

이 표는 “빠뜨린 항목” 목록이 아닙니다. 현재 프로그램의 원칙, 즉 **FFmpeg 소스 코드를 호환성 확보 목적으로 임의 수정하지 않고, 검토 가능한 로컬 빌드 절차 없이 외부 SDK·런타임에 기대지 않는다**는 기준 때문에 막아 둔 항목입니다.

| ID | 표시 이름 | 경로 | 현재 판단 |
|---|---|---|---|
| `cuda-nvcc` | cuda-nvcc / NVIDIA CUDA 필터 컴파일 | 외부 SDK/가져오기 경로 | 현재 프로그램에 일반 빌드/가져오기 절차가 없어 선택 시 빌드 계획이 차단됩니다. FFmpeg의 CUDA 가속 필터를 NVIDIA nvcc 컴파일러로 컴파일합니다. nvcc는 독점 CUDA Toolkit에만 포함되어 있으며 MSYS2 패키지로 제공되지 않습니다. |
| `dc1394` | dc1394 / IEEE 1394 카메라 캡처 | 내부 소스 준비 경로 | 현재 프로그램에 일반 빌드/가져오기 절차가 없어 선택 시 빌드 계획이 차단됩니다. IEEE 1394(FireWire) 카메라에서 비디오를 캡처합니다. Windows에서는 차단됩니다: MSYS2 패키지가 없으며, 유일한 Windows 빌드 방법은 독점 FireWire 커널 드라이버와 FireWire 하드웨어를 필요로 하므로 결과물이 이식 가능하지 않습니다. |
| `decklink` | decklink / Blackmagic 캡처 및 재생 | 외부 SDK/가져오기 경로 | 현재 프로그램에 일반 빌드/가져오기 절차가 없어 선택 시 빌드 계획이 차단됩니다. Blackmagic DeckLink 캡처 및 재생 지원을 추가합니다. 독점 DeckLink SDK 헤더를 대상으로 빌드되며, 이 헤더는 MSYS2 패키지로 재배포할 수 없습니다. |
| `lensfun` | lensfun / 렌즈 보정 | MSYS2 패키지 경로 | 일반 UI에서는 비활성화되어 있습니다. 렌즈 왜곡, 비네팅, 카메라 렌즈 특유의 결함을 보정합니다. 알려진 렌즈로 촬영한 영상 정리에 유용합니다. |
| `libmfx` | libmfx / 레거시 Intel Media SDK (FFmpeg 7.0+에서 제거됨) | 외부 SDK/가져오기 경로 | 현재 프로그램에 일반 빌드/가져오기 절차가 없어 선택 시 빌드 계획이 차단됩니다. 오래된 Media SDK 디스패처를 통한 레거시 Intel 하드웨어 인코드/디코드입니다. FFmpeg가 7.0에서 지원을 제거하여 이 빌더가 대상으로 하는 이후 소스에서는 작동하지 않습니다. 대신 Intel QSV(oneVPL)를 사용하세요. |
| `onnxruntime` | ONNX Runtime / AI 모델 추론 | MSYS2 패키지 경로 | 일반 UI에서는 비활성화되어 있으며, mingw64 프로필에서는 패키지 조건 때문에 카탈로그에서 제외됩니다. ONNX Runtime으로 지원되는 딥러닝 필터를 실행합니다. 모델 기반 분석이나 향상 작업에 유용합니다. |
| `openvino` | OpenVINO / AI 모델 추론 | 외부 SDK/가져오기 경로 | 현재 프로그램에 일반 빌드/가져오기 절차가 없어 선택 시 빌드 계획이 차단됩니다. Intel 계열 가속에 맞춘 AI 추론 필터를 실행합니다. 모델 기반 영상 또는 이미지 처리에 유용합니다. |
| `pocketsphinx` | pocketsphinx / 음성 인식 | 내부 소스 준비 경로 | 현재 프로그램에 일반 빌드/가져오기 절차가 없어 선택 시 빌드 계획이 차단됩니다. CMU PocketSphinx를 사용하는 asr 음성 인식 오디오 필터를 추가합니다. 현재는 빌드할 수 없습니다. FFmpeg의 asr 필터가 최신 PocketSphinx 릴리스와 호환되지 않으므로, 선택하면 빌드가 차단됩니다. |
| `smbclient` | libsmbclient / SMB 네트워크 파일 접근 | 외부 SDK/가져오기 경로 | 현재 프로그램에 일반 빌드/가져오기 절차가 없어 선택 시 빌드 계획이 차단됩니다. SMB/CIFS 네트워크 공유에서 미디어를 읽고 쓸 수 있게 합니다. Windows식 네트워크 폴더에 저장된 미디어에 유용합니다. 아직 Windows에서 빌드할 수 없어 이 섹션 맨 아래에 표시되며, Windows용 libsmbclient가 나올 때까지 사용할 수 없습니다. |
| `svtjpegxs` | svtjpegxs / JPEG XS 인코딩 | MSYS2 패키지 경로 | 일반 UI에서는 비활성화되어 있습니다. 저지연 전문 미디어 작업을 위한 JPEG XS 영상을 만듭니다. |
| `tensorflow` | TensorFlow / AI 모델 추론 | 외부 SDK/가져오기 경로 | 일반 UI에서는 비활성화되어 있습니다. TensorFlow C API로 지원되는 딥러닝 필터를 실행합니다. 모델 기반 이미지나 영상 분석에 유용합니다. |
| `torch` | Torch / libtorch | 외부 SDK/가져오기 경로 | 현재 프로그램에 일반 빌드/가져오기 절차가 없어 선택 시 빌드 계획이 차단됩니다. Torch 기반 모델 실행으로 지원되는 딥러닝 필터를 돌립니다. PyTorch 계열 추론 작업에 유용합니다. |
| `vapoursynth` | VapourSynth / 스크립트 기반 비디오 처리 | MSYS2 패키지 경로 | 일반 UI에서는 비활성화되어 있습니다. VapourSynth 스크립트 기반 영상 처리 체인을 엽니다. 고급 스크립트 필터링 결과를 미디어 작업에 넣을 때 유용합니다. |

## 전체 라이브러리 카탈로그

아래 표는 현재 백엔드 카탈로그를 기준으로 정리한 전체 목록입니다. 패키지 이름의 `<profile>`은 선택된 MSYS2 셸 프로필의 패키지 접두사를 뜻합니다. 예를 들어 `mingw-w64-ucrt-x86_64`, `mingw-w64-x86_64`, `mingw-w64-clang-x86_64` 같은 값으로 바뀝니다.

### 기본 포함 항목(공식 FFmpeg 소스)

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `ffmpeg-program` | ffmpeg.exe | FFmpeg 기본 구성 | 기본 포함 / 고정 | 기본 포함 | 없음 | FFmpeg 소스에 포함 | 변환, 필터링, 녹화, 스트리밍, 패키징을 수행하는 기본 명령줄 미디어 처리 프로그램입니다. |
| `ffprobe-program` | ffprobe.exe | FFmpeg 기본 구성 | 기본 포함 / 고정 | 기본 포함 | 없음 | FFmpeg 소스에 포함 | 미디어 파일의 스트림, 코덱, 메타데이터, 챕터, 길이, 내부 구조를 검사합니다. |
| `libavutil` | libavutil | FFmpeg 기본 구성 | 기본 포함 / 고정 | 기본 포함 | 없음 | FFmpeg 소스에 포함 | 코덱, 필터, 포맷, 도구 전반에서 공유되는 FFmpeg 공통 유틸리티 코드입니다. |
| `libavcodec` | libavcodec | FFmpeg 기본 구성 | 기본 포함 / 고정 | 기본 포함 | 없음 | FFmpeg 소스에 포함 | 다양한 미디어 형식의 디코딩과 인코딩을 담당하는 FFmpeg의 기본 코덱 라이브러리입니다. |
| `libavformat` | libavformat | FFmpeg 기본 구성 | 기본 포함 / 고정 | 기본 포함 | 없음 | FFmpeg 소스에 포함 | MP4, MOV, MKV, WAV, MPEG-TS 같은 컨테이너를 읽고 쓰는 FFmpeg의 기본 컨테이너 라이브러리입니다. |
| `libavfilter` | libavfilter | FFmpeg 기본 구성 | 기본 포함 / 고정 | 기본 포함 | 없음 | FFmpeg 소스에 포함 | 영상, 오디오, 자막, 스케일링, 오버레이, 효과 처리를 담당하는 FFmpeg의 기본 필터 라이브러리입니다. |
| `libswscale` | libswscale | FFmpeg 기본 구성 | 기본 포함 / 고정 | 기본 포함 | 없음 | FFmpeg 소스에 포함 | 영상 크기와 픽셀 형식을 변환합니다. |
| `libswresample` | libswresample | FFmpeg 기본 구성 | 기본 포함 / 고정 | 기본 포함 | 없음 | FFmpeg 소스에 포함 | 오디오 샘플레이트, 샘플 형식, 채널 레이아웃을 변환합니다. |
| `native-codecs` | FFmpeg 기본 코덱 | FFmpeg 기본 구성 | 기본 포함 / 고정 | 기본 포함 | 없음 | FFmpeg 소스에 포함 | 외부 코덱 라이브러리를 더하기 전에 FFmpeg 내부 코덱 기능을 사용합니다. |
| `native-formats` | 기본 형식 및 muxer | FFmpeg 기본 구성 | 기본 포함 / 고정 | 기본 포함 | 없음 | FFmpeg 소스에 포함 | 많은 미디어 컨테이너를 읽고 쓰는 FFmpeg 내부 포맷 기능을 사용합니다. |

### 비디오 인코더

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `x264` | x264 / H.264 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | GPL 경계 | `--enable-libx264` | `<profile>-libx264` | H.264 소프트웨어 인코딩을 지원합니다. 품질, 성능, 폭넓은 호환성으로 잘 알려진 대표적인 H.264/AVC 인코더입니다. |
| `x265` | x265 / H.265 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | GPL 경계 | `--enable-libx265` | `<profile>-x265` | 비슷한 체감 품질에서 H.264보다 작은 파일을 목표로 하는 HEVC/H.265 소프트웨어 인코딩을 지원합니다. 압축 효율이 중요할 때 유용합니다. |
| `svt-av1` | svt-av1 / AV1 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libsvtav1` | `<profile>-svt-av1` | 실용적인 속도로 AV1 영상을 인코딩합니다. 최신 압축 효율과 사용 가능한 처리량을 함께 원할 때 자주 쓰입니다. |
| `libvpx` | libvpx / VP8 및 VP9 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libvpx` | `<profile>-libvpx` | WebM과 웹 배포에 자주 쓰이는 VP8, VP9 영상을 만듭니다. |
| `aom` | aom / AV1 레퍼런스 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libaom` | `<profile>-aom` | AV1 참조 계열 코덱으로 AV1 인코딩을 지원합니다. 속도보다 품질과 표준 지향적인 결과가 중요한 경우에 어울립니다. |
| `openh264` | openh264 / OpenH264 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libopenh264` | `<profile>-openh264` | x264보다 단순한 성격의 H.264 소프트웨어 인코딩을 제공합니다. 기본적인 H.264 출력이 필요할 때 유용합니다. |
| `rav1e` | rav1e / AV1 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-librav1e` | `<profile>-rav1e` | Rust 기반 AV1 소프트웨어 인코딩을 제공합니다. AV1의 속도, 품질, 인코더 특성을 비교할 때 유용합니다. |
| `xvid` | xvid / MPEG-4 Part 2 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | GPL 경계 | `--enable-libxvid` | `<profile>-xvidcore` | 오래된 플레이어와 기기를 위한 MPEG-4 Part 2 영상을 만듭니다. |
| `theora` | theora / Theora 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libtheora` | `<profile>-libtheora` | 오래된 개방형 비디오 호환성을 위한 Ogg Theora 영상을 만듭니다. |
| `kvazaar` | Kvazaar (HEVC) | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libkvazaar` | `<profile>-kvazaar` | x265가 아닌 경로로 HEVC/H.265 소프트웨어 인코딩을 제공합니다. 다른 HEVC 인코더 특성이 필요할 때 유용합니다. |
| `xeve` | xeve / EVC 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libxeve` | `<profile>-xeve` | 테스트와 새로운 코덱 워크플로를 위한 EVC 영상을 만듭니다. |
| `xeveb` | xeveb / EVC 기본 프로파일 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libxeveb` | `<profile>-xeve` | EVC의 로열티 없는 부분집합인 EVC 기본 프로파일로 MPEG-5 EVC 인코딩을 활성화합니다. 메인 xeve 행과 동일한 XEVE 패키지를 사용합니다. |
| `oapv` | oapv / APV 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-liboapv` | `<profile>-openapv` | 전문 영상 작업에 쓰이는 APV 인코딩을 지원합니다. 특수한 프로덕션 워크플로에 유용합니다. |
| `xavs` | xavs / AVS1 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | GPL 경계 | `--enable-libxavs` | `<profile>-xavs` | AVS1(중국 AVS) 형식으로 비디오를 인코딩합니다. AVS1 방송 또는 보관 자료와 호환할 때 유용합니다. GPL이므로 빌드가 GPL 라이선스 경계로 이동합니다. |
| `vvenc` | vvenc / VVC/H.266 인코딩 | 내부 소스 준비 경로 | 일반 선택 가능: 소스 준비/가져오기 | LGPL 안전 경계 | `--enable-libvvenc` | 고정된 소스/가져오기 절차로 준비 | 차세대 코덱 테스트와 높은 압축 효율 실험을 위한 VVC/H.266 영상을 만듭니다. |
| `xavs2` | xavs2 / AVS2 인코딩 | 내부 소스 준비 경로 | 일반 선택 가능: 소스 준비/가져오기 | GPL 경계 | `--enable-libxavs2` | 고정된 소스/가져오기 절차로 준비 | AVS2 표준이 필요한 작업을 위해 AVS2 영상을 만듭니다. 호환성, 테스트, 특수 배포에 유용합니다. |

### 하드웨어 인코더

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `nvenc` | nvenc / NVIDIA GPU 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-ffnvcodec` | `<profile>-ffnvcodec-headers` | 지원되는 NVIDIA 시스템에서 H.264, HEVC, AV1을 빠르게 인코딩합니다. CPU 부담을 낮춘 고속 출력에 유용합니다. |
| `qsv` | qsv / Intel 하드웨어 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libvpl` | `<profile>-libvpl` | Intel Quick Sync 하드웨어 인코딩으로 낮은 CPU 사용량의 빠른 영상 출력을 제공합니다. |
| `amf` | amf / AMD GPU 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-amf` | `<profile>-amf-headers` | 지원되는 Radeon 시스템에서 H.264, HEVC, AV1을 빠르게 인코딩합니다. 속도와 낮은 CPU 사용량이 중요한 작업에 잘 맞습니다. |
| `libmfx` | libmfx / 레거시 Intel Media SDK (FFmpeg 7.0+에서 제거됨) | 외부 SDK/가져오기 경로 | 차단: 일반 준비 절차 없음 | LGPL 안전 경계 | `--enable-libmfx` | 일반 준비 절차 없음 | 오래된 Media SDK 디스패처를 통한 레거시 Intel 하드웨어 인코드/디코드입니다. FFmpeg가 7.0에서 지원을 제거하여 이 빌더가 대상으로 하는 이후 소스에서는 작동하지 않습니다. 대신 Intel QSV(oneVPL)를 사용하세요. |

### 비디오 디코더

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `dav1d` | dav1d / AV1 디코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libdav1d` | `<profile>-dav1d` | AV1 영상을 빠르게 디코딩합니다. AV1 파일의 재생, 검사, 변환을 부드럽게 처리할 때 유용합니다. |
| `xevd` | xevd / EVC 디코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libxevd` | `<profile>-xevd` | EVC 영상 스트림을 디코딩합니다. 새로운 Essential Video Coding 자료를 읽을 때 유용합니다. |
| `xevdb` | xevdb / EVC 기본 프로파일 디코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libxevdb` | `<profile>-xevd` | EVC 기본 프로파일로 MPEG-5 EVC 디코딩을 활성화합니다. 메인 xevd 행과 동일한 XEVD 패키지를 사용합니다. |
| `davs2` | libdavs2 / AVS2 디코딩 | 내부 소스 준비 경로 | 일반 선택 가능: 소스 준비/가져오기 | GPL 경계 | `--enable-libdavs2` | 고정된 소스/가져오기 절차로 준비 | AVS2 영상 스트림을 디코딩합니다. AVS2 자료가 포함된 지역 표준, 방송, 호환성 작업에 유용합니다. |
| `uavs3d` | libuavs3d / AVS3 디코딩 | 내부 소스 준비 경로 | 일반 선택 가능: 소스 준비/가져오기 | LGPL 안전 경계 | `--enable-libuavs3d` | 고정된 소스/가져오기 절차로 준비 | AVS3 영상 스트림을 디코딩합니다. AVS3 테스트 자료, 지역 표준, 호환성 작업에 유용합니다. |
| `lcevc-dec` | liblcevc-dec / LCEVC 디코딩 | 내부 소스 준비 경로 | 일반 선택 가능: 소스 준비/가져오기 | LGPL 안전 경계 | `--enable-liblcevc-dec` | 고정된 소스/가져오기 절차로 준비 | 기본 영상 스트림을 보강하는 LCEVC 향상 계층을 디코딩합니다. 계층형 영상 보강 콘텐츠에 유용합니다. |
| `avisynthplus` | AviSynth+ / 스크립트 기반 비디오 처리 | 내부 소스 준비 경로 | 일반 선택 가능: 소스 준비/가져오기 | GPL 경계 | `--enable-avisynth` | 고정된 소스/가져오기 절차로 준비 | AviSynth 스크립트 기반 영상 처리 체인을 직접 엽니다. 기존 작업이 AviSynth 필터나 스크립트 처리에 의존할 때 유용합니다. |

### 이미지 코덱

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `png` | png / PNG 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | 없음 | `<profile>-libpng` | 무손실 정지 이미지와 이미지 시퀀스에 쓰이는 PNG 입력과 출력을 처리합니다. |
| `webp` | webp / WebP 이미지 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libwebp` | `<profile>-libwebp` | 웹사이트에서 널리 쓰이는 WebP 이미지를 읽고 씁니다. 정지 이미지, 썸네일, 이미지 시퀀스에 유용합니다. |
| `openjpeg` | openjpeg / JPEG 2000 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libopenjpeg` | `<profile>-openjpeg2` | 아카이브, 디지털 시네마, 전문 이미징에서 쓰이는 JPEG 2000 이미지를 읽고 씁니다. |
| `libjxl` | libjxl / JPEG XL 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libjxl` | `<profile>-libjxl` | 현대적인 고품질 정지 이미지 압축 형식인 JPEG XL을 읽고 씁니다. 이미지 시퀀스와 고급 이미지 작업에 유용합니다. |
| `rsvg` | rsvg / SVG 렌더링 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-librsvg` | `<profile>-librsvg` | SVG 벡터 그래픽을 일반 이미지 프레임으로 렌더링합니다. 로고, 오버레이, 그래픽 자산에 유용합니다. |
| `snappy` | snappy / Snappy 압축 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libsnappy` | `<profile>-snappy` | 일부 포맷 내부에서 쓰이는 빠른 Snappy 데이터 압축을 제공합니다. |
| `lcms2` | lcms2 / 색상 관리 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-lcms2` | `<profile>-lcms2` | 색상 프로파일 관리를 적용해 더 정확한 색 변환을 수행합니다. 색 표현을 보존해야 할 때 유용합니다. |
| `svtjpegxs` | svtjpegxs / JPEG XS 인코딩 | MSYS2 패키지 경로 | UI 비활성화 | LGPL 안전 경계 | `--enable-libsvtjpegxs` | `<profile>-svt-jpeg-xs`, `git`, `<profile>-cmake`, `<profile>-ninja`, `<profile>-yasm` | 저지연 전문 미디어 작업을 위한 JPEG XS 영상을 만듭니다. |

### 필터와 처리

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `zimg` | zimg / 고품질 스케일링 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libzimg` | `<profile>-zimg` | 이미지 리사이즈, 픽셀 형식 변환, 색 처리를 더 깨끗하게 수행합니다. 정교한 영상 변환에 유용합니다. |
| `libplacebo` | libplacebo / GPU 비디오 처리 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libplacebo`, `--enable-vulkan` | `<profile>-libplacebo`, `<profile>-vulkan-loader`, `<profile>-vulkan-headers` | 스케일링, 색 변환, 톤 매핑, 렌더링 경로에 고품질 GPU 영상 처리를 제공합니다. |
| `vmaf` | vmaf / 비디오 품질 측정 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libvmaf` | `<profile>-vmaf` | 인코딩 결과를 기준 영상과 비교해 지각 품질을 측정합니다. 압축 설정을 평가할 때 유용합니다. |
| `vidstab` | vidstab / 비디오 흔들림 보정 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libvidstab` | `<profile>-vid.stab` | 카메라 흔들림을 줄여 손떨림이 있는 영상을 더 안정적으로 보이게 합니다. |
| `opencolorio` | opencolorio / 전문 색상 관리 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libopencolorio` | `<profile>-opencolorio` | 영화, 애니메이션, 후반작업에 맞는 전문 색상 관리를 적용합니다. |
| `cairo` | cairo / 2D 그래픽 렌더링 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-cairo` | `<profile>-cairo` | 생성 그래픽, 오버레이, 필터 시각 요소에 쓸 2D 벡터 드로잉을 제공합니다. 도형이나 그래픽 요소를 영상에 그릴 때 유용합니다. |
| `opencl` | opencl / OpenCL 처리 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-opencl` | `<profile>-opencl-icd`, `<profile>-opencl-headers` | OpenCL을 지원하는 GPU나 가속기에서 일부 compute 필터를 실행합니다. 무거운 이미지 처리를 분산할 때 유용합니다. |
| `shaderc` | shaderc / Vulkan 셰이더 컴파일러 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libshaderc` | `<profile>-shaderc` | GPU 기반 영상 필터에 쓰이는 셰이더 프로그램을 컴파일합니다. 고급 렌더링과 GPU 처리 경로에 유용합니다. |
| `glslang` | glslang / GLSL 셰이더 컴파일러 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libglslang` | `<profile>-glslang` | GPU 기반 영상 처리에 쓰이는 GLSL 셰이더 코드를 컴파일합니다. 고급 그래픽 및 컴퓨트형 필터 경로에 유용합니다. |
| `frei0r` | frei0r / 추가 비디오 효과 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | GPL 경계 | `--enable-frei0r` | `<profile>-frei0r-plugins` | 기본 필터 외의 창의적인 영상 효과와 플러그인을 제공합니다. 스타일화된 영상 처리에 유용합니다. |
| `opencv` | opencv / 컴퓨터 비전 필터 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libopencv` | `<profile>-opencv` | 이미지 분석과 실험적 시각 필터를 위한 컴퓨터 비전 처리를 제공합니다. |
| `ladspa` | ladspa / LADSPA 오디오 플러그인 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-ladspa` | `<profile>-ladspa-sdk`, `<profile>-dlfcn` | LADSPA 오디오 효과 플러그인을 불러와 추가 오디오 처리를 수행합니다. 기존 플러그인 체인을 미디어 변환에 활용할 때 유용합니다. |
| `lv2` | lv2 / LV2 오디오 플러그인 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-lv2` | `<profile>-lilv` | LV2 오디오 플러그인을 불러와 더 고급 오디오 효과와 처리 체인을 사용할 수 있게 합니다. |
| `qrencode` | qrencode / QR 코드 생성 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libqrencode` | `<profile>-qrencode` | 이미지나 영상 프레임에 넣을 수 있는 QR 코드를 생성합니다. |
| `cuda-nvcc` | cuda-nvcc / NVIDIA CUDA 필터 컴파일 | 외부 SDK/가져오기 경로 | 차단: 일반 준비 절차 없음 | LGPL 안전 경계 | `--enable-cuda-nvcc` | 일반 준비 절차 없음 | FFmpeg의 CUDA 가속 필터를 NVIDIA nvcc 컴파일러로 컴파일합니다. nvcc는 독점 CUDA Toolkit에만 포함되어 있으며 MSYS2 패키지로 제공되지 않습니다. |
| `lensfun` | lensfun / 렌즈 보정 | MSYS2 패키지 경로 | UI 비활성화 | LGPL 안전 경계 | `--enable-liblensfun` | `<profile>-lensfun` | 렌즈 왜곡, 비네팅, 카메라 렌즈 특유의 결함을 보정합니다. 알려진 렌즈로 촬영한 영상 정리에 유용합니다. |
| `vapoursynth` | VapourSynth / 스크립트 기반 비디오 처리 | MSYS2 패키지 경로 | UI 비활성화 | LGPL 안전 경계 | `--enable-vapoursynth` | `<profile>-vapoursynth` | VapourSynth 스크립트 기반 영상 처리 체인을 엽니다. 고급 스크립트 필터링 결과를 미디어 작업에 넣을 때 유용합니다. |

### 오디오

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `opus` | opus / Opus 오디오 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libopus` | `<profile>-opus` | 음성, 음악, 스트리밍, 저지연 통신에 적합한 현대적인 Opus 오디오를 만듭니다. |
| `fdk-aac` | fdk-aac / AAC 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | 비자유 경계 | `--enable-libfdk-aac`, `--enable-nonfree` | `<profile>-fdk-aac` | MP4, 모바일 기기, 스트리밍, 웹 재생에 적합한 고품질 AAC 오디오를 만듭니다. AAC 품질이 중요할 때 많이 쓰입니다. |
| `mp3lame` | mp3lame / MP3 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libmp3lame` | `<profile>-lame` | 거의 모든 장치와 플레이어에서 재생되는 MP3 오디오를 만듭니다. 호환성이 가장 중요할 때 유용합니다. |
| `vorbis` | vorbis / Vorbis 오디오 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libvorbis` | `<profile>-libvorbis` | Ogg 워크플로에 잘 맞는 Vorbis 오디오를 만듭니다. 음악과 일반 오디오에 쓰이는 개방형 형식입니다. |
| `soxr` | soxr / 고품질 오디오 리샘플링 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libsoxr` | `<profile>-libsoxr` | 고품질 오디오 샘플레이트 변환을 제공합니다. 오디오 rate를 바꾸면서 선명도를 유지할 때 유용합니다. |
| `rubberband` | rubberband / 템포 및 피치 변경 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | GPL 경계 | `--enable-librubberband` | `<profile>-rubberband` | 오디오 속도와 피치를 더 좋은 품질로 바꿉니다. 음악, 대사 타이밍, 음정 조정에 유용합니다. |
| `chromaprint` | chromaprint / 오디오 지문 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-chromaprint` | `<profile>-chromaprint` | 트랙 식별과 매칭에 쓰이는 작은 오디오 지문을 만듭니다. 파일명이나 메타데이터가 없어도 오디오를 알아볼 때 유용합니다. |
| `twolame` | twolame / MP2 오디오 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libtwolame` | `<profile>-twolame` | 방송, DVD, 오래된 미디어 작업에 쓰이는 MP2 오디오를 만듭니다. |
| `speex` | speex / Speex 음성 오디오 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libspeex` | `<profile>-speex` | 오래된 Speex 음성 오디오를 처리합니다. 레거시 음성 녹음과 통신 자료에 유용합니다. |
| `opencore-amr` | opencore-amr / AMR 오디오 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libopencore-amrnb`, `--enable-libopencore-amrwb` | `<profile>-opencore-amr` | 오래된 모바일 및 음성 시스템에서 쓰인 AMR 음성 오디오를 읽고 씁니다. |
| `vo-amrwbenc` | vo-amrwbenc / AMR-WB 음성 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libvo-amrwbenc` | `<profile>-vo-amrwbenc` | 모바일식 음성 호환성을 위한 AMR-WB 광대역 음성 오디오를 만듭니다. |
| `gsm` | gsm / GSM 오디오 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libgsm` | `<profile>-gsm` | 오래된 GSM 음성 오디오를 처리합니다. 과거 통화 녹음, 전화 음성 자료, 호환성 작업에 유용합니다. |
| `lc3` | lc3 / LC3 오디오 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-liblc3` | `<profile>-liblc3` | 현대적인 저복잡도 통신 오디오에 쓰이는 LC3를 처리합니다. 블루투스 및 음성 중심 호환성 작업에 유용합니다. |
| `ilbc` | ilbc / iLBC 음성 오디오 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libilbc` | `<profile>-libilbc` | 오래된 인터넷 음성 통신에서 쓰인 iLBC 음성 오디오를 지원합니다. 음성 통화 녹음과 호환성 작업에 유용합니다. |
| `whisper` | whisper / 음성 인식 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-whisper` | `<profile>-whisper.cpp`, `<profile>-ggml` | whisper.cpp 음성 인식으로 말소리를 텍스트로 변환합니다. 전사와 자막 생성에 유용합니다. |
| `mysofa` | mysofa / 공간 음향 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libmysofa` | `<profile>-libmysofa` | 헤드폰 기반 3D 오디오 렌더링에 필요한 공간 오디오 필터 데이터를 제공합니다. |
| `bs2b` | bs2b / 헤드폰 크로스피드 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libbs2b` | `<profile>-libbs2b` | 헤드폰용 크로스피드를 적용해 좌우가 과하게 분리된 스테레오를 더 스피커처럼 들리게 만듭니다. 헤드폰 청취용 오디오 처리에 유용합니다. |
| `gme` | gme / 게임 음악 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libgme` | `<profile>-libgme` | 고전 콘솔과 시스템의 게임 음악 형식을 읽습니다. 칩튠 계열 소스의 재생이나 변환에 유용합니다. |
| `shine` | shine / 고정소수점 MP3 인코딩 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libshine` | `<profile>-shine` | 간단한 fixed-point MP3 인코더로 MP3 오디오를 만듭니다. 제약이 있거나 예측 가능한 동작이 필요한 경우에 가깝습니다. |
| `codec2` | codec2 / 저비트레이트 음성 오디오 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libcodec2` | `<profile>-codec2` | 무전, 음성 통신, 실험용 작업에 맞는 초저비트레이트 음성 코딩을 지원합니다. 아주 작은 용량으로 말소리를 전달할 때 유용합니다. |
| `mpeghdec` | libmpeghdec / MPEG-H 오디오 디코딩 | 내부 소스 준비 경로 | 일반 선택 가능: 소스 준비/가져오기 | 비자유 경계 | `--enable-libmpeghdec`, `--enable-nonfree` | 고정된 소스/가져오기 절차로 준비 | 몰입형 및 객체 기반 오디오 콘텐츠를 위한 MPEG-H 3D Audio를 디코딩합니다. |
| `pocketsphinx` | pocketsphinx / 음성 인식 | 내부 소스 준비 경로 | 차단: 일반 준비 절차 없음 | LGPL 안전 경계 | `--enable-pocketsphinx` | 일반 준비 절차 없음 | CMU PocketSphinx를 사용하는 asr 음성 인식 오디오 필터를 추가합니다. 현재는 빌드할 수 없습니다. FFmpeg의 asr 필터가 최신 PocketSphinx 릴리스와 호환되지 않으므로, 선택하면 빌드가 차단됩니다. |

### 자막과 텍스트

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `ass` | ass / 스타일 자막 렌더링 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libass` | `<profile>-libass` | 글꼴, 색상, 위치, 외곽선, 애니메이션식 효과가 들어간 ASS/SSA 자막을 렌더링합니다. 복잡한 자막 표현에 자주 쓰입니다. |
| `freetype` | freetype / 글꼴 렌더링 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libfreetype` | `<profile>-freetype` | 자막, 오버레이, 텍스트 필터에 사용할 글꼴 글리프를 렌더링합니다. 글자를 실제 그래픽으로 그려야 할 때 중요합니다. |
| `fontconfig` | fontconfig / 글꼴 검색 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libfontconfig` | `<profile>-fontconfig` | 설치된 글꼴을 찾아 자막과 텍스트 렌더링이 알맞은 서체를 쓰도록 합니다. 글자 모양의 안정성이 필요할 때 유용합니다. |
| `harfbuzz` | harfbuzz / 복잡한 텍스트 셰이핑 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libharfbuzz` | `<profile>-harfbuzz` | 복잡한 문자가 올바르게 결합, 재배치, 위치 조정되도록 글자를 shaping합니다. 여러 비라틴 문자 자막에 중요합니다. |
| `fribidi` | fribidi / 오른쪽에서 왼쪽 텍스트 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libfribidi` | `<profile>-fribidi` | 아랍어, 히브리어처럼 오른쪽에서 왼쪽으로 쓰는 문자의 방향 처리를 지원합니다. 양방향 텍스트가 있는 자막이나 오버레이에 유용합니다. |
| `aribcaption` | aribcaption / ARIB 캡션 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libaribcaption` | `<profile>-libaribcaption` | 일본 방송 캡션을 현대적인 ARIB 처리 방식으로 읽습니다. TV 녹화 파일이나 방송용 전송 스트림에 유용합니다. |
| `aribb24` | aribb24 / ARIB B24 캡션 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libaribb24` | `<profile>-aribb24` | 일본 방송 자료에 쓰이는 ARIB B24 캡션을 처리합니다. 오래된 방송 자막 데이터를 보존하거나 변환할 때 유용합니다. |
| `zvbi` | zvbi / 텔레텍스트 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | GPL 경계 | `--enable-libzvbi` | `<profile>-zvbi` | 오래된 방송 소스에서 teletext와 VBI 데이터를 추출합니다. 숨은 자막, 페이지, 방송 문자 정보에 유용합니다. |

### 디스크 및 장치 입력

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `bluray` | bluray / Blu-ray 입력 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libbluray` | `<profile>-libbluray` | Blu-ray 디스크 구조를 읽어 검사, 변환, 추출 작업에 사용합니다. 단일 파일이 아니라 Blu-ray 폴더 구조가 소스일 때 유용합니다. |
| `dvdread` | dvdread / DVD 입력 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | GPL 경계 | `--enable-libdvdread` | `<profile>-libdvdread` | DVD 미디어 구조를 읽어 추출, 검사, 변환 작업에 사용합니다. DVD 폴더나 디스크 레이아웃이 소스일 때 유용합니다. |
| `dvdnav` | libdvdnav / DVD 내비게이션 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | GPL 경계 | `--enable-libdvdnav` | `<profile>-libdvdnav`, `<profile>-libdvdread` | DVD-Video의 메뉴, 타이틀, 챕터, 프로그램 체인 같은 내비게이션 구조를 읽습니다. 디스크 방식의 DVD 작업에 유용합니다. |
| `openmpt` | openmpt / 모듈 음악 입력 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libopenmpt` | `<profile>-libopenmpt` | tracker module 음악을 비교적 정확한 재생 동작으로 읽습니다. 오래된 게임, 데모신, 트래커 음악 형식에 유용합니다. |
| `sdl2` | sdl2 / SDL2 출력 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-sdl2` | `<profile>-SDL2` | SDL2를 통해 간단한 미디어 출력과 미리보기 기능을 제공합니다. 재생식 테스트와 표시 작업에 유용합니다. |
| `openal` | openal / 오디오 장치 입력 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-openal` | `<profile>-openal` | 라이브 오디오 장치를 다루기 위한 추가 입력 및 출력 경로를 제공합니다. |
| `cdio` | cdio / CD 입력 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | GPL 경계 | `--enable-libcdio` | `<profile>-libcdio`, `<profile>-libcdio-paranoia` | CD 입력을 읽어 오디오 추출이나 미디어 검사 작업에 사용합니다. 복사된 파일이 아니라 실제 CD가 소스일 때 유용합니다. |
| `modplug` | modplug / 모듈 음악 입력 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libmodplug` | `<profile>-libmodplug` | 오래된 tracker module 음악 형식을 읽습니다. 레거시 장면 음악이나 게임풍 음악 파일 변환에 유용합니다. |
| `jack` | jack / JACK 오디오 입력 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libjack` | `<profile>-jack2` | 스튜디오식 오디오 라우팅을 위해 JACK 오디오에 연결합니다. 전문 Linux 오디오 환경에서 캡처와 재생에 유용합니다. |
| `pulse` | pulse / PulseAudio 입력 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libpulse` | `<profile>-pulseaudio` | Linux 데스크톱 오디오 캡처나 재생을 위해 PulseAudio에 연결합니다. |
| `caca` | caca / 텍스트 모드 비디오 출력 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libcaca` | `<profile>-libcaca` | 영상을 색이 있는 텍스트 화면처럼 변환합니다. 실험, 터미널 표시, 특수한 미리보기 효과에 가깝습니다. |
| `opengl` | opengl / OpenGL 출력 장치 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-opengl` | `<profile>-mesa` | OpenGL을 통한 하드웨어 가속 비디오 표시를 위한 OpenGL 출력 장치를 추가합니다. Mesa 패키지가 제공합니다. |
| `dc1394` | dc1394 / IEEE 1394 카메라 캡처 | 내부 소스 준비 경로 | 차단: 일반 준비 절차 없음 | LGPL 안전 경계 | `--enable-libdc1394` | 일반 준비 절차 없음 | IEEE 1394(FireWire) 카메라에서 비디오를 캡처합니다. Windows에서는 차단됩니다: MSYS2 패키지가 없으며, 유일한 Windows 빌드 방법은 독점 FireWire 커널 드라이버와 FireWire 하드웨어를 필요로 하므로 결과물이 이식 가능하지 않습니다. |
| `decklink` | decklink / Blackmagic 캡처 및 재생 | 외부 SDK/가져오기 경로 | 차단: 일반 준비 절차 없음 | LGPL 안전 경계 | `--enable-decklink` | 일반 준비 절차 없음 | Blackmagic DeckLink 캡처 및 재생 지원을 추가합니다. 독점 DeckLink SDK 헤더를 대상으로 빌드되며, 이 헤더는 MSYS2 패키지로 재배포할 수 없습니다. |

### 네트워크

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `srt` | srt / 안정적인 라이브 스트리밍 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libsrt` | `<profile>-srt` | 불안정한 네트워크에서도 라이브 전송을 안정적으로 유지하는 SRT 스트리밍을 제공합니다. |
| `rtmp` | rtmp / RTMP 스트리밍 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-librtmp` | `<profile>-rtmpdump` | 오래된 라이브 스트리밍 서버와 워크플로에 필요한 RTMP 호환성을 제공합니다. |
| `rist` | rist / 안정적인 스트림 전송 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-librist` | `<profile>-librist` | 불안정한 네트워크에서 전문 라이브 스트리밍을 안정화하는 RIST 전송을 제공합니다. |
| `ssh` | ssh / SSH 파일 접근 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libssh` | `<profile>-libssh` | SSH와 SFTP 계열 원격 접근으로 미디어를 읽고 씁니다. |
| `zmq` | zmq / ZeroMQ 제어 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libzmq` | `<profile>-zeromq` | 지원되는 처리 흐름에서 메시지 기반 런타임 제어를 가능하게 합니다. 자동화와 상호작용식 제어에 유용합니다. |
| `rabbitmq` | rabbitmq / RabbitMQ 메시징 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-librabbitmq` | `<profile>-rabbitmq-c` | RabbitMQ 메시지 큐와 미디어 처리 흐름을 연결합니다. 자동화된 처리 시스템에 유용합니다. |
| `smbclient` | libsmbclient / SMB 네트워크 파일 접근 | 외부 SDK/가져오기 경로 | 차단: 일반 준비 절차 없음 | GPL 경계 | `--enable-libsmbclient` | 일반 준비 절차 없음 | SMB/CIFS 네트워크 공유에서 미디어를 읽고 쓸 수 있게 합니다. Windows식 네트워크 폴더에 저장된 미디어에 유용합니다. 아직 Windows에서 빌드할 수 없어 이 섹션 맨 아래에 표시되며, Windows용 libsmbclient가 나올 때까지 사용할 수 없습니다. |

### 보안 네트워크(TLS)

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `openssl` | openssl / 보안 네트워크 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | 비자유 경계 | `--enable-openssl` | `<profile>-openssl` | HTTPS와 다른 보안 미디어 프로토콜에 필요한 암호화 네트워크 연결을 제공합니다. |
| `gnutls` | gnutls / 보안 네트워크 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-gnutls` | `<profile>-gnutls` | HTTPS와 TLS 기반 미디어 접근에 필요한 암호화 네트워크 연결을 제공합니다. 안전한 스트리밍과 원격 소스에 유용합니다. |
| `mbedtls` | mbedTLS / 보안 네트워크 접근 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-mbedtls` | `<profile>-mbedtls` | 가벼운 TLS 기능으로 암호화된 네트워크 미디어 접근을 제공합니다. 작은 보안 백엔드가 필요할 때 유용합니다. |
| `libtls` | libtls / 보안 네트워크 접근 | 내부 소스 준비 경로 | 일반 선택 가능: 소스 준비/가져오기 | LGPL 안전 경계 | `--enable-libtls` | 고정된 소스/가져오기 절차로 준비 | 간결한 TLS 인터페이스로 암호화된 네트워크 통신을 제공합니다. 보안이 필요한 네트워크 미디어 접근에 유용합니다. |

### OCR

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `tesseract` | tesseract / OCR 텍스트 인식 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libtesseract` | `<profile>-tesseract-ocr` | 이미지나 영상 프레임에 보이는 글자를 읽습니다. 화면에 박힌 제목, 표지판, 자막, 문서 글자 추출에 유용합니다. |

### AI 지원

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `onnxruntime` | ONNX Runtime / AI 모델 추론 | MSYS2 패키지 경로 | UI 비활성화; mingw64 프로필 제외 | LGPL 안전 경계 | `--enable-libonnxruntime` | `<profile>-onnxruntime` | ONNX Runtime으로 지원되는 딥러닝 필터를 실행합니다. 모델 기반 분석이나 향상 작업에 유용합니다. |
| `openvino` | OpenVINO / AI 모델 추론 | 외부 SDK/가져오기 경로 | 차단: 일반 준비 절차 없음 | LGPL 안전 경계 | `--enable-libopenvino` | 일반 준비 절차 없음 | Intel 계열 가속에 맞춘 AI 추론 필터를 실행합니다. 모델 기반 영상 또는 이미지 처리에 유용합니다. |
| `torch` | Torch / libtorch | 외부 SDK/가져오기 경로 | 차단: 일반 준비 절차 없음 | LGPL 안전 경계 | `--enable-libtorch` | 일반 준비 절차 없음 | Torch 기반 모델 실행으로 지원되는 딥러닝 필터를 돌립니다. PyTorch 계열 추론 작업에 유용합니다. |
| `tensorflow` | TensorFlow / AI 모델 추론 | 외부 SDK/가져오기 경로 | UI 비활성화 | LGPL 안전 경계 | `--enable-libtensorflow` | 고정된 소스/가져오기 절차로 준비 | TensorFlow C API로 지원되는 딥러닝 필터를 실행합니다. 모델 기반 이미지나 영상 분석에 유용합니다. |

### 지원 라이브러리

| ID | 표시 이름 | 경로 | 일반 상태 | 라이선스 경계 | configure 플래그 | 패키지 / 준비 | 용도 |
|---|---|---|---|---|---|---|---|
| `xml2` | xml2 / XML 지원 | MSYS2 패키지 경로 | 일반 선택 가능: MSYS2 패키지 | LGPL 안전 경계 | `--enable-libxml2` | `<profile>-libxml2` | 일부 미디어 형식, 자막, manifest, 메타데이터 작업에 쓰이는 XML 구조 데이터를 읽습니다. |
| `quirc` | libquirc / QR 코드 디코딩 | 내부 소스 준비 경로 | 일반 선택 가능: 소스 준비/가져오기 | LGPL 안전 경계 | `--enable-libquirc` | 고정된 소스/가져오기 절차로 준비 | 영상 프레임이나 이미지에서 QR 코드를 읽습니다. 자동화, 스캔, 시각 메타데이터 작업에 유용합니다. |
| `klvanc` | libklvanc / 방송 메타데이터 | 내부 소스 준비 경로 | 일반 선택 가능: 소스 준비/가져오기 | LGPL 안전 경계 | `--enable-libklvanc` | 고정된 소스/가져오기 절차로 준비 | 방송 영상의 수직 보조 데이터(VANC)를 처리합니다. 영상 라인에 함께 실리는 메타데이터 작업에 유용합니다. |

## 라이선스 제한

이 프로그램은 사용자가 `--enable-gpl`, `--enable-nonfree`, `--enable-version3`를 직접 고르는 방식이 아니라 선택된 라이브러리와 최종 configure 플래그를 보고 빌드 계획의 라이선스 경계를 계산합니다. GPL 항목이 들어가면 GPL 로컬 경계로 이동하고, 비자유 항목이 들어가면 비자유 로컬 경계와 재배포 경고가 따라붙습니다. version 3 조건이 필요한 항목은 필요한 경우 `--enable-version3`가 자동으로 반영됩니다.

| 라이선스 경계 | 행 수 | 대표 항목 |
|---|---:|---|
| 기본 포함 | 10 | `ffmpeg-program`, `ffprobe-program`, `libavutil`, `libavcodec`, `libavformat`, `libavfilter`, `libswscale`, `libswresample`, `native-codecs`, `native-formats` |
| LGPL 안전 경계 | 98 | `svt-av1`, `libvpx`, `aom`, `openh264`, `rav1e`, `theora`, `kvazaar`, `xeve`, `xeveb`, `oapv`, `vvenc`, `nvenc` 등 |
| GPL 경계 | 14 | `x264`, `x265`, `xvid`, `xavs`, `xavs2`, `davs2`, `avisynthplus`, `frei0r`, `rubberband`, `zvbi`, `dvdread`, `dvdnav` 등 |
| 비자유 경계 | 3 | `fdk-aac`, `mpeghdec`, `openssl` |
