import en from "../../shared/localization/en.json";
import ko from "../../shared/localization/ko.json";

type LDictionary = Record<string, string>;
export type LLocaleCode = "en" | "ko";

const LLocaleEnglishDictionary: LDictionary = en as LDictionary;
const LLocaleKoreanDictionary: LDictionary = ko as LDictionary;
const LLocaleDictionaryMap: Record<LLocaleCode, LDictionary> = { en: LLocaleEnglishDictionary, ko: LLocaleKoreanDictionary };
const LLocaleStorageKey = "customffmpeg.locale";

function LLocaleNormalize(locale: string | null): LLocaleCode {
  return locale === "ko" ? "ko" : "en";
}

let LLocaleCurrent: LLocaleCode = LLocaleNormalize(globalThis.localStorage?.getItem(LLocaleStorageKey) ?? null);

export function LLocaleGet(): LLocaleCode {
  return LLocaleCurrent;
}

export function LLocaleSet(locale: LLocaleCode): void {
  LLocaleCurrent = LLocaleNormalize(locale);
  globalThis.localStorage?.setItem(LLocaleStorageKey, LLocaleCurrent);
  globalThis.dispatchEvent(new CustomEvent("customffmpeg-locale-change", { detail: LLocaleCurrent }));
}

export function LLocaleTranslationCheck(id: string): boolean {
  return Object.prototype.hasOwnProperty.call(LLocaleDictionaryMap[LLocaleCurrent], id) || Object.prototype.hasOwnProperty.call(LLocaleEnglishDictionary, id);
}

export function LLocaleTextGet(id: string, values: Record<string, string | number> = {}): string {
  const template = LLocaleDictionaryMap[LLocaleCurrent][id] ?? LLocaleEnglishDictionary[id] ?? id;
  return template.replace(/\{(\w+)\}/g, (_, key: string) => String(values[key] ?? `{${key}}`));
}

export function LLocaleFallbackGet(id: string, fallback: string, values: Record<string, string | number> = {}): string {
  if (!LLocaleTranslationCheck(id)) return fallback;
  return LLocaleTextGet(id, values);
}

export function LLocaleStatusGet(status: string): string {
  return LLocaleFallbackGet(`status.${status}`, status);
}
