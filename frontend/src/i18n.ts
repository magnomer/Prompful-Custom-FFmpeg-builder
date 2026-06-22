import en from "../../shared/localization/en.json";
import ko from "../../shared/localization/ko.json";

type Dictionary = Record<string, string>;
export type LocaleCode = "en" | "ko";

const enDictionary: Dictionary = en as Dictionary;
const koDictionary: Dictionary = ko as Dictionary;
const dictionaries: Record<LocaleCode, Dictionary> = { en: enDictionary, ko: koDictionary };
const localeStorageKey = "customffmpeg.locale";

function normalizeLocale(locale: string | null): LocaleCode {
  return locale === "ko" ? "ko" : "en";
}

let currentLocale: LocaleCode = normalizeLocale(globalThis.localStorage?.getItem(localeStorageKey) ?? null);

export function getLocale(): LocaleCode {
  return currentLocale;
}

export function setLocale(locale: LocaleCode): void {
  currentLocale = normalizeLocale(locale);
  globalThis.localStorage?.setItem(localeStorageKey, currentLocale);
  globalThis.dispatchEvent(new CustomEvent("customffmpeg-locale-change", { detail: currentLocale }));
}

export function hasTranslation(id: string): boolean {
  return Object.prototype.hasOwnProperty.call(dictionaries[currentLocale], id) || Object.prototype.hasOwnProperty.call(enDictionary, id);
}

export function t(id: string, values: Record<string, string | number> = {}): string {
  const template = dictionaries[currentLocale][id] ?? enDictionary[id] ?? id;
  return template.replace(/\{(\w+)\}/g, (_, key: string) => String(values[key] ?? `{${key}}`));
}

export function tFallback(id: string, fallback: string, values: Record<string, string | number> = {}): string {
  if (!hasTranslation(id)) return fallback;
  return t(id, values);
}

export function tStatus(status: string): string {
  return tFallback(`status.${status}`, status);
}
