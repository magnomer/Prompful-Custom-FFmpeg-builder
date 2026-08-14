import { useState } from "react";
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

  async function refreshLocalLogRecords() {
    if (!dir) { setLocalLogRecords([]); setLocalLogRecordsError(""); return; }
    try {
      const records = await LLogRecordList(dir);
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
    catch (err) { setLocalLogRecords([]); setLocalLogRecordsError(err instanceof Error ? err.message : String(err)); }
  }

  async function loadLocalLogRecord(runId: string) {
    if (!dir || !runId || runId.startsWith("live-")) return;
    const existing = localLogRecords.find((record) => record.runId === runId);
    if (existing && ((existing.entries?.length ?? 0) > 0 || (existing.rawText ?? "").trim())) return;
    try {
      const record = await LRecordLogGet(dir, runId);
      const normalizedRecord = {
        ...record,
        entries: (record.entries ?? []).map((entry) => ({ ...entry, level: LLogLevelNormalize(entry.level) })),
      };
      setLocalLogRecords((records) => records.map((item) => item.runId === runId ? normalizedRecord : item));
      setLocalLogRecordsError("");
    }
    catch (err) { setLocalLogRecordsError(err instanceof Error ? err.message : String(err)); }
  }

  async function openLocalLogsFolder() {
    if (!dir) return;
    await LFolderLogOpen(dir);
  }

  async function openLocalLogRecordFolder(runId: string) {
    if (!dir || !runId || runId.startsWith("live-")) return;
    await LFolderRecordOpen(dir, runId);
  }

  async function openLocalLogRecordFile(runId: string, fileName: string) {
    if (!dir || !runId || runId.startsWith("live-")) return;
    await LFileRecordOpen(dir, runId, fileName);
  }

  return {
    localLogRecords, localLogRecordsError, setLocalLogRecordsError,
    refreshLocalLogRecords, loadLocalLogRecord,
    openLocalLogsFolder, openLocalLogRecordFolder, openLocalLogRecordFile,
  };
}
