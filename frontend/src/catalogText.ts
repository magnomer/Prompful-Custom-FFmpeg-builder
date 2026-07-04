import { LLocaleTextGet, LLocaleFallbackGet } from "./i18n";

export function LLibraryTextGet(library: LLibraryChoice, field: "displayName" | "categoryName" | "plainExplanation" | "technicalExplanation"): string {
  return LLocaleFallbackGet(`catalog.libraries.${library.libraryId}.${field}`, library[field] ?? "");
}

export function LLicenseLabelGet(licenseEffectName: string): string {
  return LLocaleFallbackGet(`libraries.row.license.${licenseEffectName}`, licenseEffectName);
}

export function LOptionTextGet(option: LOptionChoice, field: "displayName" | "categoryName" | "plainExplanation" | "technicalNote"): string {
  return LLocaleFallbackGet(`catalog.options.${option.optionId}.${field}`, option[field] ?? "");
}

export function LOptionNameGet(optionId: string): string {
  return LLocaleTextGet(`catalog.options.${optionId}.displayName`);
}
