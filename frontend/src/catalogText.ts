import { t, tFallback } from "./i18n";

export function libraryText(library: LibraryChoice, field: "displayName" | "categoryName" | "plainExplanation" | "technicalExplanation"): string {
  return tFallback(`catalog.libraries.${library.libraryId}.${field}`, library[field] ?? "");
}

export function libraryLicenseLabel(licenseEffectName: string): string {
  return tFallback(`libraries.row.license.${licenseEffectName}`, licenseEffectName);
}

export function configureOptionText(option: ConfigureOptionChoice, field: "displayName" | "categoryName" | "plainExplanation" | "technicalNote"): string {
  return tFallback(`catalog.options.${option.optionId}.${field}`, option[field] ?? "");
}

export function catalogOptionName(optionId: string): string {
  return t(`catalog.options.${optionId}.displayName`);
}
