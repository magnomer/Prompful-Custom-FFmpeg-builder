import { useEffect, useRef, useState } from "react";
import {
  LResultBuildGet,
  LVerificationBuildRun,
  LDirectoryResultOpen,
  LReportResultOpen,
} from "../wailsjs/go/program/LProgram";
import { LLocaleTextGet, type LLocalizedMessage } from "./i18n";

// Owns the Result tab state: the loaded build result, its deep-verify result,
// and the actions that read, verify, and open the build artifacts. All actions
// derive from the workspace directory passed in by the builder hook.
export function LStateResultUse(dir: string) {
  const [buildResult, setBuildResult] = useState<LResultState | null>(null);
  const [buildResultError, setBuildResultError] = useState("");
  const [isLoadingBuildResult, setIsLoadingBuildResult] = useState(false);
  const [buildVerification, setBuildVerification] = useState<LVerificationState | null>(null);
  const [buildVerificationError, setBuildVerificationError] = useState<LLocalizedMessage | null>(null);
  const [isVerifyingBuild, setIsVerifyingBuild] = useState(false);
  const currentDirectoryRef = useRef(dir);
  const refreshRequestRef = useRef(0);
  const verificationRequestRef = useRef(0);
  currentDirectoryRef.current = dir;

  useEffect(() => {
    refreshRequestRef.current += 1;
    verificationRequestRef.current += 1;
    setBuildResult(null);
    setBuildResultError("");
    setBuildVerification(null);
    setBuildVerificationError(null);
    setIsLoadingBuildResult(false);
    setIsVerifyingBuild(false);
  }, [dir]);

  function LResultClear() {
    refreshRequestRef.current += 1;
    verificationRequestRef.current += 1;
    setBuildResult(null); setBuildResultError("");
    setBuildVerification(null); setBuildVerificationError(null);
    setIsLoadingBuildResult(false); setIsVerifyingBuild(false);
  }

  async function refreshBuildResult() {
    const requestDirectory = dir;
    const requestId = ++refreshRequestRef.current;
    if (!requestDirectory) { setBuildResult(null); setBuildResultError(LLocaleTextGet("result.error.chooseWorkspaceFirst")); return; }
    setBuildVerification(null); setBuildVerificationError(null);
    setIsLoadingBuildResult(true);
    try {
      const r = await LResultBuildGet(requestDirectory);
      if (requestId !== refreshRequestRef.current || currentDirectoryRef.current !== requestDirectory) return;
      setBuildResult(r); setBuildResultError("");
    }
    catch (err) {
      if (requestId !== refreshRequestRef.current || currentDirectoryRef.current !== requestDirectory) return;
      setBuildResult(null); setBuildResultError(err instanceof Error ? err.message : String(err));
    }
    finally {
      if (requestId === refreshRequestRef.current && currentDirectoryRef.current === requestDirectory) setIsLoadingBuildResult(false);
    }
  }

  async function verifyBuildResult() {
    const requestDirectory = dir;
    const requestId = ++verificationRequestRef.current;
    if (!requestDirectory) { setBuildVerificationError({ messageKey: "result.error.chooseWorkspaceFirst" }); return; }
    setIsVerifyingBuild(true);
    setBuildVerificationError(null);
    try {
      const verification = await LVerificationBuildRun(requestDirectory);
      if (requestId !== verificationRequestRef.current || currentDirectoryRef.current !== requestDirectory) return;
      setBuildVerification(verification);
    }
    catch (err) {
      if (requestId !== verificationRequestRef.current || currentDirectoryRef.current !== requestDirectory) return;
      setBuildVerification(null); setBuildVerificationError({ messageKey: "verify.build.requestFailed", messageValues: { error: err instanceof Error ? err.message : String(err) } });
    }
    finally {
      if (requestId === verificationRequestRef.current && currentDirectoryRef.current === requestDirectory) setIsVerifyingBuild(false);
    }
  }

  async function openResultFolder() {
    if (!dir) { setBuildResultError(LLocaleTextGet("result.error.chooseWorkspaceFirst")); return; }
    if (!buildResult?.artifactsDirectory) return;
    try { await LDirectoryResultOpen(dir, buildResult.artifactsDirectory); setBuildResultError(""); }
    catch (err) { setBuildResultError(err instanceof Error ? err.message : String(err)); }
  }

  async function openResultReport() {
    if (!dir) { setBuildResultError(LLocaleTextGet("result.error.chooseWorkspaceFirst")); return; }
    if (!buildResult?.reportPath) return;
    try { await LReportResultOpen(dir, buildResult.reportPath); setBuildResultError(""); }
    catch (err) { setBuildResultError(err instanceof Error ? err.message : String(err)); }
  }

  return {
    buildResult, buildResultError, isLoadingBuildResult,
    buildVerification, buildVerificationError, isVerifyingBuild,
    LResultClear,
    refreshBuildResult, verifyBuildResult, openResultFolder, openResultReport,
  };
}
