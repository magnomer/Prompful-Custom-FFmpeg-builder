// Package shared contains the small manipulation surface used by coded
// version/library work hooks under /versions/{ffmpegVersion}.
package shared

// LSourcePatchEntry is one exact full-line replacement applied to the extracted source tree.
type LSourcePatchEntry struct {
	File    string
	Find    string
	Replace string
}

// LGeneratedFile is a file written into the extracted source tree before configure/build.
type LGeneratedFile struct {
	Path  string
	Lines []string
}

// LPackagePatchEntry replaces the Libs line in an installed pkg-config file.
type LPackagePatchEntry struct {
	Module   string
	LibsLine string
}

// LPreparationPlan is the mutable plan changed by a version/library hook.
// It intentionally contains no download URL, hash, source version, or FFmpeg question-answering
// metadata. Source pins are resolved elsewhere; this object is only for build manipulation.
type LPreparationPlan struct {
	FfmpegVersion         string
	LibraryId             string
	VersionSpecificGoFile string
	DisplayName           string
	TrackName             string
	Method                string
	BuildSystem           string

	BuildDependencyPackages     []string
	MsysBuildDependencyPackages []string
	CFlags                      []string

	CMakeOptions      []string
	CMakeBuildTargets []string

	ConfigureSubdir    string
	ConfigureOptions   []string
	MakeBuildTargets   []string
	MakeInstallTargets []string
	RunAutogen         bool

	MakeVariables          []string
	MakeInstallHeaderFiles []string
	MakeStaticLibFile      string

	ImportIncludeSubdir string
	ImportLibSubdir     string

	PkgConfigName            string
	PkgConfigAppendLibs      []string
	PkgConfigAppendCFlags    []string
	PkgConfigLibsLine        string
	PkgConfigLibsLinePatches []LPackagePatchEntry
	PrivatePrefixInstall     bool

	VerifyHeaderRelativePath string
	VerifyLibStem            string

	SourcePatches        []LSourcePatchEntry
	GeneratedSourceFiles []LGeneratedFile
}

// LPreparationManipulator is real coded version/library manipulation.
type LPreparationManipulator func(*LPreparationPlan)

func LPreparationPlanCreate(ffmpegVersion string, libraryId string, goFilePath string) *LPreparationPlan {
	return &LPreparationPlan{FfmpegVersion: ffmpegVersion, LibraryId: libraryId, VersionSpecificGoFile: goFilePath}
}

func (plan *LPreparationPlan) LSourceCompilationUse(displayName string, buildSystem string) {
	plan.DisplayName = displayName
	plan.TrackName = "internal"
	plan.Method = "internal-source-build"
	plan.BuildSystem = buildSystem
}

func (plan *LPreparationPlan) LVendorSourceUse(displayName string) {
	plan.DisplayName = displayName
	plan.TrackName = "external"
	plan.Method = "external-vendor-import"
}

func (plan *LPreparationPlan) LBuildPackageRequire(packageSuffixes ...string) {
	plan.BuildDependencyPackages = append(plan.BuildDependencyPackages, packageSuffixes...)
}

func (plan *LPreparationPlan) LMSYSPackageRequire(packageNames ...string) {
	plan.MsysBuildDependencyPackages = append(plan.MsysBuildDependencyPackages, packageNames...)
}

func (plan *LPreparationPlan) LCompilerFlagAdd(flags ...string) {
	plan.CFlags = append(plan.CFlags, flags...)
}

func (plan *LPreparationPlan) LCMakeOptionAdd(options ...string) {
	plan.CMakeOptions = append(plan.CMakeOptions, options...)
}

func (plan *LPreparationPlan) LCMakeTargetAdd(targets ...string) {
	plan.CMakeBuildTargets = append(plan.CMakeBuildTargets, targets...)
}

func (plan *LPreparationPlan) LConfigureSubdirectoryUse(subdir string) {
	plan.ConfigureSubdir = subdir
}

func (plan *LPreparationPlan) LConfigureOptionAdd(options ...string) {
	plan.ConfigureOptions = append(plan.ConfigureOptions, options...)
}

func (plan *LPreparationPlan) LMakeTargetAdd(targets ...string) {
	plan.MakeBuildTargets = append(plan.MakeBuildTargets, targets...)
}

func (plan *LPreparationPlan) LMakeTargetInstall(targets ...string) {
	plan.MakeInstallTargets = append(plan.MakeInstallTargets, targets...)
}

func (plan *LPreparationPlan) LAutogenBeforeRun() {
	plan.RunAutogen = true
}

func (plan *LPreparationPlan) LMakeVariableAdd(variables ...string) {
	plan.MakeVariables = append(plan.MakeVariables, variables...)
}

func (plan *LPreparationPlan) LHeaderInstall(paths ...string) {
	plan.MakeInstallHeaderFiles = append(plan.MakeInstallHeaderFiles, paths...)
}

func (plan *LPreparationPlan) LLibraryStaticInstall(path string) {
	plan.MakeStaticLibFile = path
}

func (plan *LPreparationPlan) LSubdirectoryLoad(includeSubdir string, libSubdir string) {
	plan.ImportIncludeSubdir = includeSubdir
	plan.ImportLibSubdir = libSubdir
}

func (plan *LPreparationPlan) LPackageConfigurationUse(module string) {
	plan.PkgConfigName = module
}

func (plan *LPreparationPlan) LLibraryLineAppend(libs ...string) {
	plan.PkgConfigAppendLibs = append(plan.PkgConfigAppendLibs, libs...)
}

func (plan *LPreparationPlan) LCFlagAppend(flags ...string) {
	plan.PkgConfigAppendCFlags = append(plan.PkgConfigAppendCFlags, flags...)
}

func (plan *LPreparationPlan) LLibraryLineOverride(line string) {
	plan.PkgConfigLibsLine = line
}

func (plan *LPreparationPlan) LModuleLineOverride(module string, line string) {
	plan.PkgConfigLibsLinePatches = append(plan.PkgConfigLibsLinePatches, LPackagePatchEntry{Module: module, LibsLine: line})
}

func (plan *LPreparationPlan) LInstallPrivateUse() {
	plan.PrivatePrefixInstall = true
}

func (plan *LPreparationPlan) LCommandVerify(headerRelativePath string, libraryStem string) {
	plan.VerifyHeaderRelativePath = headerRelativePath
	plan.VerifyLibStem = libraryStem
}

func (plan *LPreparationPlan) LPreparationModificationAdd(file string, find string, replace string) {
	plan.SourcePatches = append(plan.SourcePatches, LSourcePatchEntry{File: file, Find: find, Replace: replace})
}

func (plan *LPreparationPlan) LGeneratedFileWrite(path string, lines ...string) {
	plan.GeneratedSourceFiles = append(plan.GeneratedSourceFiles, LGeneratedFile{Path: path, Lines: append([]string{}, lines...)})
}
