package version812

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryLibtlsPrepare performs the coded preparation manipulation for libtls on FFmpeg 8.1.2.
func LLibraryLibtlsPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "8.1.2"
	plan.LibraryId = "libtls"
	plan.VersionSpecificGoFile = "versions/8.1.2/libtls.go"
	plan.UseInternalSourceBuild("libtls", "cmake")
	plan.AddCMakeOptions("-DBUILD_SHARED_LIBS=OFF", "-DLIBRESSL_APPS=OFF", "-DLIBRESSL_TESTS=OFF")
	plan.UsePkgConfig("libtls")
	plan.OverridePkgConfigLibsLine("${libdir}/libtls.a ${libdir}/libssl.a ${libdir}/libcrypto.a -lws2_32 -lbcrypt -lntdll")
	plan.UsePrivatePrefixInstall()
	plan.Verify("tls.h", "tls")
}
