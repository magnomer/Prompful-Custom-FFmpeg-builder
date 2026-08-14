export namespace audit {
	
	export class LAuditWriter {
	    LIdentifierRun: string;
	    LDirectoryLog: string;
	    // Go type: sync
	    LMutex: any;
	
	    static createFrom(source: any = {}) {
	        return new LAuditWriter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.LIdentifierRun = source["LIdentifierRun"];
	        this.LDirectoryLog = source["LDirectoryLog"];
	        this.LMutex = this.convertValues(source["LMutex"], null);
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

export namespace consent {
	
	export class LArchiveConsentState {
	    consentId: string;
	    kind: string;
	    approvedActionName: string;
	    approvedPlanHash: string;
	    approvedAtUnixTime: number;
	    consentText: string;
	
	    static createFrom(source: any = {}) {
	        return new LArchiveConsentState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.consentId = source["consentId"];
	        this.kind = source["kind"];
	        this.approvedActionName = source["approvedActionName"];
	        this.approvedPlanHash = source["approvedPlanHash"];
	        this.approvedAtUnixTime = source["approvedAtUnixTime"];
	        this.consentText = source["consentText"];
	    }
	}
	export class LConsentCommand {
	    consentId: string;
	    kind: string;
	    approvedActionName: string;
	    approvedPlanHash: string;
	    approvedAtUnixTime: number;
	    consentText: string;
	
	    static createFrom(source: any = {}) {
	        return new LConsentCommand(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.consentId = source["consentId"];
	        this.kind = source["kind"];
	        this.approvedActionName = source["approvedActionName"];
	        this.approvedPlanHash = source["approvedPlanHash"];
	        this.approvedAtUnixTime = source["approvedAtUnixTime"];
	        this.consentText = source["consentText"];
	    }
	}
	export class LConsentFfmpeg {
	    consentId: string;
	    kind: string;
	    approvedActionName: string;
	    approvedPlanHash: string;
	    approvedAtUnixTime: number;
	    consentText: string;
	
	    static createFrom(source: any = {}) {
	        return new LConsentFfmpeg(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.consentId = source["consentId"];
	        this.kind = source["kind"];
	        this.approvedActionName = source["approvedActionName"];
	        this.approvedPlanHash = source["approvedPlanHash"];
	        this.approvedAtUnixTime = source["approvedAtUnixTime"];
	        this.consentText = source["consentText"];
	    }
	}
	export class LConsentMsys {
	    consentId: string;
	    kind: string;
	    approvedActionName: string;
	    approvedPlanHash: string;
	    approvedAtUnixTime: number;
	    consentText: string;
	
	    static createFrom(source: any = {}) {
	        return new LConsentMsys(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.consentId = source["consentId"];
	        this.kind = source["kind"];
	        this.approvedActionName = source["approvedActionName"];
	        this.approvedPlanHash = source["approvedPlanHash"];
	        this.approvedAtUnixTime = source["approvedAtUnixTime"];
	        this.consentText = source["consentText"];
	    }
	}
	export class LConsentPacman {
	    consentId: string;
	    kind: string;
	    approvedActionName: string;
	    approvedPlanHash: string;
	    approvedAtUnixTime: number;
	    consentText: string;
	
	    static createFrom(source: any = {}) {
	        return new LConsentPacman(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.consentId = source["consentId"];
	        this.kind = source["kind"];
	        this.approvedActionName = source["approvedActionName"];
	        this.approvedPlanHash = source["approvedPlanHash"];
	        this.approvedAtUnixTime = source["approvedAtUnixTime"];
	        this.consentText = source["consentText"];
	    }
	}
	export class LRequestApproval {
	    approvedActionName: string;
	    approvedPlanHash: string;
	    consentText: string;
	
	    static createFrom(source: any = {}) {
	        return new LRequestApproval(source);
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
	
	export class LFileGenerated {
	    path: string;
	    lines: string[];
	
	    static createFrom(source: any = {}) {
	        return new LFileGenerated(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.lines = source["lines"];
	    }
	}
	export class LLibraryCompatibility {
	    supported: boolean;
	    available: boolean;
	    minVersion?: string;
	
	    static createFrom(source: any = {}) {
	        return new LLibraryCompatibility(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supported = source["supported"];
	        this.available = source["available"];
	        this.minVersion = source["minVersion"];
	    }
	}
	export class LLibraryPreparationStatus {
	    required: boolean;
	    kind?: string;
	    implemented: boolean;
	    implementation?: string;
	    implementationLanguage?: string;
	    reason?: string;
	
	    static createFrom(source: any = {}) {
	        return new LLibraryPreparationStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.required = source["required"];
	        this.kind = source["kind"];
	        this.implemented = source["implemented"];
	        this.implementation = source["implementation"];
	        this.implementationLanguage = source["implementationLanguage"];
	        this.reason = source["reason"];
	    }
	}
	export class LLibraryChoice {
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
	    supportState?: string;
	    preparationStatus?: LLibraryPreparationStatus;
	    unavailableReasons?: string[];
	    unavailableProfiles?: string[];
	    versionCompatibility?: LLibraryCompatibility;
	
	    static createFrom(source: any = {}) {
	        return new LLibraryChoice(source);
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
	        this.supportState = source["supportState"];
	        this.preparationStatus = this.convertValues(source["preparationStatus"], LLibraryPreparationStatus);
	        this.unavailableReasons = source["unavailableReasons"];
	        this.unavailableProfiles = source["unavailableProfiles"];
	        this.versionCompatibility = this.convertValues(source["versionCompatibility"], LLibraryCompatibility);
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
	
	export class LLibraryPatchEntry {
	    module: string;
	    libsLine: string;
	
	    static createFrom(source: any = {}) {
	        return new LLibraryPatchEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.module = source["module"];
	        this.libsLine = source["libsLine"];
	    }
	}
	export class LSourcePatch {
	    file: string;
	    find: string;
	    replace: string;
	
	    static createFrom(source: any = {}) {
	        return new LSourcePatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.find = source["find"];
	        this.replace = source["replace"];
	    }
	}
	export class LLibraryPreparation {
	    libraryId: string;
	    displayName: string;
	    trackName: string;
	    method: string;
	    buildSystem: string;
	    cFlags?: string[];
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
	    pkgConfigAppendCFlags?: string[];
	    pkgConfigLibsLine?: string;
	    pkgConfigLibsLinePatches?: LLibraryPatchEntry[];
	    privatePrefixInstall?: boolean;
	    verifyHeaderRelativePath: string;
	    verifyLibStem: string;
	    sourcePatches?: LSourcePatch[];
	    generatedSourceFiles?: LFileGenerated[];
	
	    static createFrom(source: any = {}) {
	        return new LLibraryPreparation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.libraryId = source["libraryId"];
	        this.displayName = source["displayName"];
	        this.trackName = source["trackName"];
	        this.method = source["method"];
	        this.buildSystem = source["buildSystem"];
	        this.cFlags = source["cFlags"];
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
	        this.pkgConfigAppendCFlags = source["pkgConfigAppendCFlags"];
	        this.pkgConfigLibsLine = source["pkgConfigLibsLine"];
	        this.pkgConfigLibsLinePatches = this.convertValues(source["pkgConfigLibsLinePatches"], LLibraryPatchEntry);
	        this.privatePrefixInstall = source["privatePrefixInstall"];
	        this.verifyHeaderRelativePath = source["verifyHeaderRelativePath"];
	        this.verifyLibStem = source["verifyLibStem"];
	        this.sourcePatches = this.convertValues(source["sourcePatches"], LSourcePatch);
	        this.generatedSourceFiles = this.convertValues(source["generatedSourceFiles"], LFileGenerated);
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
	
	export class LLibraryTrackSelection {
	    trackName: string;
	    libraries: LLibraryChoice[];
	
	    static createFrom(source: any = {}) {
	        return new LLibraryTrackSelection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.trackName = source["trackName"];
	        this.libraries = this.convertValues(source["libraries"], LLibraryChoice);
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
	export class LOperationPlan {
	    operationName: string;
	    summary: string;
	    summaryKey?: string;
	    summaryValues?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new LOperationPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.operationName = source["operationName"];
	        this.summary = source["summary"];
	        this.summaryKey = source["summaryKey"];
	        this.summaryValues = source["summaryValues"];
	    }
	}
	export class LOptionChoice {
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
	        return new LOptionChoice(source);
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
	export class LVersionLibraryWork {
	    workId: string;
	    ffmpegVersion: string;
	    libraryId: string;
	    goFilePath: string;
	    phaseNames: string[];
	    summary?: string;
	
	    static createFrom(source: any = {}) {
	        return new LVersionLibraryWork(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workId = source["workId"];
	        this.ffmpegVersion = source["ffmpegVersion"];
	        this.libraryId = source["libraryId"];
	        this.goFilePath = source["goFilePath"];
	        this.phaseNames = source["phaseNames"];
	        this.summary = source["summary"];
	    }
	}
	export class LResolvedBuildPlan {
	    ffmpegVersion: string;
	    requestedFfmpegVersion?: string;
	    compatibilityFfmpegVersion?: string;
	    ffmpegSourceArchiveUrl: string;
	    ffmpegSourceSignatureUrl: string;
	    selectedLibraries: LResolvedLibrary[];
	    versionLibraryWorks?: LVersionLibraryWork[];
	    requiredMsys2PackageNames?: string[];
	    configureFlags: string[];
	    warnings?: LWarningPlan[];
	
	    static createFrom(source: any = {}) {
	        return new LResolvedBuildPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ffmpegVersion = source["ffmpegVersion"];
	        this.requestedFfmpegVersion = source["requestedFfmpegVersion"];
	        this.compatibilityFfmpegVersion = source["compatibilityFfmpegVersion"];
	        this.ffmpegSourceArchiveUrl = source["ffmpegSourceArchiveUrl"];
	        this.ffmpegSourceSignatureUrl = source["ffmpegSourceSignatureUrl"];
	        this.selectedLibraries = this.convertValues(source["selectedLibraries"], LResolvedLibrary);
	        this.versionLibraryWorks = this.convertValues(source["versionLibraryWorks"], LVersionLibraryWork);
	        this.requiredMsys2PackageNames = source["requiredMsys2PackageNames"];
	        this.configureFlags = source["configureFlags"];
	        this.warnings = this.convertValues(source["warnings"], LWarningPlan);
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
	export class LResolvedLibrary {
	    libraryId: string;
	    displayName: string;
	    categoryName: string;
	    trackName: string;
	    supportState: string;
	    configureFlags: string[];
	    packageNames: string[];
	    officialWebpageUrl?: string;
	    licenseEffectName?: string;
	    plainExplanation?: string;
	    technicalExplanation?: string;
	    defaultChecked: boolean;
	    locked: boolean;
	    workIds?: string[];
	    preparationStatus?: LLibraryPreparationStatus;
	    unavailableReasons?: string[];
	    unavailableProfiles?: string[];
	    warnings?: LWarningPlan[];
	    versionCompatibility?: LLibraryCompatibility;
	
	    static createFrom(source: any = {}) {
	        return new LResolvedLibrary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.libraryId = source["libraryId"];
	        this.displayName = source["displayName"];
	        this.categoryName = source["categoryName"];
	        this.trackName = source["trackName"];
	        this.supportState = source["supportState"];
	        this.configureFlags = source["configureFlags"];
	        this.packageNames = source["packageNames"];
	        this.officialWebpageUrl = source["officialWebpageUrl"];
	        this.licenseEffectName = source["licenseEffectName"];
	        this.plainExplanation = source["plainExplanation"];
	        this.technicalExplanation = source["technicalExplanation"];
	        this.defaultChecked = source["defaultChecked"];
	        this.locked = source["locked"];
	        this.workIds = source["workIds"];
	        this.preparationStatus = this.convertValues(source["preparationStatus"], LLibraryPreparationStatus);
	        this.unavailableReasons = source["unavailableReasons"];
	        this.unavailableProfiles = source["unavailableProfiles"];
	        this.warnings = this.convertValues(source["warnings"], LWarningPlan);
	        this.versionCompatibility = this.convertValues(source["versionCompatibility"], LLibraryCompatibility);
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
	export class LResolvedVersionPlan {
	    ffmpegVersion: string;
	    requestedFfmpegVersion?: string;
	    compatibilityFfmpegVersion?: string;
	    visibleLibraries: LResolvedLibrary[];
	    hiddenLibraries?: LResolvedLibrary[];
	    unsupportedLibraries?: LResolvedLibrary[];
	    selectedLibraryIds: string[];
	    normalizedLibraryIds: string[];
	    requiredWorkIds?: string[];
	    configureFlags: string[];
	    requiredPackageNames?: string[];
	    warnings?: LWarningPlan[];
	
	    static createFrom(source: any = {}) {
	        return new LResolvedVersionPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ffmpegVersion = source["ffmpegVersion"];
	        this.requestedFfmpegVersion = source["requestedFfmpegVersion"];
	        this.compatibilityFfmpegVersion = source["compatibilityFfmpegVersion"];
	        this.visibleLibraries = this.convertValues(source["visibleLibraries"], LResolvedLibrary);
	        this.hiddenLibraries = this.convertValues(source["hiddenLibraries"], LResolvedLibrary);
	        this.unsupportedLibraries = this.convertValues(source["unsupportedLibraries"], LResolvedLibrary);
	        this.selectedLibraryIds = source["selectedLibraryIds"];
	        this.normalizedLibraryIds = source["normalizedLibraryIds"];
	        this.requiredWorkIds = source["requiredWorkIds"];
	        this.configureFlags = source["configureFlags"];
	        this.requiredPackageNames = source["requiredPackageNames"];
	        this.warnings = this.convertValues(source["warnings"], LWarningPlan);
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
	export class LWarningPlan {
	    riskLevelName: string;
	    message: string;
	    messageKey?: string;
	    messageValues?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new LWarningPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.riskLevelName = source["riskLevelName"];
	        this.message = source["message"];
	        this.messageKey = source["messageKey"];
	        this.messageValues = source["messageValues"];
	    }
	}
	export class LPlanFfmpeg {
	    actionName: string;
	    planHash: string;
	    workspaceDirectory: string;
	    msys2RootDirectory: string;
	    ffmpegSourceArchiveUrl: string;
	    ffmpegSourceSignatureUrl: string;
	    ffmpegSourceSha256Hash: string;
	    requestedFfmpegVersion?: string;
	    compatibilityFfmpegVersion?: string;
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
	    downloadConflictPolicyName: string;
	    extractionDestinationPolicyName: string;
	    operations: LOperationPlan[];
	    warnings: LWarningPlan[];
	    resolvedVersionPlan?: LResolvedVersionPlan;
	    resolvedBuildPlan?: LResolvedBuildPlan;
	    isExecutable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LPlanFfmpeg(source);
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
	        this.requestedFfmpegVersion = source["requestedFfmpegVersion"];
	        this.compatibilityFfmpegVersion = source["compatibilityFfmpegVersion"];
	        this.selectedLibraryIds = source["selectedLibraryIds"];
	        this.selectedLibraries = this.convertValues(source["selectedLibraries"], LLibraryChoice);
	        this.selectedNativeLibraries = this.convertValues(source["selectedNativeLibraries"], LLibraryChoice);
	        this.selectedInternalLibraries = this.convertValues(source["selectedInternalLibraries"], LLibraryChoice);
	        this.selectedExternalLibraries = this.convertValues(source["selectedExternalLibraries"], LLibraryChoice);
	        this.selectedLibrariesByTrack = this.convertValues(source["selectedLibrariesByTrack"], LLibraryTrackSelection);
	        this.libraryPreparations = this.convertValues(source["libraryPreparations"], LLibraryPreparation);
	        this.requiredMsys2PackageNames = source["requiredMsys2PackageNames"];
	        this.generatedConfigureFlags = source["generatedConfigureFlags"];
	        this.selectedConfigureOptions = this.convertValues(source["selectedConfigureOptions"], LOptionChoice);
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
	        this.operations = this.convertValues(source["operations"], LOperationPlan);
	        this.warnings = this.convertValues(source["warnings"], LWarningPlan);
	        this.resolvedVersionPlan = this.convertValues(source["resolvedVersionPlan"], LResolvedVersionPlan);
	        this.resolvedBuildPlan = this.convertValues(source["resolvedBuildPlan"], LResolvedBuildPlan);
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
	export class LPlanToolchain {
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
	    operations: LOperationPlan[];
	    warnings: LWarningPlan[];
	    isExecutable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LPlanToolchain(source);
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
	        this.operations = this.convertValues(source["operations"], LOperationPlan);
	        this.warnings = this.convertValues(source["warnings"], LWarningPlan);
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
	export class LPresetLibraryChoice {
	    presetId: string;
	    libraryIds: string[];
	    extendedLibraryIds?: string[];
	    hidden?: boolean;
	    dev?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LPresetLibraryChoice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.presetId = source["presetId"];
	        this.libraryIds = source["libraryIds"];
	        this.extendedLibraryIds = source["extendedLibraryIds"];
	        this.hidden = source["hidden"];
	        this.dev = source["dev"];
	    }
	}
	export class LReleaseChoice {
	    version: string;
	    codename: string;
	    archiveUrl: string;
	    signatureUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new LReleaseChoice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.codename = source["codename"];
	        this.archiveUrl = source["archiveUrl"];
	        this.signatureUrl = source["signatureUrl"];
	    }
	}
	
	
	
	export class LReviewFfmpeg {
	    reviewSessionId: string;
	    expectedLConsentText: string;
	    expectedLConsentTextHash: string;
	    expiresAtUnixTime: number;
	    plan: LPlanFfmpeg;
	
	    static createFrom(source: any = {}) {
	        return new LReviewFfmpeg(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reviewSessionId = source["reviewSessionId"];
	        this.expectedLConsentText = source["expectedLConsentText"];
	        this.expectedLConsentTextHash = source["expectedLConsentTextHash"];
	        this.expiresAtUnixTime = source["expiresAtUnixTime"];
	        this.plan = this.convertValues(source["plan"], LPlanFfmpeg);
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
	export class LReviewToolchain {
	    reviewSessionId: string;
	    expectedLConsentText: string;
	    expectedLConsentTextHash: string;
	    expiresAtUnixTime: number;
	    plan: LPlanToolchain;
	
	    static createFrom(source: any = {}) {
	        return new LReviewToolchain(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reviewSessionId = source["reviewSessionId"];
	        this.expectedLConsentText = source["expectedLConsentText"];
	        this.expectedLConsentTextHash = source["expectedLConsentTextHash"];
	        this.expiresAtUnixTime = source["expiresAtUnixTime"];
	        this.plan = this.convertValues(source["plan"], LPlanToolchain);
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
	export class LSettingsFfmpeg {
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
	        return new LSettingsFfmpeg(source);
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
	export class LSettingsToolchain {
	    workspaceDirectory: string;
	    msys2ArchiveUrl: string;
	    msys2ArchiveSha256Hash: string;
	    msys2ArchiveSignatureUrl: string;
	    msys2PackageNames: string[];
	    windowsShellProfileName: string;
	
	    static createFrom(source: any = {}) {
	        return new LSettingsToolchain(source);
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
	
	

}

export namespace program {
	
	export class LFileResult {
	    name: string;
	    path: string;
	    sizeBytes: number;
	    sha256Hash: string;
	
	    static createFrom(source: any = {}) {
	        return new LFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.sizeBytes = source["sizeBytes"];
	        this.sha256Hash = source["sha256Hash"];
	    }
	}
	export class LLogLocalEntry {
	    level: string;
	    message: string;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new LLogLocalEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.message = source["message"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class LRecordLog {
	    runId: string;
	    createdAt: string;
	    displayTime: string;
	    kind: string;
	    status: string;
	    directory: string;
	    entries: LLogLocalEntry[];
	    rawText: string;
	    errorCount: number;
	    warnCount: number;
	    hasStdoutLog: boolean;
	    hasStderrLog: boolean;
	    hasSecurityLAuditEvents: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LRecordLog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.createdAt = source["createdAt"];
	        this.displayTime = source["displayTime"];
	        this.kind = source["kind"];
	        this.status = source["status"];
	        this.directory = source["directory"];
	        this.entries = this.convertValues(source["entries"], LLogLocalEntry);
	        this.rawText = source["rawText"];
	        this.errorCount = source["errorCount"];
	        this.warnCount = source["warnCount"];
	        this.hasStdoutLog = source["hasStdoutLog"];
	        this.hasStderrLog = source["hasStderrLog"];
	        this.hasSecurityLAuditEvents = source["hasSecurityLAuditEvents"];
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
	export class LResultAction {
	    runId: string;
	    startedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new LResultAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.startedAt = source["startedAt"];
	    }
	}
	export class LResultState {
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
	
	    static createFrom(source: any = {}) {
	        return new LResultState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.artifactsDirectory = source["artifactsDirectory"];
	        this.reportPath = source["reportPath"];
	        this.ffmpegVersion = source["ffmpegVersion"];
	        this.files = this.convertValues(source["files"], LFileResult);
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
	export class LReviewFfmpegStored {
	    ReviewSession: reviewsession.LSessionReview;
	    Plan: planning.LPlanFfmpeg;
	
	    static createFrom(source: any = {}) {
	        return new LReviewFfmpegStored(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ReviewSession = this.convertValues(source["ReviewSession"], reviewsession.LSessionReview);
	        this.Plan = this.convertValues(source["Plan"], planning.LPlanFfmpeg);
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
	export class LReviewToolchainStored {
	    ReviewSession: reviewsession.LSessionReview;
	    Plan: planning.LPlanToolchain;
	
	    static createFrom(source: any = {}) {
	        return new LReviewToolchainStored(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ReviewSession = this.convertValues(source["ReviewSession"], reviewsession.LSessionReview);
	        this.Plan = this.convertValues(source["Plan"], planning.LPlanToolchain);
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
	export class LStateInitial {
	    hostOs: string;
	    kindExplanation: string;
	    securityRuleSummary: string;
	    namingRuleSummary: string;
	    defaultBuildConfigSettings: planning.LSettingsToolchain;
	    defaultFfmpegBuildSettings: planning.LSettingsFfmpeg;
	    defaultLibraryCatalog: planning.LLibraryChoice[];
	    defaultLibraryPresetCatalog: planning.LPresetLibraryChoice[];
	    defaultConfigureOptionCatalog: planning.LOptionChoice[];
	    supportedFfmpegReleases: planning.LReleaseChoice[];
	
	    static createFrom(source: any = {}) {
	        return new LStateInitial(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hostOs = source["hostOs"];
	        this.kindExplanation = source["kindExplanation"];
	        this.securityRuleSummary = source["securityRuleSummary"];
	        this.namingRuleSummary = source["namingRuleSummary"];
	        this.defaultBuildConfigSettings = this.convertValues(source["defaultBuildConfigSettings"], planning.LSettingsToolchain);
	        this.defaultFfmpegBuildSettings = this.convertValues(source["defaultFfmpegBuildSettings"], planning.LSettingsFfmpeg);
	        this.defaultLibraryCatalog = this.convertValues(source["defaultLibraryCatalog"], planning.LLibraryChoice);
	        this.defaultLibraryPresetCatalog = this.convertValues(source["defaultLibraryPresetCatalog"], planning.LPresetLibraryChoice);
	        this.defaultConfigureOptionCatalog = this.convertValues(source["defaultConfigureOptionCatalog"], planning.LOptionChoice);
	        this.supportedFfmpegReleases = this.convertValues(source["supportedFfmpegReleases"], planning.LReleaseChoice);
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
	export class LStatusToolchain {
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
	        return new LStatusToolchain(source);
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
	export class LVerificationLibrary {
	    libraryId: string;
	    displayName: string;
	    method: string;
	    expectedFlags: string[];
	    missingFlags: string[];
	    components: string[];
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new LVerificationLibrary(source);
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
	export class LVerificationState {
	    ffmpegPath: string;
	    ffmpegVersion: string;
	    libraries: LVerificationLibrary[];
	    unexpectedEnableFlags: string[];
	    okCount: number;
	    totalCount: number;
	    overall: string;
	    message: string;
	    verifiedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new LVerificationState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ffmpegPath = source["ffmpegPath"];
	        this.ffmpegVersion = source["ffmpegVersion"];
	        this.libraries = this.convertValues(source["libraries"], LVerificationLibrary);
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
	export class LVerificationToolchain {
	    verified: boolean;
	    checkedPackageCount: number;
	    missingPackageNames: string[];
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new LVerificationToolchain(source);
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

export namespace reviewsession {
	
	export class LSessionReview {
	    reviewSessionId: string;
	    actionName: string;
	    planHash: string;
	    expectedLConsentText: string;
	    expectedLConsentTextHash: string;
	    createdAtUnixTime: number;
	    expiresAtUnixTime: number;
	    wasUsed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LSessionReview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reviewSessionId = source["reviewSessionId"];
	        this.actionName = source["actionName"];
	        this.planHash = source["planHash"];
	        this.expectedLConsentText = source["expectedLConsentText"];
	        this.expectedLConsentTextHash = source["expectedLConsentTextHash"];
	        this.createdAtUnixTime = source["createdAtUnixTime"];
	        this.expiresAtUnixTime = source["expiresAtUnixTime"];
	        this.wasUsed = source["wasUsed"];
	    }
	}

}

export namespace workspace {
	
	export class LWorkspaceLayout {
	    workspaceDirectory: string;
	    cacheDirectory: string;
	    downloadsDirectory: string;
	    sourcesDirectory: string;
	    buildDirectory: string;
	    prefixDirectory: string;
	    artifactsBaseDirectory: string;
	    artifactsDirectory: string;
	    logsDirectory: string;
	    toolchainsDirectory: string;
	
	    static createFrom(source: any = {}) {
	        return new LWorkspaceLayout(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workspaceDirectory = source["workspaceDirectory"];
	        this.cacheDirectory = source["cacheDirectory"];
	        this.downloadsDirectory = source["downloadsDirectory"];
	        this.sourcesDirectory = source["sourcesDirectory"];
	        this.buildDirectory = source["buildDirectory"];
	        this.prefixDirectory = source["prefixDirectory"];
	        this.artifactsBaseDirectory = source["artifactsBaseDirectory"];
	        this.artifactsDirectory = source["artifactsDirectory"];
	        this.logsDirectory = source["logsDirectory"];
	        this.toolchainsDirectory = source["toolchainsDirectory"];
	    }
	}

}

