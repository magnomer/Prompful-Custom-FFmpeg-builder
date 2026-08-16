import { useEffect, useRef, useState } from "react";
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
  const identity = `${dir}\u0000${windowsShellProfileName}`;
  const currentIdentityRef = useRef(identity);
  const refreshRequestRef = useRef(0);
  const verificationRequestRef = useRef(0);
  currentIdentityRef.current = identity;

  // A changed workspace or shell profile invalidates any prior deep-verify result.
  useEffect(() => {
    refreshRequestRef.current += 1;
    verificationRequestRef.current += 1;
    setToolchainStatus(null);
    setToolchainVerification(null);
    setIsVerifyingToolchain(false);
  }, [dir, windowsShellProfileName]);

  useEffect(() => {
    setInstalledToolchainProfiles([]);
  }, [dir]);

  async function refreshToolchainStatus() {
    const requestDirectory = dir;
    const requestProfile = windowsShellProfileName;
    const requestIdentity = `${requestDirectory}\u0000${requestProfile}`;
    const requestId = ++refreshRequestRef.current;
    if (!requestDirectory) { setToolchainStatus(null); setInstalledToolchainProfiles([]); return; }
    let status: LStatusToolchain | null = null;
    let profiles: LStatusToolchain[] = [];
    try { status = await LStatusToolchainGet(requestDirectory, requestProfile); }
    catch { status = null; }
    try { profiles = await LToolchainProfileList(requestDirectory); }
    catch { profiles = []; }
    if (requestId !== refreshRequestRef.current || currentIdentityRef.current !== requestIdentity) return;
    setToolchainStatus(status);
    setInstalledToolchainProfiles(profiles);
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
    const requestDirectory = dir;
    const requestProfile = windowsShellProfileName;
    const requestIdentity = `${requestDirectory}\u0000${requestProfile}`;
    const requestId = ++verificationRequestRef.current;
    if (!requestDirectory) return;
    setIsVerifyingToolchain(true);
    setToolchainVerification(null);
    try {
      const verification = await LToolchainInstallVerify(requestDirectory, requestProfile);
      if (requestId !== verificationRequestRef.current || currentIdentityRef.current !== requestIdentity) return;
      setToolchainVerification(verification);
    }
    catch (err) {
      if (requestId !== verificationRequestRef.current || currentIdentityRef.current !== requestIdentity) return;
      setToolchainVerification({ verified: false, checkedPackageCount: 0, missingPackageNames: [], message: "", messageKey: "verify.toolchain.requestFailed", messageValues: { error: err instanceof Error ? err.message : String(err) } });
    }
    finally {
      if (requestId === verificationRequestRef.current && currentIdentityRef.current === requestIdentity) setIsVerifyingToolchain(false);
    }
  }

  return {
    toolchainStatus, installedToolchainProfiles, toolchainVerification, isVerifyingToolchain,
    refreshToolchainStatus, verifyToolchain, clearBuildEnvironments,
  };
}
