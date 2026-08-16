import { useEffect, useState } from "react";
import {
  LToolchainEnvironmentClear,
  LStatusToolchainGet,
  LToolchainProfileList,
  LToolchainInstallVerify,
} from "../wailsjs/go/program/LProgram";
import { LLocaleTextGet } from "./i18n";

// Owns the Prep tab toolchain status: the on-disk install status, the installed
// profiles, and the deep-verify result, plus the actions that read, verify, and
// clear them. Depends only on the workspace directory and shell profile passed in.
export function LStateToolchainStatusUse(dir: string, windowsShellProfileName: string) {
  const [toolchainStatus, setToolchainStatus] = useState<LStatusToolchain | null>(null);
  const [installedToolchainProfiles, setInstalledToolchainProfiles] = useState<LStatusToolchain[]>([]);
  const [toolchainVerification, setToolchainVerification] = useState<LVerificationToolchain | null>(null);
  const [isVerifyingToolchain, setIsVerifyingToolchain] = useState(false);

  // A changed workspace or shell profile invalidates any prior deep-verify result.
  useEffect(() => {
    setToolchainVerification(null);
  }, [dir, windowsShellProfileName]);

  async function refreshToolchainStatus() {
    if (!dir) { setToolchainStatus(null); setInstalledToolchainProfiles([]); return; }
    try { setToolchainStatus(await LStatusToolchainGet(dir, windowsShellProfileName)); }
    catch { setToolchainStatus(null); }
    try { setInstalledToolchainProfiles(await LToolchainProfileList(dir)); }
    catch { setInstalledToolchainProfiles([]); }
  }

  async function clearBuildEnvironments() {
    if (!dir) return;
    const confirmed = window.confirm(LLocaleTextGet("prep.profiles.clearConfirm"));
    if (!confirmed) return;
    setToolchainVerification(null);
    await LToolchainEnvironmentClear(dir);
    await refreshToolchainStatus();
  }

  async function verifyToolchain() {
    if (!dir) return;
    setIsVerifyingToolchain(true);
    setToolchainVerification(null);
    try { setToolchainVerification(await LToolchainInstallVerify(dir, windowsShellProfileName)); }
    catch (err) { setToolchainVerification({ verified: false, checkedPackageCount: 0, missingPackageNames: [], message: "", messageKey: "verify.toolchain.requestFailed", messageValues: { error: err instanceof Error ? err.message : String(err) } }); }
    finally { setIsVerifyingToolchain(false); }
  }

  return {
    toolchainStatus, installedToolchainProfiles, toolchainVerification, isVerifyingToolchain,
    refreshToolchainStatus, verifyToolchain, clearBuildEnvironments,
  };
}
