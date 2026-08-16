import type React from "react";

// Implements the horizontal ARIA tabs keyboard model and moves focus together
// with selection so the tablist remains a single stop in the document tab order.
export function LTabKeyDown(
  event: React.KeyboardEvent<HTMLButtonElement>,
  currentIndex: number,
  tabCount: number,
  onSelect: (nextIndex: number) => void,
) {
  if (tabCount < 1) return;
  let nextIndex: number | null = null;
  if (event.key === "ArrowRight") nextIndex = (currentIndex + 1) % tabCount;
  if (event.key === "ArrowLeft") nextIndex = (currentIndex - 1 + tabCount) % tabCount;
  if (event.key === "Home") nextIndex = 0;
  if (event.key === "End") nextIndex = tabCount - 1;
  if (nextIndex === null) return;

  event.preventDefault();
  onSelect(nextIndex);
  const tabButtons = event.currentTarget.parentElement?.querySelectorAll<HTMLButtonElement>('[role="tab"]');
  tabButtons?.item(nextIndex).focus();
}
