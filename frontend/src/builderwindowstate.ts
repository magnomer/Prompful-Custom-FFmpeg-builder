import { useEffect } from "react";
import { WindowGetPosition, WindowGetSize, WindowSetPosition, WindowSetSize } from "../wailsjs/runtime/runtime";
import { LStateWindowSaved, LStateWindowKey, LStateWindowRead } from "./programstate";

function LRuntimeNumberRead(value: unknown, tupleIndex: number, primary: string, fallback: string): number {
  if (Array.isArray(value)) return Number(value[tupleIndex]);
  if (value && typeof value === "object") {
    const r = value as Record<string, unknown>;
    return Number(r[primary] ?? r[fallback] ?? 0);
  }
  return 0;
}

export function LStateWindowRestore() {
  const s = LStateWindowRead();
  if (Number.isFinite(s.width) && Number.isFinite(s.height)) WindowSetSize(Number(s.width), Number(s.height));
  if (Number.isFinite(s.x) && Number.isFinite(s.y)) WindowSetPosition(Number(s.x), Number(s.y));
}

export async function LStateWindowSave() {
  try {
    const sz = await WindowGetSize();
    const pos = await WindowGetPosition();
    const width  = LRuntimeNumberRead(sz,  0, "w", "width");
    const height = LRuntimeNumberRead(sz,  1, "h", "height");
    const x = LRuntimeNumberRead(pos, 0, "x", "left");
    const y = LRuntimeNumberRead(pos, 1, "y", "top");
    window.localStorage.setItem(LStateWindowKey, JSON.stringify({ width, height, x, y } satisfies LStateWindowSaved));
  } catch {
    // Window persistence is best-effort. Never block the program if the runtime is unavailable.
  }
}

// Persist window geometry periodically and on unload. Restore happens separately,
// after the saved program state has loaded.
export function LStateWindowUse() {
  useEffect(() => {
    const id = window.setInterval(() => { LStateWindowSave(); }, 2000);
    window.addEventListener("beforeunload", LStateWindowSave);
    return () => { window.clearInterval(id); window.removeEventListener("beforeunload", LStateWindowSave); LStateWindowSave(); };
  }, []);
}
