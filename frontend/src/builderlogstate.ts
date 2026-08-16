import { useEffect, useRef, useState } from "react";
import {
  LRecordLogGet,
  LLogRecordList,
  LFolderRecordOpen,
  LFileRecordOpen,
  LFolderLogOpen,
} from "../wailsjs/go/program/LProgram";
import { LLogLevelNormalize } from "./programstate";

// Owns the Logs tab state: the local run records read from the workspace and the
// actions that list, load, and open them. The workspace directory is supplied by
// the builder hook.
export function LStateLogUse(dir: string) {
  const [localLogRecords, setLocalLogRecords] = useState<LRecordLog[]>([]);
  const [localLogRecordsError, setLocalLogRecordsError] = useState("");
  const currentDirectoryRef = useRef(dir);
  const refreshRequestRef = useRef(0);
  const detailRequestRef = useRef(new Map<string, number>());
  currentDirectoryRef.current = dir;

  useEffect(() => {
    refreshRequestRef.current += 1;
    detailRequestRef.current.clear();
    setLocalLogRecords([]);
    setLocalLogRecordsError("");
  }, [dir]);

  async function refreshLocalLogRecords() {
    const requestDirectory = dir;
    const requestId = ++refreshRequestRef.current;
    if (!requestDirectory) { setLocalLogRecords([]); setLocalLogRecordsError(""); return; }
    try {
      const records = await LLogRecordList(requestDirectory);
      if (requestId !== refreshRequestRef.current || currentDirectoryRef.current !== requestDirectory) return;
      setLocalLogRecords((previousRecords) => {
        const previousByRunId = new Map(previousRecords.map((record) => [record.runId, record]));
        return records.map((record) => {
          const previous = previousByRunId.get(record.runId);
          const normalizedRecord = {
            ...record,
            entries: (record.entries ?? []).map((entry) => ({ ...entry, level: LLogLevelNormalize(entry.level) })),
          };
          if (previous && ((previous.entries?.length ?? 0) > 0 || (previous.rawText ?? "").trim())) {
            return { ...normalizedRecord, entries: previous.entries, rawText: previous.rawText };
          }
          return normalizedRecord;
        });
      });
      setLocalLogRecordsError("");
    }
    catch (err) {
      if (requestId !== refreshRequestRef.current || currentDirectoryRef.current !== requestDirectory) return;
      setLocalLogRecords([]); setLocalLogRecordsError(err instanceof Error ? err.message : String(err));
    }
  }

  async function loadLocalLogRecord(runId: string) {
    const requestDirectory = dir;
    if (!requestDirectory || !runId || runId.startsWith("live-")) return;
    const existing = localLogRecords.find((record) => record.runId === runId);
    if (existing && ((existing.entries?.length ?? 0) > 0 || (existing.rawText ?? "").trim())) return;
    const requestKey = `${requestDirectory}\u0000${runId}`;
    const requestId = (detailRequestRef.current.get(requestKey) ?? 0) + 1;
    detailRequestRef.current.set(requestKey, requestId);
    try {
      const record = await LRecordLogGet(requestDirectory, runId);
      if (currentDirectoryRef.current !== requestDirectory || detailRequestRef.current.get(requestKey) !== requestId) return;
      const normalizedRecord = {
        ...record,
        entries: (record.entries ?? []).map((entry) => ({ ...entry, level: LLogLevelNormalize(entry.level) })),
      };
      setLocalLogRecords((records) => records.map((item) => item.runId === runId ? normalizedRecord : item));
      setLocalLogRecordsError("");
    }
    catch (err) {
      if (currentDirectoryRef.current !== requestDirectory || detailRequestRef.current.get(requestKey) !== requestId) return;
      setLocalLogRecordsError(err instanceof Error ? err.message : String(err));
    }
  }

  async function openLocalLogsFolder() {
    if (!dir) return;
    try { await LFolderLogOpen(dir); setLocalLogRecordsError(""); }
    catch (err) { setLocalLogRecordsError(err instanceof Error ? err.message : String(err)); }
  }

  async function openLocalLogRecordFolder(runId: string) {
    if (!dir || !runId || runId.startsWith("live-")) return;
    try { await LFolderRecordOpen(dir, runId); setLocalLogRecordsError(""); }
    catch (err) { setLocalLogRecordsError(err instanceof Error ? err.message : String(err)); }
  }

  async function openLocalLogRecordFile(runId: string, fileName: string) {
    if (!dir || !runId || runId.startsWith("live-")) return;
    try { await LFileRecordOpen(dir, runId, fileName); setLocalLogRecordsError(""); }
    catch (err) { setLocalLogRecordsError(err instanceof Error ? err.message : String(err)); }
  }

  return {
    localLogRecords, localLogRecordsError, setLocalLogRecordsError,
    refreshLocalLogRecords, loadLocalLogRecord,
    openLocalLogsFolder, openLocalLogRecordFolder, openLocalLogRecordFile,
  };
}
