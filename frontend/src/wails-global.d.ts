type PlanWarning = {
  // Wails-generated Go models expose this as string, so keep the frontend type compatible.
  riskLevelName: string;
  message: string;
  messageKey?: string;
  messageValues?: Record<string, string | number>;
};

type PlanOperation = {
  operationName: string;
  summary: string;
  summaryKey?: string;
  summaryValues?: Record<string, string | number>;
};

type LibraryChoice = {
  libraryId: string;
  trackName: 'native' | 'internal' | 'external' | string;
  displayName: string;
  categoryName: string;
  configureFlags: string[];
  packageNames: string[];
  officialWebpageUrl: string;
  licenseEffectName: string;
  plainExplanation: string;
  technicalExplanation: string;
  defaultChecked: boolean;
  locked: boolean;
};


type TrackedLibrarySelection = {
  trackName: 'native' | 'internal' | 'external' | string;
  libraries: LibraryChoice[];
};

type LibraryPreparation = {
  libraryId: string;
  displayName: string;
  trackName: 'internal' | 'external' | string;
  method: 'internal-source-build' | 'external-vendor-import' | string;
  buildSystem: 'cmake' | 'autotools' | 'make' | string;
  version: string;
  buildDependencyPackages?: string[];
  archiveUrl: string;
  archiveSha256Hash: string;
  allowedDownloadHost: string;
  archiveFormat: string;
  cmakeOptions?: string[];
  cmakeBuildTargets?: string[];
  configureSubdir?: string;
  configureOptions?: string[];
  makeBuildTargets?: string[];
  makeInstallTargets?: string[];
  importIncludeSubdir?: string;
  importLibSubdir?: string;
  pkgConfigName?: string;
  pkgConfigAppendLibs?: string[];
  verifyHeaderRelativePath: string;
  verifyLibStem: string;
};

type ConfigureOptionChoice = {
  optionId: string;
  displayName: string;
  categoryName: string;
  configureFlags: string[];
  plainExplanation: string;
  technicalNote: string;
  defaultEnabled: boolean;
  locked: boolean;
  riskLevelName: string;
};

type BuildConfigSettings = {
  workspaceDirectory: string;
  msys2ArchiveUrl: string;
  msys2ArchiveSha256Hash: string;
  msys2ArchiveSignatureUrl: string;
  msys2PackageNames: string[];
  windowsShellProfileName: string;
};

type FfmpegBuildSettings = {
  workspaceDirectory: string;
  ffmpegSourceArchiveUrl: string;
  ffmpegSourceSignatureUrl: string;
  ffmpegSourceSha256Hash: string;
  selectedLibraryIds: string[];
  selectedConfigureOptionIds: string[];
  extraConfigureFlags: string[];
  configureFlags: string[];
  parallelJobCount: number;
  windowsShellProfileName: string;
  licenseProfileName: string;
};

type ToolchainPreparationPlan = {
  actionName: string;
  planHash: string;
  workspaceDirectory: string;
  msys2RootDirectory: string;
  msys2ArchiveUrl: string;
  msys2ArchiveSha256Hash: string;
  msys2ArchiveSignatureUrl: string;
  msys2PackageNames: string[];
  windowsShellProfileName: string;
  willModifySystemPath: boolean;
  willRequireAdminRights: boolean;
  willUseExistingMsys2: boolean;
  willDeleteFiles: boolean;
  operations: PlanOperation[];
  warnings: PlanWarning[];
  isExecutable: boolean;
};

type FfmpegBuildPlan = {
  actionName: string;
  planHash: string;
  workspaceDirectory: string;
  msys2RootDirectory: string;
  ffmpegSourceArchiveUrl: string;
  ffmpegSourceSignatureUrl: string;
  ffmpegSourceSha256Hash: string;
  selectedLibraryIds: string[];
  selectedLibraries: LibraryChoice[];
  selectedNativeLibraries: LibraryChoice[];
  selectedInternalLibraries: LibraryChoice[];
  selectedExternalLibraries: LibraryChoice[];
  selectedLibrariesByTrack: TrackedLibrarySelection[];
  libraryPreparations: LibraryPreparation[];
  requiredMsys2PackageNames: string[];
  generatedConfigureFlags: string[];
  selectedConfigureOptions: ConfigureOptionChoice[];
  generatedOptionFlags: string[];
  extraConfigureFlags: string[];
  configureFlags: string[];
  parallelJobCount: number;
  windowsShellProfileName: string;
  licenseProfileName: string;
  willModifySystemPath: boolean;
  willRequireAdminRights: boolean;
  willUseExistingMsys2: boolean;
  willDeleteFiles: boolean;
  operations: PlanOperation[];
  warnings: PlanWarning[];
  isExecutable: boolean;
};

type ToolchainPreparationPlanReview = {
  reviewSessionId: string;
  expectedConsentText: string;
  expectedConsentTextHash: string;
  expiresAtUnixTime: number;
  plan: ToolchainPreparationPlan;
};

type FfmpegBuildPlanReview = {
  reviewSessionId: string;
  expectedConsentText: string;
  expectedConsentTextHash: string;
  expiresAtUnixTime: number;
  plan: FfmpegBuildPlan;
};

type InitialApplicationState = {
  hostOs: string;
  kindExplanation: string;
  securityRuleSummary: string;
  namingRuleSummary: string;
  defaultBuildConfigSettings: BuildConfigSettings;
  defaultFfmpegBuildSettings: FfmpegBuildSettings;
  defaultLibraryCatalog: LibraryChoice[];
  defaultConfigureOptionCatalog: ConfigureOptionChoice[];
};

type ApprovalRequest = {
  approvedActionName: string;
  approvedPlanHash: string;
  consentText: string;
};

type ApprovedActionResult = {
  runId: string;
  startedAt: string;
};

type BuildResultFile = {
  name: string;
  path: string;
  sizeBytes: number;
  sha256Hash: string;
};

type BuildResult = {
  artifactsDirectory: string;
  reportPath: string;
  files: BuildResultFile[];
  selectedLibraries: string[];
  selectedConfigureOptions: string[];
  requiredMsys2PackageNames: string[];
  configureFlags: string[];
  licenseProfileName: string;
  createdAt: string;
};

type ToolchainStatus = {
  installed: boolean;
  healthy: boolean;
  msys2RootDirectory: string;
  createdAt: string;
  windowsShellProfileName: string;
  msys2ArchiveUrl: string;
  packageCount: number;
  packageNames: string[];
  planHash: string;
};

type ToolchainVerification = {
  verified: boolean;
  checkedPackageCount: number;
  missingPackageNames: string[];
  message: string;
};
