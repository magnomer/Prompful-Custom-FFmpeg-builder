import en from "../../localization/en.json";
import ko from "../../localization/ko.json";
import { LLocaleSet as LLocaleBackendSet } from "../wailsjs/go/program/LProgram";

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

export async function LLocaleSet(locale: LLocaleCode): Promise<void> {
  const normalizedLocale = LLocaleNormalize(locale);
  // Commit the visible locale only after the backend has accepted the same
  // value. Locale-aware dialogs and verification calls can therefore never
  // observe an older locale than the UI that launched them.
  await LLocaleBackendSet(normalizedLocale);
  LLocaleCurrent = normalizedLocale;
  globalThis.localStorage?.setItem(LLocaleStorageKey, LLocaleCurrent);
  globalThis.dispatchEvent(new CustomEvent("customffmpeg-locale-change", { detail: LLocaleCurrent }));
}

export function LLocaleSynchronize(): Promise<void> {
  return LLocaleBackendSet(LLocaleCurrent);
}

export function LLocaleTranslationCheck(id: string): boolean {
  return Object.prototype.hasOwnProperty.call(LLocaleDictionaryMap[LLocaleCurrent], id) || Object.prototype.hasOwnProperty.call(LLocaleEnglishDictionary, id);
}

export function LLocaleTextGet(id: string, values: Record<string, string | number> = {}): string {
  const template = LLocaleDictionaryMap[LLocaleCurrent][id] ?? LLocaleEnglishDictionary[id] ?? id;
  return template.replace(/\{(\w+)\}/g, (_, key: string) => String(values[key] ?? `{${key}}`));
}

export function LLocaleTextForGet(locale: LLocaleCode, id: string, values: Record<string, string | number> = {}): string {
  const template = LLocaleDictionaryMap[locale][id] ?? LLocaleEnglishDictionary[id] ?? id;
  return template.replace(/\{(\w+)\}/g, (_, key: string) => String(values[key] ?? `{${key}}`));
}

export function LLocaleFallbackGet(id: string, fallback: string, values: Record<string, string | number> = {}): string {
  if (!LLocaleTranslationCheck(id)) return fallback;
  return LLocaleTextGet(id, values);
}

export type LLocalizedMessage = {
  messageKey?: string;
  messageValues?: Record<string, string | number>;
  message?: string;
};

export function LLocaleMessageGet(message: LLocalizedMessage | null | undefined): string {
  if (!message) return "";
  if (message.messageKey) return LLocaleFallbackGet(message.messageKey, message.message ?? message.messageKey, message.messageValues ?? {});
  return message.message ?? "";
}

export function LLocaleStatusGet(status: string): string {
  return LLocaleFallbackGet(`status.${status}`, status);
}
