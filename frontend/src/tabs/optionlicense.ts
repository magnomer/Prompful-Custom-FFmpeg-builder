import { LLocaleTextGet } from "../i18n";

export function LLicenseBoundaryResolve(licenseProfileName: string): string {
  switch (licenseProfileName) {
    case "gpl-local":
      return LLocaleTextGet("options.licenseBoundary.gpl");
    case "nonfree-local":
      return LLocaleTextGet("options.licenseBoundary.nonfree");
    case "lgpl-local":
    default:
      return LLocaleTextGet("options.licenseBoundary.lgpl");
  }
}

export function LLicenseShortGet(licenseProfileName: string): string {
  switch (licenseProfileName) {
    case "gpl-local":
      return LLocaleTextGet("options.summary.license.gpl-local");
    case "nonfree-local":
      return LLocaleTextGet("options.summary.license.nonfree-local");
    case "lgpl-local":
    default:
      return LLocaleTextGet("options.summary.license.lgpl-local");
  }
}

export function LLicenseBoundaryNormalize(licenseProfileName: string): string {
  switch (licenseProfileName) {
    case "gpl-local":
    case "nonfree-local":
      return licenseProfileName;
    default:
      return "lgpl-local";
  }
}
