// Package shared contains the small manipulation surface used by coded
// version/library work hooks under /versions/{ffmpegVersion}.
package shared

// SourcePatch is one exact full-line replacement applied to the extracted source tree.
type SourcePatch struct {
	File    string
	Find    string
	Replace string
}

// GeneratedFile is a file written into the extracted source tree before configure/build.
type GeneratedFile struct {
	Path  string
	Lines []string
}

// PkgConfigLibsLinePatch replaces the Libs line in an installed pkg-config file.
type PkgConfigLibsLinePatch struct {
	Module   string
	LibsLine string
}

// LibraryPreparationPlan is the mutable plan changed by a version/library hook.
// It intentionally contains no download URL, hash, source version, or FFmpeg question-answering
// metadata. Source pins are resolved elsewhere; this object is only for build manipulation.
type LibraryPreparationPlan struct {
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
	PkgConfigLibsLinePatches []PkgConfigLibsLinePatch
	PrivatePrefixInstall     bool

	VerifyHeaderRelativePath string
	VerifyLibStem            string

	SourcePatches        []SourcePatch
	GeneratedSourceFiles []GeneratedFile
}

// LibraryPreparationManipulator is real coded version/library manipulation.
type LibraryPreparationManipulator func(*LibraryPreparationPlan)

func NewLibraryPreparationPlan(ffmpegVersion string, libraryId string, goFilePath string) *LibraryPreparationPlan {
	return &LibraryPreparationPlan{FfmpegVersion: ffmpegVersion, LibraryId: libraryId, VersionSpecificGoFile: goFilePath}
}

func (plan *LibraryPreparationPlan) UseInternalSourceBuild(displayName string, buildSystem string) {
	plan.DisplayName = displayName
	plan.TrackName = "internal"
	plan.Method = "internal-source-build"
	plan.BuildSystem = buildSystem
}

func (plan *LibraryPreparationPlan) UseExternalVendorImport(displayName string) {
	plan.DisplayName = displayName
	plan.TrackName = "external"
	plan.Method = "external-vendor-import"
}

func (plan *LibraryPreparationPlan) RequireBuildPackages(packageSuffixes ...string) {
	plan.BuildDependencyPackages = append(plan.BuildDependencyPackages, packageSuffixes...)
}

func (plan *LibraryPreparationPlan) RequireMsysBuildPackages(packageNames ...string) {
	plan.MsysBuildDependencyPackages = append(plan.MsysBuildDependencyPackages, packageNames...)
}

func (plan *LibraryPreparationPlan) AddCFlags(flags ...string) {
	plan.CFlags = append(plan.CFlags, flags...)
}

func (plan *LibraryPreparationPlan) AddCMakeOptions(options ...string) {
	plan.CMakeOptions = append(plan.CMakeOptions, options...)
}

func (plan *LibraryPreparationPlan) BuildCMakeTargets(targets ...string) {
	plan.CMakeBuildTargets = append(plan.CMakeBuildTargets, targets...)
}

func (plan *LibraryPreparationPlan) ConfigureInSubdir(subdir string) {
	plan.ConfigureSubdir = subdir
}

func (plan *LibraryPreparationPlan) AddConfigureOptions(options ...string) {
	plan.ConfigureOptions = append(plan.ConfigureOptions, options...)
}

func (plan *LibraryPreparationPlan) BuildMakeTargets(targets ...string) {
	plan.MakeBuildTargets = append(plan.MakeBuildTargets, targets...)
}

func (plan *LibraryPreparationPlan) InstallMakeTargets(targets ...string) {
	plan.MakeInstallTargets = append(plan.MakeInstallTargets, targets...)
}

func (plan *LibraryPreparationPlan) RunAutogenBeforeConfigure() {
	plan.RunAutogen = true
}

func (plan *LibraryPreparationPlan) AddMakeVariables(variables ...string) {
	plan.MakeVariables = append(plan.MakeVariables, variables...)
}

func (plan *LibraryPreparationPlan) InstallMakeHeaders(paths ...string) {
	plan.MakeInstallHeaderFiles = append(plan.MakeInstallHeaderFiles, paths...)
}

func (plan *LibraryPreparationPlan) InstallMakeStaticLibrary(path string) {
	plan.MakeStaticLibFile = path
}

func (plan *LibraryPreparationPlan) ImportFromSubdirs(includeSubdir string, libSubdir string) {
	plan.ImportIncludeSubdir = includeSubdir
	plan.ImportLibSubdir = libSubdir
}

func (plan *LibraryPreparationPlan) UsePkgConfig(module string) {
	plan.PkgConfigName = module
}

func (plan *LibraryPreparationPlan) AppendPkgConfigLibs(libs ...string) {
	plan.PkgConfigAppendLibs = append(plan.PkgConfigAppendLibs, libs...)
}

func (plan *LibraryPreparationPlan) AppendPkgConfigCFlags(flags ...string) {
	plan.PkgConfigAppendCFlags = append(plan.PkgConfigAppendCFlags, flags...)
}

func (plan *LibraryPreparationPlan) OverridePkgConfigLibsLine(line string) {
	plan.PkgConfigLibsLine = line
}

func (plan *LibraryPreparationPlan) OverridePkgConfigModuleLibsLine(module string, line string) {
	plan.PkgConfigLibsLinePatches = append(plan.PkgConfigLibsLinePatches, PkgConfigLibsLinePatch{Module: module, LibsLine: line})
}

func (plan *LibraryPreparationPlan) UsePrivatePrefixInstall() {
	plan.PrivatePrefixInstall = true
}

func (plan *LibraryPreparationPlan) Verify(headerRelativePath string, libraryStem string) {
	plan.VerifyHeaderRelativePath = headerRelativePath
	plan.VerifyLibStem = libraryStem
}

func (plan *LibraryPreparationPlan) PatchSource(file string, find string, replace string) {
	plan.SourcePatches = append(plan.SourcePatches, SourcePatch{File: file, Find: find, Replace: replace})
}

func (plan *LibraryPreparationPlan) WriteGeneratedSourceFile(path string, lines ...string) {
	plan.GeneratedSourceFiles = append(plan.GeneratedSourceFiles, GeneratedFile{Path: path, Lines: append([]string{}, lines...)})
}
