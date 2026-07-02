type LWarningPlan = {
  // Wails-generated Go models expose this as string, so keep the frontend type compatible.
  riskLevelName: string;
  message: string;
  messageKey?: string;
  messageValues?: Record<string, string | number>;
};

type LOperationPlan = {
  operationName: string;
  summary: string;
  summaryKey?: string;
  summaryValues?: Record<string, string | number>;
};

type LLibraryChoice = {
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
  supportState?: string;
  preparationStatus?: {
    required: boolean;
    kind?: string;
    implemented: boolean;
    implementation?: string;
    implementationLanguage?: string;
    reason?: string;
  };
  unavailableReasons?: string[];
  unavailableProfiles?: string[];
  versionCompatibility?: {
    supported: boolean;
    available: boolean;
    minVersion?: string;
  };
};


type LLibraryTrackSelection = {
  trackName: 'native' | 'internal' | 'external' | string;
  libraries: LLibraryChoice[];
};

type LLibraryPreparation = {
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

type LOptionChoice = {
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

type LSettingsBuild = {
  workspaceDirectory: string;
  msys2ArchiveUrl: string;
  msys2ArchiveSha256Hash: string;
  msys2ArchiveSignatureUrl: string;
  msys2PackageNames: string[];
  windowsShellProfileName: string;
};

type LSettingsFFmpeg = {
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

type LPlanToolchain = {
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
  operations: LOperationPlan[];
  warnings: LWarningPlan[];
  isExecutable: boolean;
};

type LPlanFFmpeg = {
  actionName: string;
  planHash: string;
  workspaceDirectory: string;
  msys2RootDirectory: string;
  ffmpegSourceArchiveUrl: string;
  ffmpegSourceSignatureUrl: string;
  ffmpegSourceSha256Hash: string;
  selectedLibraryIds: string[];
  selectedLibraries: LLibraryChoice[];
  selectedNativeLibraries: LLibraryChoice[];
  selectedInternalLibraries: LLibraryChoice[];
  selectedExternalLibraries: LLibraryChoice[];
  selectedLibrariesByTrack: LLibraryTrackSelection[];
  libraryPreparations: LLibraryPreparation[];
  requiredMsys2PackageNames: string[];
  generatedConfigureFlags: string[];
  selectedConfigureOptions: LOptionChoice[];
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
  operations: LOperationPlan[];
  warnings: LWarningPlan[];
  isExecutable: boolean;
};

type LReviewToolchain = {
  reviewSessionId: string;
  expectedLConsentText: string;
  expectedLConsentTextHash: string;
  expiresAtUnixTime: number;
  plan: LPlanToolchain;
};

type LReviewFFmpeg = {
  reviewSessionId: string;
  expectedLConsentText: string;
  expectedLConsentTextHash: string;
  expiresAtUnixTime: number;
  plan: LPlanFFmpeg;
};

type LStateInitial = {
  hostOs: string;
  kindExplanation: string;
  securityRuleSummary: string;
  namingRuleSummary: string;
  defaultBuildConfigSettings: LSettingsBuild;
  defaultFfmpegBuildSettings: LSettingsFFmpeg;
  defaultLibraryCatalog: LLibraryChoice[];
  defaultLibraryPresetCatalog: LPresetLibraryChoice[];
  defaultConfigureOptionCatalog: LOptionChoice[];
  supportedFfmpegReleases: LReleaseChoice[];
};


type LPresetLibraryChoice = {
  presetId: string;
  libraryIds: string[];
  extendedLibraryIds?: string[];
  hidden?: boolean;
  dev?: boolean;
};

type LReleaseChoice = {
  version: string;
  codename: string;
  archiveUrl: string;
  signatureUrl: string;
};

type LRequestApproval = {
  approvedActionName: string;
  approvedPlanHash: string;
  consentText: string;
};

type LResultAction = {
  runId: string;
  startedAt: string;
};

type LFileResult = {
  name: string;
  path: string;
  sizeBytes: number;
  sha256Hash: string;
};

type LResultBuild = {
  artifactsDirectory: string;
  reportPath: string;
  ffmpegVersion: string;
  files: LFileResult[];
  selectedLibraries: string[];
  selectedConfigureOptions: string[];
  requiredMsys2PackageNames: string[];
  configureFlags: string[];
  licenseProfileName: string;
  createdAt: string;
};

type LVerificationLibrary = {
  libraryId: string;
  displayName: string;
  method: string;
  expectedFlags: string[];
  missingFlags: string[];
  components: string[];
  status: string;
};

type LVerificationBuild = {
  ffmpegPath: string;
  ffmpegVersion: string;
  libraries: LVerificationLibrary[];
  unexpectedEnableFlags: string[];
  okCount: number;
  totalCount: number;
  overall: string;
  message: string;
  verifiedAt: string;
};


type LLogLocalEntry = {
  level: 'info' | 'warn' | 'error';
  message: string;
  timestamp: string;
};

type LRecordLog = {
  runId: string;
  createdAt: string;
  displayTime: string;
  kind: 'toolchain' | 'ffmpeg' | 'unknown' | string;
  status: string;
  directory: string;
  entries: LLogLocalEntry[];
  rawText: string;
  errorCount: number;
  warnCount: number;
  hasStdoutLog: boolean;
  hasStderrLog: boolean;
  hasSecurityLAuditEvents: boolean;
};

type LStatusToolchain = {
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

type LVerificationToolchain = {
  verified: boolean;
  checkedPackageCount: number;
  missingPackageNames: string[];
  message: string;
};
