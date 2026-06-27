export namespace app {
	
	export class ApprovedActionResult {
	    runId: string;
	    startedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new ApprovedActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.startedAt = source["startedAt"];
	    }
	}
	export class BuildResultFile {
	    name: string;
	    path: string;
	    sizeBytes: number;
	    sha256Hash: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildResultFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.sizeBytes = source["sizeBytes"];
	        this.sha256Hash = source["sha256Hash"];
	    }
	}
	export class BuildResult {
	    artifactsDirectory: string;
	    reportPath: string;
	    files: BuildResultFile[];
	    selectedLibraries: string[];
	    selectedConfigureOptions: string[];
	    requiredMsys2PackageNames: string[];
	    configureFlags: string[];
	    licenseProfileName: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.artifactsDirectory = source["artifactsDirectory"];
	        this.reportPath = source["reportPath"];
	        this.files = this.convertValues(source["files"], BuildResultFile);
	        this.selectedLibraries = source["selectedLibraries"];
	        this.selectedConfigureOptions = source["selectedConfigureOptions"];
	        this.requiredMsys2PackageNames = source["requiredMsys2PackageNames"];
	        this.configureFlags = source["configureFlags"];
	        this.licenseProfileName = source["licenseProfileName"];
	        this.createdAt = source["createdAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class LibraryVerification {
	    libraryId: string;
	    displayName: string;
	    method: string;
	    expectedFlags: string[];
	    missingFlags: string[];
	    components: string[];
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new LibraryVerification(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.libraryId = source["libraryId"];
	        this.displayName = source["displayName"];
	        this.method = source["method"];
	        this.expectedFlags = source["expectedFlags"];
	        this.missingFlags = source["missingFlags"];
	        this.components = source["components"];
	        this.status = source["status"];
	    }
	}
	export class BuildVerification {
	    ffmpegPath: string;
	    ffmpegVersion: string;
	    libraries: LibraryVerification[];
	    unexpectedEnableFlags: string[];
	    okCount: number;
	    totalCount: number;
	    overall: string;
	    message: string;
	    verifiedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildVerification(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ffmpegPath = source["ffmpegPath"];
	        this.ffmpegVersion = source["ffmpegVersion"];
	        this.libraries = this.convertValues(source["libraries"], LibraryVerification);
	        this.unexpectedEnableFlags = source["unexpectedEnableFlags"];
	        this.okCount = source["okCount"];
	        this.totalCount = source["totalCount"];
	        this.overall = source["overall"];
	        this.message = source["message"];
	        this.verifiedAt = source["verifiedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InitialApplicationState {
	    hostOs: string;
	    kindExplanation: string;
	    securityRuleSummary: string;
	    namingRuleSummary: string;
	    defaultBuildConfigSettings: planning.BuildConfigSettings;
	    defaultFfmpegBuildSettings: planning.FfmpegBuildSettings;
	    defaultLibraryCatalog: planning.LibraryChoice[];
	    defaultConfigureOptionCatalog: planning.ConfigureOptionChoice[];
	
	    static createFrom(source: any = {}) {
	        return new InitialApplicationState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hostOs = source["hostOs"];
	        this.kindExplanation = source["kindExplanation"];
	        this.securityRuleSummary = source["securityRuleSummary"];
	        this.namingRuleSummary = source["namingRuleSummary"];
	        this.defaultBuildConfigSettings = this.convertValues(source["defaultBuildConfigSettings"], planning.BuildConfigSettings);
	        this.defaultFfmpegBuildSettings = this.convertValues(source["defaultFfmpegBuildSettings"], planning.FfmpegBuildSettings);
	        this.defaultLibraryCatalog = this.convertValues(source["defaultLibraryCatalog"], planning.LibraryChoice);
	        this.defaultConfigureOptionCatalog = this.convertValues(source["defaultConfigureOptionCatalog"], planning.ConfigureOptionChoice);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class LocalLogEntry {
	    level: string;
	    message: string;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new LocalLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.message = source["message"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class LocalLogRecord {
	    runId: string;
	    createdAt: string;
	    displayTime: string;
	    kind: string;
	    status: string;
	    directory: string;
	    entries: LocalLogEntry[];
	    rawText: string;
	    errorCount: number;
	    warnCount: number;
	    hasStdoutLog: boolean;
	    hasStderrLog: boolean;
	    hasSecurityEvents: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LocalLogRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.createdAt = source["createdAt"];
	        this.displayTime = source["displayTime"];
	        this.kind = source["kind"];
	        this.status = source["status"];
	        this.directory = source["directory"];
	        this.entries = this.convertValues(source["entries"], LocalLogEntry);
	        this.rawText = source["rawText"];
	        this.errorCount = source["errorCount"];
	        this.warnCount = source["warnCount"];
	        this.hasStdoutLog = source["hasStdoutLog"];
	        this.hasStderrLog = source["hasStderrLog"];
	        this.hasSecurityEvents = source["hasSecurityEvents"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ToolchainStatus {
	    installed: boolean;
	    healthy: boolean;
	    msys2RootDirectory: string;
	    createdAt: string;
	    windowsShellProfileName: string;
	    msys2ArchiveUrl: string;
	    packageCount: number;
	    packageNames: string[];
	    planHash: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolchainStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.healthy = source["healthy"];
	        this.msys2RootDirectory = source["msys2RootDirectory"];
	        this.createdAt = source["createdAt"];
	        this.windowsShellProfileName = source["windowsShellProfileName"];
	        this.msys2ArchiveUrl = source["msys2ArchiveUrl"];
	        this.packageCount = source["packageCount"];
	        this.packageNames = source["packageNames"];
	        this.planHash = source["planHash"];
	    }
	}
	export class ToolchainVerification {
	    verified: boolean;
	    checkedPackageCount: number;
	    missingPackageNames: string[];
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolchainVerification(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.verified = source["verified"];
	        this.checkedPackageCount = source["checkedPackageCount"];
	        this.missingPackageNames = source["missingPackageNames"];
	        this.message = source["message"];
	    }
	}

}

export namespace consent {
	
	export class ApprovalRequest {
	    approvedActionName: string;
	    approvedPlanHash: string;
	    consentText: string;
	
	    static createFrom(source: any = {}) {
	        return new ApprovalRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.approvedActionName = source["approvedActionName"];
	        this.approvedPlanHash = source["approvedPlanHash"];
	        this.consentText = source["consentText"];
	    }
	}

}

export namespace planning {
	
	export class BuildConfigSettings {
	    workspaceDirectory: string;
	    msys2ArchiveUrl: string;
	    msys2ArchiveSha256Hash: string;
	    msys2ArchiveSignatureUrl: string;
	    msys2PackageNames: string[];
	    windowsShellProfileName: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildConfigSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workspaceDirectory = source["workspaceDirectory"];
	        this.msys2ArchiveUrl = source["msys2ArchiveUrl"];
	        this.msys2ArchiveSha256Hash = source["msys2ArchiveSha256Hash"];
	        this.msys2ArchiveSignatureUrl = source["msys2ArchiveSignatureUrl"];
	        this.msys2PackageNames = source["msys2PackageNames"];
	        this.windowsShellProfileName = source["windowsShellProfileName"];
	    }
	}
	export class ConfigureOptionChoice {
	    optionId: string;
	    displayName: string;
	    categoryName: string;
	    configureFlags: string[];
	    plainExplanation: string;
	    technicalNote: string;
	    defaultEnabled: boolean;
	    locked: boolean;
	    riskLevelName: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigureOptionChoice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.optionId = source["optionId"];
	        this.displayName = source["displayName"];
	        this.categoryName = source["categoryName"];
	        this.configureFlags = source["configureFlags"];
	        this.plainExplanation = source["plainExplanation"];
	        this.technicalNote = source["technicalNote"];
	        this.defaultEnabled = source["defaultEnabled"];
	        this.locked = source["locked"];
	        this.riskLevelName = source["riskLevelName"];
	    }
	}
	export class PlanWarning {
	    riskLevelName: string;
	    message: string;
	    messageKey?: string;
	    messageValues?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new PlanWarning(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.riskLevelName = source["riskLevelName"];
	        this.message = source["message"];
	        this.messageKey = source["messageKey"];
	        this.messageValues = source["messageValues"];
	    }
	}
	export class PlanOperation {
	    operationName: string;
	    summary: string;
	    summaryKey?: string;
	    summaryValues?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new PlanOperation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.operationName = source["operationName"];
	        this.summary = source["summary"];
	        this.summaryKey = source["summaryKey"];
	        this.summaryValues = source["summaryValues"];
	    }
	}
	export class LibrarySourcePatch {
	    file: string;
	    find: string;
	    replace: string;
	
	    static createFrom(source: any = {}) {
	        return new LibrarySourcePatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.find = source["find"];
	        this.replace = source["replace"];
	    }
	}
	export class LibraryPreparation {
	    libraryId: string;
	    displayName: string;
	    trackName: string;
	    method: string;
	    buildSystem: string;
	    version: string;
	    buildDependencyPackages?: string[];
	    msysBuildDependencyPackages?: string[];
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
	    runAutogen?: boolean;
	    makeVariables?: string[];
	    makeInstallHeaderFiles?: string[];
	    makeStaticLibFile?: string;
	    importIncludeSubdir?: string;
	    importLibSubdir?: string;
	    pkgConfigName?: string;
	    pkgConfigAppendLibs?: string[];
	    pkgConfigLibsLine?: string;
	    privatePrefixInstall?: boolean;
	    verifyHeaderRelativePath: string;
	    verifyLibStem: string;
	    sourcePatches?: LibrarySourcePatch[];
	
	    static createFrom(source: any = {}) {
	        return new LibraryPreparation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.libraryId = source["libraryId"];
	        this.displayName = source["displayName"];
	        this.trackName = source["trackName"];
	        this.method = source["method"];
	        this.buildSystem = source["buildSystem"];
	        this.version = source["version"];
	        this.buildDependencyPackages = source["buildDependencyPackages"];
	        this.msysBuildDependencyPackages = source["msysBuildDependencyPackages"];
	        this.archiveUrl = source["archiveUrl"];
	        this.archiveSha256Hash = source["archiveSha256Hash"];
	        this.allowedDownloadHost = source["allowedDownloadHost"];
	        this.archiveFormat = source["archiveFormat"];
	        this.cmakeOptions = source["cmakeOptions"];
	        this.cmakeBuildTargets = source["cmakeBuildTargets"];
	        this.configureSubdir = source["configureSubdir"];
	        this.configureOptions = source["configureOptions"];
	        this.makeBuildTargets = source["makeBuildTargets"];
	        this.makeInstallTargets = source["makeInstallTargets"];
	        this.runAutogen = source["runAutogen"];
	        this.makeVariables = source["makeVariables"];
	        this.makeInstallHeaderFiles = source["makeInstallHeaderFiles"];
	        this.makeStaticLibFile = source["makeStaticLibFile"];
	        this.importIncludeSubdir = source["importIncludeSubdir"];
	        this.importLibSubdir = source["importLibSubdir"];
	        this.pkgConfigName = source["pkgConfigName"];
	        this.pkgConfigAppendLibs = source["pkgConfigAppendLibs"];
	        this.pkgConfigLibsLine = source["pkgConfigLibsLine"];
	        this.privatePrefixInstall = source["privatePrefixInstall"];
	        this.verifyHeaderRelativePath = source["verifyHeaderRelativePath"];
	        this.verifyLibStem = source["verifyLibStem"];
	        this.sourcePatches = this.convertValues(source["sourcePatches"], LibrarySourcePatch);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TrackedLibrarySelection {
	    trackName: string;
	    libraries: LibraryChoice[];
	
	    static createFrom(source: any = {}) {
	        return new TrackedLibrarySelection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trackName = source["trackName"];
	        this.libraries = this.convertValues(source["libraries"], LibraryChoice);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LibraryChoice {
	    libraryId: string;
	    trackName: string;
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
	
	    static createFrom(source: any = {}) {
	        return new LibraryChoice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.libraryId = source["libraryId"];
	        this.trackName = source["trackName"];
	        this.displayName = source["displayName"];
	        this.categoryName = source["categoryName"];
	        this.configureFlags = source["configureFlags"];
	        this.packageNames = source["packageNames"];
	        this.officialWebpageUrl = source["officialWebpageUrl"];
	        this.licenseEffectName = source["licenseEffectName"];
	        this.plainExplanation = source["plainExplanation"];
	        this.technicalExplanation = source["technicalExplanation"];
	        this.defaultChecked = source["defaultChecked"];
	        this.locked = source["locked"];
	    }
	}
	export class FfmpegBuildPlan {
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
	    downloadConflictPolicyName: string;
	    extractionDestinationPolicyName: string;
	    operations: PlanOperation[];
	    warnings: PlanWarning[];
	    isExecutable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FfmpegBuildPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.actionName = source["actionName"];
	        this.planHash = source["planHash"];
	        this.workspaceDirectory = source["workspaceDirectory"];
	        this.msys2RootDirectory = source["msys2RootDirectory"];
	        this.ffmpegSourceArchiveUrl = source["ffmpegSourceArchiveUrl"];
	        this.ffmpegSourceSignatureUrl = source["ffmpegSourceSignatureUrl"];
	        this.ffmpegSourceSha256Hash = source["ffmpegSourceSha256Hash"];
	        this.selectedLibraryIds = source["selectedLibraryIds"];
	        this.selectedLibraries = this.convertValues(source["selectedLibraries"], LibraryChoice);
	        this.selectedNativeLibraries = this.convertValues(source["selectedNativeLibraries"], LibraryChoice);
	        this.selectedInternalLibraries = this.convertValues(source["selectedInternalLibraries"], LibraryChoice);
	        this.selectedExternalLibraries = this.convertValues(source["selectedExternalLibraries"], LibraryChoice);
	        this.selectedLibrariesByTrack = this.convertValues(source["selectedLibrariesByTrack"], TrackedLibrarySelection);
	        this.libraryPreparations = this.convertValues(source["libraryPreparations"], LibraryPreparation);
	        this.requiredMsys2PackageNames = source["requiredMsys2PackageNames"];
	        this.generatedConfigureFlags = source["generatedConfigureFlags"];
	        this.selectedConfigureOptions = this.convertValues(source["selectedConfigureOptions"], ConfigureOptionChoice);
	        this.generatedOptionFlags = source["generatedOptionFlags"];
	        this.extraConfigureFlags = source["extraConfigureFlags"];
	        this.configureFlags = source["configureFlags"];
	        this.parallelJobCount = source["parallelJobCount"];
	        this.windowsShellProfileName = source["windowsShellProfileName"];
	        this.licenseProfileName = source["licenseProfileName"];
	        this.willModifySystemPath = source["willModifySystemPath"];
	        this.willRequireAdminRights = source["willRequireAdminRights"];
	        this.willUseExistingMsys2 = source["willUseExistingMsys2"];
	        this.willDeleteFiles = source["willDeleteFiles"];
	        this.downloadConflictPolicyName = source["downloadConflictPolicyName"];
	        this.extractionDestinationPolicyName = source["extractionDestinationPolicyName"];
	        this.operations = this.convertValues(source["operations"], PlanOperation);
	        this.warnings = this.convertValues(source["warnings"], PlanWarning);
	        this.isExecutable = source["isExecutable"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FfmpegBuildPlanReview {
	    reviewSessionId: string;
	    expectedConsentText: string;
	    expectedConsentTextHash: string;
	    expiresAtUnixTime: number;
	    plan: FfmpegBuildPlan;
	
	    static createFrom(source: any = {}) {
	        return new FfmpegBuildPlanReview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reviewSessionId = source["reviewSessionId"];
	        this.expectedConsentText = source["expectedConsentText"];
	        this.expectedConsentTextHash = source["expectedConsentTextHash"];
	        this.expiresAtUnixTime = source["expiresAtUnixTime"];
	        this.plan = this.convertValues(source["plan"], FfmpegBuildPlan);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FfmpegBuildSettings {
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
	
	    static createFrom(source: any = {}) {
	        return new FfmpegBuildSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workspaceDirectory = source["workspaceDirectory"];
	        this.ffmpegSourceArchiveUrl = source["ffmpegSourceArchiveUrl"];
	        this.ffmpegSourceSignatureUrl = source["ffmpegSourceSignatureUrl"];
	        this.ffmpegSourceSha256Hash = source["ffmpegSourceSha256Hash"];
	        this.selectedLibraryIds = source["selectedLibraryIds"];
	        this.selectedConfigureOptionIds = source["selectedConfigureOptionIds"];
	        this.extraConfigureFlags = source["extraConfigureFlags"];
	        this.configureFlags = source["configureFlags"];
	        this.parallelJobCount = source["parallelJobCount"];
	        this.windowsShellProfileName = source["windowsShellProfileName"];
	        this.licenseProfileName = source["licenseProfileName"];
	    }
	}
	
	
	
	
	
	export class ToolchainPreparationPlan {
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
	    downloadConflictPolicyName: string;
	    extractionDestinationPolicyName: string;
	    operations: PlanOperation[];
	    warnings: PlanWarning[];
	    isExecutable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToolchainPreparationPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.actionName = source["actionName"];
	        this.planHash = source["planHash"];
	        this.workspaceDirectory = source["workspaceDirectory"];
	        this.msys2RootDirectory = source["msys2RootDirectory"];
	        this.msys2ArchiveUrl = source["msys2ArchiveUrl"];
	        this.msys2ArchiveSha256Hash = source["msys2ArchiveSha256Hash"];
	        this.msys2ArchiveSignatureUrl = source["msys2ArchiveSignatureUrl"];
	        this.msys2PackageNames = source["msys2PackageNames"];
	        this.windowsShellProfileName = source["windowsShellProfileName"];
	        this.willModifySystemPath = source["willModifySystemPath"];
	        this.willRequireAdminRights = source["willRequireAdminRights"];
	        this.willUseExistingMsys2 = source["willUseExistingMsys2"];
	        this.willDeleteFiles = source["willDeleteFiles"];
	        this.downloadConflictPolicyName = source["downloadConflictPolicyName"];
	        this.extractionDestinationPolicyName = source["extractionDestinationPolicyName"];
	        this.operations = this.convertValues(source["operations"], PlanOperation);
	        this.warnings = this.convertValues(source["warnings"], PlanWarning);
	        this.isExecutable = source["isExecutable"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ToolchainPreparationPlanReview {
	    reviewSessionId: string;
	    expectedConsentText: string;
	    expectedConsentTextHash: string;
	    expiresAtUnixTime: number;
	    plan: ToolchainPreparationPlan;
	
	    static createFrom(source: any = {}) {
	        return new ToolchainPreparationPlanReview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reviewSessionId = source["reviewSessionId"];
	        this.expectedConsentText = source["expectedConsentText"];
	        this.expectedConsentTextHash = source["expectedConsentTextHash"];
	        this.expiresAtUnixTime = source["expiresAtUnixTime"];
	        this.plan = this.convertValues(source["plan"], ToolchainPreparationPlan);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

