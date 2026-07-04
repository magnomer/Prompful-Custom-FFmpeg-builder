package version519

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryLibplaceboPrepare performs the coded preparation manipulation for libplacebo on FFmpeg 5.1.9.
func LLibraryLibplaceboPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "5.1.9"
	plan.LibraryId = "libplacebo"
	plan.VersionSpecificGoFile = "versions/5.1.9/libplacebo.go"
	plan.LSourceCompilationUse("libplacebo", "meson")
	plan.LBuildPackageRequire("meson", "ninja", "python-mako", "vulkan-headers", "vulkan-loader", "shaderc", "lcms2")
	plan.LCMakeOptionAdd("-Ddemos=false", "-Dtests=false", "-Dopengl=disabled", "-Dd3d11=disabled", "-Dglslang=disabled", "-Dshaderc=enabled", "-Dvulkan=enabled", "-Dlcms=enabled")
	plan.LPreparationModificationAdd("src/vulkan/utils_gen.py", "    registry = ET.parse(xmlfile)", "    registry = ET.parse(xmlfile); [node.clear() for node in registry.iterfind('.//feature[@api=\"vulkansc\"]')]; [node.clear() for node in registry.iterfind('.//require[@api=\"vulkansc\"]')]; [node.clear() for node in registry.iterfind('.//require[@depends=\"VKSC_VERSION_1_0\"]')]; [node.clear() for node in registry.iterfind('.//extension[@platform]')]; [node.clear() for node in registry.iterfind('.//extension[@supported=\"vulkansc\"]')]")
	plan.LPreparationModificationAdd("src/vulkan/utils_gen.py", "        if 'objtypeenum' in t.attrib:", "        if 'objtypeenum' in t.attrib and not any([str in t.attrib['objtypeenum'] for str in ['SCI', 'OHOS', 'QNX']]):")
	plan.LPreparationModificationAdd("src/vulkan/utils_gen.py", "            'ANDROID', 'Surface', 'Win32', 'D3D12', 'GGP', 'FUCHSIA',", "            'ANDROID', 'Surface', 'Win32', 'D3D12', 'GGP', 'FUCHSIA', 'Metal', 'Sci', 'QNX', 'OHOS', 'VulkanSC', 'Reservation', 'ApplicationParameters', 'Fault', 'RefreshObjectList', 'PipelineOffline', 'PipelinePool', 'CommandPoolMemory',")
	plan.LPackageConfigurationUse("libplacebo")
	plan.LLibraryLineOverride("-L${libdir} -l:libplacebo.a -lm -lstdc++")
	plan.LCommandVerify("libplacebo/renderer.h", "placebo")
}
