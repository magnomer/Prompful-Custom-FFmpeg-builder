import { LLocaleTextGet } from "../i18n";
import { LLibraryTextGet } from "../catalogText";
import { LUnlockBasicCheck } from "../devUnlock";

// A category-filter selection value is the category name string, kept as a distinct
// alias so the section-filter code reads intent rather than a bare string.
export type LSectionState = string;

// Library availability is now driven by the backend resolved catalog. The frontend
// must not carry hardcoded unimplemented-library or profile-unavailable lists: those
// are version/profile facts in internal/catalogfacts/catalogdata/libraries/*.json.
export function LLibraryReasonCheck(library: LLibraryChoice, reason: string): boolean {
  return (library.unavailableReasons ?? []).includes(reason);
}

export function LLibraryDisabledCheck(library: LLibraryChoice): boolean {
  return LLibraryReasonCheck(library, "disabled-in-current-catalog-ui") || library.supportState === "ui-disabled";
}

export function LLibraryPrepCheck(library: LLibraryChoice): boolean {
  return Boolean(library.preparationStatus?.required) && library.preparationStatus?.implemented !== true;
}

export function LLibraryProfileCheck(library: LLibraryChoice, windowsShellProfileName: string): boolean {
  return !(library.unavailableProfiles ?? []).includes(windowsShellProfileName);
}

export function LLibrarySupportedCheck(library: LLibraryChoice): boolean {
  return library.versionCompatibility?.supported !== false;
}

// Available means FFmpeg has the switch for the chosen release AND the package this builder
// can supply can satisfy it on that release. It is false for a release-support-unavailable row
// (e.g. lensfun, whose package FFmpeg cannot use) even though the switch nominally exists, so
// it is a stricter gate than LLibrarySupportedCheck. Both are per-FFmpeg-version,
// driven by the backend release-support manifest annotation, never a global list.
export function LLibraryAvailableCheck(library: LLibraryChoice): boolean {
  return library.versionCompatibility?.available !== false;
}

export function LLibraryTrackGet(trackName: string): string {
  return LLocaleTextGet(`libraries.row.track.${trackName || "native"}`);
}

export function LLibraryVisibleFilter(LCatalogLibrarySource: LLibraryChoice[], windowsShellProfileName: string): LLibraryChoice[] {
  void windowsShellProfileName;
  return LCatalogLibrarySource;
}

export function LLibraryUnavailableCheck(library: LLibraryChoice, windowsShellProfileName: string): boolean {
  return LLibraryDisabledCheck(library) || LLibraryPrepCheck(library) || !LLibraryProfileCheck(library, windowsShellProfileName) || !LLibrarySupportedCheck(library) || !LLibraryAvailableCheck(library);
}

// Whether the checkbox is actually locked. Same as UI-unavailable, except the hidden
// About-tab developer unlock makes otherwise-unavailable libraries checkable for testing.
// The unavailable styling still shows regardless of unlock.
export function LLibraryLockedCheck(library: LLibraryChoice, windowsShellProfileName: string): boolean {
  if (!LLibrarySupportedCheck(library)) return true;
  if (!LLibraryAvailableCheck(library)) return true;
  return LLibraryUnavailableCheck(library, windowsShellProfileName) && !LUnlockBasicCheck();
}

// Returns the localization key suffix for an unavailable row's note, or "" when no
// note should be shown. The reason comes from the backend-resolved catalog state; the
// frontend does not own library-specific build-preparation facts.
export function LLibraryReasonGet(library: LLibraryChoice, windowsShellProfileName: string): string {
  const libraryId = library.libraryId;
  if (!LLibrarySupportedCheck(library)) return "ffmpegVersionUnsupported";
  if (!LLibraryProfileCheck(library, windowsShellProfileName)) return "profileUnavailable";
  if (LLibraryPrepCheck(library)) return "preparationUnimplemented";
  if (LLibraryDisabledCheck(library)) return libraryId;
  if (!LLibraryAvailableCheck(library)) return libraryId;
  return "";
}

export function LLibraryCategoryGet(library: LLibraryChoice): string {
  return LLibraryTextGet(library, "categoryName") || LLocaleTextGet("common.other");
}

export function LLibraryCategoryCheck(categoryName: string): boolean {
  const normalizedCategoryName = categoryName.toLocaleLowerCase();
  return normalizedCategoryName.includes("included by default") || normalizedCategoryName.includes("기본 포함");
}

export function LSectionLabelGet(categoryName: LSectionState): string {
  if (LLibraryCategoryCheck(categoryName)) return LLocaleTextGet("libraries.categoryFilter.default");
  return categoryName;
}

export function LSectionSummaryGet(selectedSections: LSectionState[]): string {
  if (selectedSections.length === 0) return LLocaleTextGet("libraries.categoryFilter.all");
  if (selectedSections.length === 1) return LSectionLabelGet(selectedSections[0]);
  return LLocaleTextGet("libraries.categoryFilter.selectedCount").replace("{count}", String(selectedSections.length));
}

export function LSectionOptionsGet(LCatalogLibrarySource: LLibraryChoice[], windowsShellProfileName: string): string[] {
  const categoryNames: string[] = [];
  for (const library of LLibraryVisibleFilter(LCatalogLibrarySource, windowsShellProfileName)) {
    const categoryName = LLibraryCategoryGet(library);
    if (!categoryNames.includes(categoryName)) categoryNames.push(categoryName);
  }
  return categoryNames;
}

export function LLibraryFilter(LCatalogLibrarySource: LLibraryChoice[], windowsShellProfileName: string, searchQuery: string, sectionFilters: LSectionState[]): LLibraryChoice[] {
  const normalizedQuery = searchQuery.trim().toLocaleLowerCase();
  return LLibraryVisibleFilter(LCatalogLibrarySource, windowsShellProfileName).filter((library) => {
    const categoryName = LLibraryCategoryGet(library);
    if (sectionFilters.length > 0 && !sectionFilters.includes(categoryName)) return false;
    if (!normalizedQuery) return true;
    const searchableText = [
      library.libraryId,
      LLibraryTextGet(library, "displayName"),
      categoryName,
      LLibraryTextGet(library, "plainExplanation"),
      LLibraryTextGet(library, "technicalExplanation"),
      library.configureFlags.join(" "),
      library.packageNames.join(" "),
      library.officialWebpageUrl || "",
    ].join(" ").toLocaleLowerCase();
    return searchableText.includes(normalizedQuery);
  });
}

export function LLibraryNameSplit(displayName: string): { buildName: string; featureName: string } {
  const separator = " / ";
  if (!displayName.includes(separator)) {
    return { buildName: displayName, featureName: "" };
  }
  const [buildName, ...featureParts] = displayName.split(separator);
  return { buildName: buildName.trim(), featureName: LLibraryFeatureClean(featureParts.join(separator).trim()) };
}

function LLibraryFeatureClean(featureName: string): string {
  const suffixes = [
    " encoding",
    " decoding",
    " support",
    " input",
    " output",
    " 인코딩",
    " 디코딩",
    " 지원",
    " 입력",
    " 출력",
  ];
  for (const suffix of suffixes) {
    if (featureName.toLocaleLowerCase().endsWith(suffix.toLocaleLowerCase())) {
      return featureName.slice(0, -suffix.length).trim();
    }
  }
  return featureName;
}

export function LLibraryFilterCreate(libraries: LLibraryChoice[]) {
  return libraries.reduce<Record<string, LLibraryChoice[]>>((groups, library) => {
    const categoryName = LLibraryCategoryGet(library);
    groups[categoryName] = groups[categoryName] || [];
    groups[categoryName].push(library);
    return groups;
  }, {});
}
