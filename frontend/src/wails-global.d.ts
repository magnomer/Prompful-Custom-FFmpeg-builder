type PlanWarning = {
  // Wails-generated Go models expose this as string, so keep the frontend type compatible.
  riskLevelName: string;
  message: string;
};

type PlanOperation = {
  operationName: string;
  summary: string;
};

type LibraryChoice = {
  libraryId: string;
  displayName: string;
  categoryName: string;
  configureFlags: string[];
  packageNames: string[];
  licenseEffectName: string;
  reviewNote: string;
  plainExplanation: string;
  technicalExplanation: string;
  defaultChecked: boolean;
  locked: boolean;
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
};

type BuildToolSettings = {
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
  defaultBuildToolSettings: BuildToolSettings;
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
