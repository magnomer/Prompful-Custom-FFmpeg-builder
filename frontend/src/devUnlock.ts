// Hidden developer feature, two tiers:
//
//   1. Basic dev unlock: clicking the version number in the About tab twelve times
//      toggles the UI-unavailable libraries (see uiUnavailableLibraryIds and
//      unimplementedBuildLibraryIds in libraries.tsx) between locked and selectable.
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
// Neither tier touches the catalog, recipes, or backend gating. Both are persisted in
// localStorage so they survive a reload during a dev session; the click counts are
// in-memory, so each twelve-click toggle must happen within one run. Sudo is gated on
// basic, and turning basic off clears sudo.

const DEV_UNLOCK_KEY = "ffbuilder.devUnlockUnavailable";
const SUDO_DEV_UNLOCK_KEY = "ffbuilder.sudoDevUnlock";
const REQUIRED_CLICKS = 12;

let versionClickCount = 0;
let devUnlockIndicatorClickCount = 0;

export function isDevUnlockEnabled(): boolean {
  try {
    return localStorage.getItem(DEV_UNLOCK_KEY) === "1";
  } catch {
    return false;
  }
}

// isSudoDevUnlockEnabled reports the sudo tier. It requires basic dev to be on, so a
// lingering sudo flag never takes effect once basic is locked again.
export function isSudoDevUnlockEnabled(): boolean {
  if (!isDevUnlockEnabled()) {
    return false;
  }
  try {
    return localStorage.getItem(SUDO_DEV_UNLOCK_KEY) === "1";
  } catch {
    return false;
  }
}

function setDevUnlockEnabled(enabled: boolean): boolean {
  try {
    if (enabled) {
      localStorage.setItem(DEV_UNLOCK_KEY, "1");
    } else {
      localStorage.removeItem(DEV_UNLOCK_KEY);
      // Basic off implies sudo off: clear it so re-enabling basic starts un-sudoed.
      localStorage.removeItem(SUDO_DEV_UNLOCK_KEY);
    }
  } catch {
    // localStorage unavailable: the in-memory UI state can still update for this
    // run, but the state simply does not persist. Not fatal.
  }
  return enabled;
}

function setSudoDevUnlockEnabled(enabled: boolean): boolean {
  try {
    if (enabled) {
      localStorage.setItem(SUDO_DEV_UNLOCK_KEY, "1");
    } else {
      localStorage.removeItem(SUDO_DEV_UNLOCK_KEY);
    }
  } catch {
    // localStorage unavailable: not fatal (see setDevUnlockEnabled).
  }
  return enabled && isDevUnlockEnabled();
}

// registerVersionClick records one click on the version number and returns the
// current basic developer-unlock state. Every twelve clicks toggle the state: locked
// -> unlocked, then unlocked -> locked, and so on.
export function registerVersionClick(): boolean {
  versionClickCount += 1;
  if (versionClickCount < REQUIRED_CLICKS) {
    return isDevUnlockEnabled();
  }

  versionClickCount = 0;
  return setDevUnlockEnabled(!isDevUnlockEnabled());
}

// registerDevUnlockIndicatorClick records one click on the "Developer mode ..."
// indicator and returns the current sudo state. Every twelve clicks toggle sudo, but
// only while basic dev is on; with basic off it is a no-op that stays un-sudoed.
export function registerDevUnlockIndicatorClick(): boolean {
  if (!isDevUnlockEnabled()) {
    devUnlockIndicatorClickCount = 0;
    return false;
  }
  devUnlockIndicatorClickCount += 1;
  if (devUnlockIndicatorClickCount < REQUIRED_CLICKS) {
    return isSudoDevUnlockEnabled();
  }

  devUnlockIndicatorClickCount = 0;
  return setSudoDevUnlockEnabled(!isSudoDevUnlockEnabled());
}
