// Hidden developer feature, two tiers:
//
//   1. Basic dev unlock: clicking the version number in the About tab twelve times
//      toggles the UI-unavailable libraries (see uiUnavailableLibraryIds and
//      LLibraryBuildUnimplementedIds in libraries.tsx) between locked and selectable.
//      It ONLY relaxes availability; the normal-mode mutual-exclusion rules (one TLS
//      backend, shaderc-over-glslang, one EVC profile binding) stay enforced so a
//      basic-dev selection is still buildable.
//
//   2. Sudo dev unlock: with basic dev already on, clicking the "Developer mode ..."
//      indicator twelve times toggles sudo mode, which additionally relaxes the
//      pick-one mutual-exclusion groups (e.g. every TLS backend selectable at once)
//      for UI testing. The backend still validates and blocks an unbuildable mix, so
//      this is a UI-exploration aid, not a way to ship a conflicting build.
//
// Neither tier touches the library catalog, recipes, or backend gating. Both are persisted in
// localStorage so they survive a reload during a dev session; the click counts are
// in-memory, so each twelve-click toggle must happen within one run. Sudo is gated on
// basic, and turning basic off clears sudo.

const LUnlockKeyBasic = "ffbuilder.devUnlockUnavailable";
const LUnlockKeySudo = "ffbuilder.sudoDevUnlock";
const LUnlockClickTarget = 12;

let LUnlockVersionCount = 0;
let LUnlockIndicatorCount = 0;

export function LUnlockBasicCheck(): boolean {
  try {
    return localStorage.getItem(LUnlockKeyBasic) === "1";
  } catch {
    return false;
  }
}

// LUnlockSudoCheck reports the sudo tier. It requires basic dev to be on, so a
// lingering sudo flag never takes effect once basic is locked again.
export function LUnlockSudoCheck(): boolean {
  if (!LUnlockBasicCheck()) {
    return false;
  }
  try {
    return localStorage.getItem(LUnlockKeySudo) === "1";
  } catch {
    return false;
  }
}

function LUnlockBasicSet(enabled: boolean): boolean {
  try {
    if (enabled) {
      localStorage.setItem(LUnlockKeyBasic, "1");
    } else {
      localStorage.removeItem(LUnlockKeyBasic);
      // Basic off implies sudo off: clear it so re-enabling basic starts un-sudoed.
      localStorage.removeItem(LUnlockKeySudo);
    }
  } catch {
    // localStorage unavailable: the in-memory UI state can still update for this
    // run, but the state simply does not persist. Not fatal.
  }
  return enabled;
}

function LUnlockSudoSet(enabled: boolean): boolean {
  try {
    if (enabled) {
      localStorage.setItem(LUnlockKeySudo, "1");
    } else {
      localStorage.removeItem(LUnlockKeySudo);
    }
  } catch {
    // localStorage unavailable: not fatal (see LUnlockBasicSet).
  }
  return enabled && LUnlockBasicCheck();
}

// LUnlockVersionAdd records one click on the version number and returns the
// current basic developer-unlock state. Every twelve clicks toggle the state: locked
// -> unlocked, then unlocked -> locked, and so on.
export function LUnlockVersionAdd(): boolean {
  LUnlockVersionCount += 1;
  if (LUnlockVersionCount < LUnlockClickTarget) {
    return LUnlockBasicCheck();
  }

  LUnlockVersionCount = 0;
  return LUnlockBasicSet(!LUnlockBasicCheck());
}

// LUnlockIndicatorAdd records one click on the "Developer mode ..."
// indicator and returns the current sudo state. Every twelve clicks toggle sudo, but
// only while basic dev is on; with basic off it is a no-op that stays un-sudoed.
export function LUnlockIndicatorAdd(): boolean {
  if (!LUnlockBasicCheck()) {
    LUnlockIndicatorCount = 0;
    return false;
  }
  LUnlockIndicatorCount += 1;
  if (LUnlockIndicatorCount < LUnlockClickTarget) {
    return LUnlockSudoCheck();
  }

  LUnlockIndicatorCount = 0;
  return LUnlockSudoSet(!LUnlockSudoCheck());
}
