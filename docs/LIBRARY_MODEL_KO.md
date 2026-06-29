# 라이브러리 선택과 빌드 옵션

이 문서는 Promptful Custom FFmpeg Builder가 라이브러리 선택, 프리셋, configure 옵션, 충돌 항목, 라이선스 영향을 처리하는 방식을 설명합니다.

화면에 보이는 라이브러리 항목은 각각 취급과 처리 방법이 다를 수 있습니다. 어떤 항목은 FFmpeg 소스에 기본으로 포함되어 있고, 어떤 항목은 MSYS2 패키지로 설치되며, 어떤 항목은 프로그램이 별도 소스 준비 절차를 거쳐야 합니다. 또한 프로그램이 항목을 알고 있더라도 현재 선택 대상으로 다루지 않는 경우도 있습니다.

## 라이브러리 항목의 종류

| 종류 | 설명 | 일반적인 빌드 영향 |
|---|---|---|
| FFmpeg 기본 포함 항목 | FFmpeg 소스 트리에 원래 들어 있는 프로그램과 핵심 라이브러리 | 항상 켜져 있습니다. 추가 MSYS2 패키지나 `--enable-lib...` 플래그를 만들지 않습니다. |
| MSYS2 패키지 항목 | MSYS2 패키지로 의존성을 충족하는 선택형 FFmpeg 연동 항목 | 패키지 이름과 FFmpeg configure 플래그를 추가합니다. |
| 내부 소스 준비 항목 | 개인용 MSYS2 환경 안에서 프로그램이 직접 준비하는 항목 | 준비 레시피를 실행한 뒤 configure 플래그를 추가합니다. |
| 외부 SDK/가져오기 항목 | 외부 SDK, import 라이브러리, 별도 준비 경로가 필요한 항목 | 안전한 준비/가져오기 경로가 없으면 일반 계획에 포함하지 않습니다. |
| 비활성화 또는 조건부 항목 | 일반 조건에서는 선택할 수 없거나, 버전/프로필에 따라 달라지는 항목 | 사용할 수 없음, 차단, 버전/프로필 제외로 표시됩니다. |

## FFmpeg 기본 포함 항목

FFmpeg 기본 포함 항목은 FFmpeg 소스 자체에서 빌드되므로 항상 선택된 것으로 봅니다. 예를 들면 다음과 같습니다.

- `ffmpeg`
- `ffprobe`
- `avcodec`
- `avformat`
- `avfilter`
- `avdevice`
- `swscale`
- `swresample`

이 항목들은 모든 빌드의 기준입니다. 외부 패키지를 설치하지 않고, 별도의 `--enable-lib...` 플래그도 추가하지 않습니다.

## MSYS2 패키지 항목

대부분의 선택형 라이브러리는 MSYS2 패키지 경로를 사용합니다. 이런 항목을 선택하면 프로그램은 필요한 MSYS2 패키지 이름과 그에 맞는 FFmpeg configure 플래그를 계획에 넣습니다.

예를 들면 `x264`, `x265`, `libvpx`, `aom`, `dav1d`, `opus`, `mp3lame`, `ass`, `freetype`, `zimg`, `vmaf`, `srt`, `openssl` 같은 항목이 있습니다.

다만 MSYS2 패키지 항목이라고 해서 모든 FFmpeg 버전과 모든 MSYS2 셸 프로필에서 항상 가능한 것은 아닙니다. 프로그램은 계획을 만들기 전에 선택한 FFmpeg 버전과 프로필에 맞지 않는 조합을 걸러냅니다.

## 내부 소스 준비 항목

일부 라이브러리는 패키지 하나를 설치하는 것만으로는 준비가 끝나지 않습니다. 프로그램이 지원하는 내부 소스 준비 항목은 개인용 빌드 환경 안에서 소스나 import 파일을 준비한 뒤 FFmpeg configure로 넘깁니다.

현재 구현된 내부 소스 준비 항목은 다음과 같습니다.

- `vvenc`
- `lcevc-dec`
- `davs2`
- `uavs3d`
- `xavs2`
- `avisynthplus`
- `mpeghdec`
- `quirc`
- `klvanc`
- `libtls`
- `libmfx`

`libmfx`도 이 구현된 그룹에 속합니다. `libmfx`는 선택한 FFmpeg 버전과 프로필에서 레거시 Intel Media SDK 경로가 필요할 때 쓰는 백엔드입니다. 최신 oneVPL 경로인 `libvpl`과 동시에 쓰지 않습니다. `dc1394`나 `pocketsphinx`처럼 내부 트랙에 속하더라도 안전한 일반 준비 절차가 없으면 여전히 차단될 수 있습니다.

## 외부 또는 차단 항목

일부 FFmpeg 연동은 외부 SDK, 특수 import 라이브러리, 별도 빌드 경로가 필요합니다. 이 프로그램이 그 경로를 안전하게 준비하지 못하는 경우에는 카탈로그에 이름이 있더라도 일반 계획에 넣지 않습니다.

현재 일반 빌드/가져오기 절차가 없어 차단되는 항목은 다음과 같습니다.

- `smbclient`
- `openvino`
- `torch`
- `pocketsphinx`
- `dc1394`
- `decklink`
- `cuda-nvcc`

이는 FFmpeg가 이러한 기술을 사용할 수 없다는 설명이 아닙니다. 이 프로그램이 현재 안전하고, 로컬에서 검토 가능하며, 반복 가능한 준비 경로를 제공하지 않는다는 설명입니다.

## 비활성화와 버전/프로필 제한

일반 UI에서 전역으로 비활성화된 항목은 두 개입니다.

- `tensorflow`
- `vapoursynth`

다른 항목들은 전역 비활성화가 아니라 선택한 FFmpeg 버전, MSYS2 셸 프로필, 패키지/API 조건에 따라 숨겨지거나 차단되거나 최종 계획에서 빠질 수 있습니다.

예를 들면 다음과 같습니다.

- `libvpl`은 최신 FFmpeg 계열에서 쓰는 Intel oneVPL 경로입니다.
- `libmfx`는 레거시 Intel Media SDK 경로이며 `libvpl`과 함께 켤 수 없습니다.
- `lensfun`은 사용 가능한 패키지가 선택한 FFmpeg 버전의 요구 API를 만족하지 못하면 차단됩니다.
- `onnxruntime`은 현재 적용한 공식 FFmpeg 릴리스 범위에서 일반 지원 항목으로 다루지 않으며, `mingw64`에서도 사용할 수 없습니다.
- `svtjpegxs`는 FFmpeg 버전 요구 사항과 pkg-config 결과에 따라 달라집니다.

## 공개 라이브러리 프리셋

라이브러리 프리셋은 선택을 시작하기 위한 기준입니다. 프리셋이 카탈로그를 대신하지 않으며, 선택한 FFmpeg 버전과 프로필에서 모든 항목이 그대로 유지된다고 보장하지도 않습니다.

공개 프리셋은 다음과 같습니다.

| 프리셋 | 목적 | 선택 규칙 |
|---|---|---|
| `minimal` | FFmpeg 소스에 기본 포함된 가장 작은 기준 | 기본 포함 항목만 선택합니다. |
| `default` | 일반 코덱, 오디오 도구, 자막/글꼴, 하드웨어 가속, 기본 네트워크/필터 보조 항목을 넣은 실용 기준 | `minimal` + 기본 추가 항목 |
| `efficiency` | 압축 효율과 품질 대비 용량에 초점을 둔 선택 | `default` + 효율 추가 항목만 |
| `compatibility` | 더 넓은 코덱, 자막, 캡션, 이미지, 음성, 프로토콜 호환성 | `default` + 호환성 추가 항목만 |
| `editor` | 편집, 필터, 색 관리, 오디오 분석, 자막, 전사, 이미지 작업 흐름 | `default` + 편집 추가 항목만 |
| `full` | 상호 배타 항목을 정리한 뒤 가능한 가장 넓은 공개 선택 | `default` + 효율 + 호환성 + 편집 + 전체 전용 추가 항목 |
| `custom` | 현재 선택이 어떤 프리셋과도 정확히 일치하지 않을 때 표시 | 적용되는 프리셋 템플릿이 아닙니다. |

`efficiency`, `compatibility`, `editor`는 서로를 차례로 상속하는 단계가 아닙니다. 각각 `default`에서 출발해 자기 목적에 맞는 항목만 더합니다. `full`은 공개 프리셋의 넓은 합집합입니다.

## 공개 프리셋의 추가 항목

| 프리셋 | 기준에서 추가되는 항목 |
|---|---|
| `default` | `nvenc`, `amf`, `libvpl`, `libmfx`, `x264`, `x265`, `libvpx`, `aom`, `svt-av1`, `dav1d`, `theora`, `xvid`, `opus`, `vorbis`, `mp3lame`, `gsm`, `speex`, `opencore-amr`, `vo-amrwbenc`, `rubberband`, `openjpeg`, `webp`, `freetype`, `fontconfig`, `fribidi`, `harfbuzz`, `ass`, `cairo`, `zimg`, `vmaf`, `vidstab`, `srt`, `ssh`, `zmq`, `openal`, `sdl2`, `gme`, `openmpt` |
| `efficiency` | `fdk-aac`, `soxr`, `rav1e` |
| `compatibility` | `openh264`, `xeve`, `xevd`, `oapv`, `xavs`, `ilbc`, `twolame`, `shine`, `codec2`, `lc3`, `snappy`, `rsvg`, `zvbi`, `aribb24`, `aribcaption`, `rtmp` |
| `editor` | `png`, `libjxl`, `lcms2`, `libplacebo`, `shaderc`, `frei0r`, `opencv`, `opencolorio`, `xml2`, `mysofa`, `bs2b`, `ladspa`, `lv2`, `chromaprint`, `qrencode`, `whisper` |
| `full` | 효율, 호환성, 편집 추가 항목 전체와 `kvazaar`, `bluray`, `dvdread`, `dvdnav`, `cdio`, `modplug`, `opengl`, `openssl`, `rist`, `rabbitmq`, `tesseract`, `jack`, `pulse`, `caca`, `opencl` |

`libvpl`과 `libmfx`는 Intel 하드웨어 가속을 서로 다른 FFmpeg 계열에서 처리하기 위해 기본 프리셋에 함께 들어 있습니다. 실제 최종 계획에서는 선택한 버전과 프로필에 맞는 백엔드만 남기며, 둘을 동시에 FFmpeg configure로 넘기지 않습니다.

## 확장 라이브러리 모드

Extended 토글은 별도 프리셋이 아닙니다. 더 넓은 공개 프리셋에 일부 소스 준비 항목을 추가하는 장치입니다. `minimal`과 `default`에는 의도적으로 영향을 주지 않습니다.

| 프리셋 | Extended에서 추가되는 항목 |
|---|---|
| `efficiency` | `vvenc`, `lcevc-dec` |
| `compatibility` | `davs2`, `uavs3d`, `xavs2`, `avisynthplus`, `klvanc` |
| `editor` | `avisynthplus`, `lcevc-dec`, `quirc` |
| `full` | `vvenc`, `lcevc-dec`, `davs2`, `uavs3d`, `xavs2`, `avisynthplus`, `mpeghdec`, `quirc`, `klvanc` |

Extended 항목은 최종 라이선스를 바꿀 수 있습니다. 예를 들어 `xavs2`, `davs2`, `avisynthplus`는 GPL 영향을 만들고, `mpeghdec`는 nonfree 영향을 만듭니다.

## 함께 켤 수 없는 선택

일부 항목은 FFmpeg가 조합을 거부하거나, 같은 역할의 다른 연결 방식이기 때문에 함께 켤 수 없습니다.

| 그룹 | 항목 | 규칙 |
|---|---|---|
| TLS 백엔드 | `openssl`, `gnutls`, `mbedtls`, `libtls` | 하나만 선택합니다. |
| 런타임 셰이더 컴파일러 | `shaderc`, `glslang` | 최대 하나만 선택합니다. |
| EVC 디코더 바인딩 | `xevd`, `xevdb` | 프로필 바인딩 하나만 선택합니다. |
| EVC 인코더 바인딩 | `xeve`, `xeveb` | 프로필 바인딩 하나만 선택합니다. |

UI는 사용자가 항목을 켜고 끌 때 충돌 항목을 정리합니다. 상호 배타 플래그가 최종 계획까지 도달하면 플래너가 다시 검사하고 차단합니다.

## 수동 configure 플래그

고급 configure 플래그 입력칸은 예외적 수동 입력을 위한 영역이며, 라이브러리 처리 규칙을 우회하는 통로가 아닙니다.

사용자가 직접 넣은 플래그가 카탈로그에 있는 라이브러리 항목의 플래그와 정확히 일치하면, 플래너는 그 항목이 사실상 선택된 것으로 봅니다. 이 처리를 통해 필요한 패키지, 라이선스 영향, 준비 절차가 함께 반영됩니다.

이 복구는 configure 플래그를 가진 카탈로그 항목에만 적용됩니다. 패키지만 추가하고 configure 플래그가 없는 항목은 수동 `--enable-lib...` 플래그만으로 복구할 수 없습니다.

## 빌드 옵션

Options 페이지는 라이브러리 항목이 아닌 FFmpeg 빌드 스위치를 다룹니다.

고정 기본값은 다음과 같습니다.

- `default-static`
- `default-programs`
- `default-ffmpeg`
- `default-ffprobe`

선택 가능한 옵션에는 다음이 있습니다.

- 출력 형식: `enable-shared`
- 프로그램: `disable-ffplay`
- 보안/재현성: `disable-autodetect`, `disable-network`
- 호환성: `disable-asm`, `disable-x86asm`, `pkg-config-static`, `enable-runtime-cpudetect`
- 크기/속도: `disable-doc`, `enable-small`, `enable-lto`
- 디버깅: `disable-debug`, `disable-stripping`

Options 페이지는 `--enable-gpl`, `--enable-nonfree`, `--enable-version3`를 일반 체크박스로 제공하지 않습니다. 이 플래그들은 선택된 라이브러리와 최종 플래그에서 계산됩니다. 사용자가 직접 입력하더라도 백엔드는 이를 다시 검증하고 맞는 라이선스를 계산합니다.

또한 `disable-programs`와 `disable-ffprobe`는 일반 옵션으로 제공하지 않습니다. 고정된 프로그램 기본값과 충돌하기 때문입니다.

## 옵션 위험도

configure 옵션의 위험도는 라이선스가 아니라 빌드 형태, 실행 안정성, 성능, 휴대성에서 예상 밖의 변화가 발생할 가능성을 기준으로 합니다.

| 위험도 | 설명 |
|---|---|
| 높음 | 빌드 형태, 휴대성, 성능을 크게 바꿀 수 있습니다. 예: `enable-shared`, `disable-asm` |
| 보통 | 특정 상황에서 기능이나 성능이 줄어들 수 있습니다. 예: `disable-network`, `disable-x86asm`, `enable-lto` |
| 낮음 | 일반 기본값이거나 비교적 안전한 선택입니다. |

`enable-shared`가 높은 위험도로 분류되는 이유는 빌드 결과를 DLL 중심으로 바꾸어 이 빌더의 정적 링크 기대와 달라질 수 있기 때문입니다. `disable-asm`은 대부분의 SIMD 최적화를 제거하므로 높은 위험도로 분류됩니다.

## 옵션 프리셋

옵션 프리셋은 목적별 선택 묶음입니다. 서로 상속되는 단계가 아닙니다.

현재 옵션 프리셋은 다음과 같습니다.

- `none`
- `standard`
- `compact`
- `portable`
- `performance`
- `custom`

`standard`는 실용적인 초기 기본값입니다. `pkg-config-static`과 `disable-doc`을 선택합니다.

`enable-shared`, `disable-asm`, `disable-x86asm`, `disable-network`처럼 위험도가 높거나 문제 해결용 성격이 강한 옵션은 일반 프리셋에 의도적으로 넣지 않습니다.

`disable-network`를 SRT, libssh, librtmp, librist, ZeroMQ, RabbitMQ-C 같은 네트워크 라이브러리와 함께 선택하면 플래너가 경고를 냅니다. 네트워크를 꺼 둔 FFmpeg 빌드에서는 이런 연동이 실질적으로 유용하지 않기 때문입니다.

## 라이선스 프로필

프로그램은 선택된 라이브러리와 최종 configure 플래그에서 로컬 라이선스 프로필을 계산합니다.

현재 로컬 프로필 이름은 다음과 같습니다.

- `lgpl-local`
- `gpl-local`
- `nonfree-local`

우선순위는 다음과 같습니다.

```text
nonfree-local > gpl-local > lgpl-local
```

GPL 라이브러리나 `--enable-gpl`은 빌드를 `gpl-local`로 올립니다. nonfree 라이브러리나 `--enable-nonfree`는 빌드를 `nonfree-local`로 올리고 재배포 경고를 냅니다.

FFmpeg의 version 3 라이선스 스위치가 필요한 라이브러리는 `--enable-version3`를 자동으로 추가하고 정보성 경고를 냅니다. 현재 구현된 version 3 대상은 다음과 같습니다.

- `opencore-amr`
- `vo-amrwbenc`
- `libaribb24`
- `lensfun`

`--enable-version3`는 별도 라이선스 프로필 이름이 아닙니다. 계산된 로컬 프로필 안에서 추가되는 최종 configure 플래그이자 경고 계층입니다.

## 지원 범위를 읽는 기준

라이브러리는 선택한 FFmpeg 버전과 셸 프로필에서 필요한 패키지, 소스 준비, import 경로를 프로그램이 검토와 승인 절차에 포함할 수 있을 때 온전히 지원된다고 볼 수 있습니다.

다음 표현들은 서로 다른 상태를 가리킵니다.

- 카탈로그에 있다.
- UI에 보인다.
- 일반 사용에서 선택할 수 있다.
- 공개 프리셋이 선택한다.
- 구현된 레시피로 소스 준비가 가능하다.
- 외부 SDK나 경로에서 가져온다.
- FFmpeg configure로 전달된다.
- FFmpeg configure가 최종적으로 받아들인다.

이 프로그램은 upstream FFmpeg에 configure 플래그가 있다는 이유만으로 지원한다고 쓰지 않습니다.
