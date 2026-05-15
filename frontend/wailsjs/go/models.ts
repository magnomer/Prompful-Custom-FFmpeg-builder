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

export namespace main {
	
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
	
	export class InitialApplicationState {
	    hostOs: string;
	    kindExplanation: string;
	    securityRuleSummary: string;
	    namingRuleSummary: string;
	    defaultBuildToolSettings: planning.BuildToolSettings;
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
	        this.defaultBuildToolSettings = this.convertValues(source["defaultBuildToolSettings"], planning.BuildToolSettings);
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

}

export namespace planning {
	
	export class BuildToolSettings {
	    workspaceDirectory: string;
	    msys2ArchiveUrl: string;
	    msys2ArchiveSha256Hash: string;
	    msys2ArchiveSignatureUrl: string;
	    msys2PackageNames: string[];
	    windowsShellProfileName: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildToolSettings(source);
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
	    }
	}
	export class PlanWarning {
	    riskLevelName: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new PlanWarning(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.riskLevelName = source["riskLevelName"];
	        this.message = source["message"];
	    }
	}
	export class PlanOperation {
	    operationName: string;
	    summary: string;
	
	    static createFrom(source: any = {}) {
	        return new PlanOperation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.operationName = source["operationName"];
	        this.summary = source["summary"];
	    }
	}
	export class LibraryChoice {
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
	
	    static createFrom(source: any = {}) {
	        return new LibraryChoice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.libraryId = source["libraryId"];
	        this.displayName = source["displayName"];
	        this.categoryName = source["categoryName"];
	        this.configureFlags = source["configureFlags"];
	        this.packageNames = source["packageNames"];
	        this.licenseEffectName = source["licenseEffectName"];
	        this.reviewNote = source["reviewNote"];
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

