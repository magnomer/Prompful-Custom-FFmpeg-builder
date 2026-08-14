package version901

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryLibtlsPrepare performs the coded preparation manipulation for libtls on FFmpeg 9.0.1.
func LLibraryLibtlsPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "9.0.1"
	plan.LibraryId = "libtls"
	plan.VersionSpecificGoFile = "versions/9.0.1/libtls.go"
	plan.LSourceCompilationUse("libtls", "cmake")
	plan.LCMakeOptionAdd("-DBUILD_SHARED_LIBS=OFF", "-DLIBRESSL_APPS=OFF", "-DLIBRESSL_TESTS=OFF")
	plan.LPackageConfigurationUse("libtls")
	plan.LLibraryLineOverride("${libdir}/libtls.a ${libdir}/libssl.a ${libdir}/libcrypto.a -lws2_32 -lbcrypt -lntdll")
	plan.LInstallPrivateUse()
	plan.LCommandVerify("tls.h", "tls")
}
