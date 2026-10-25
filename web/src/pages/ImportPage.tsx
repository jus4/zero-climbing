import { useMutation } from "@tanstack/react-query";
import { useState, type ChangeEvent } from "react";
import { useNavigate } from "react-router-dom";

import { Alert } from "../components/ui/Alert";
import { Button } from "../components/ui/Button";
import { Card } from "../components/ui/Card";
import { ApiError, apiFetch } from "../lib/api";
import type {
  ImportCandidate,
  ImportConfirmResponse,
  ImportSelection,
  ImportUploadResponse,
} from "../types/api";

const MAX_UPLOAD_BYTES = 50 * 1024 * 1024;

export function ImportPage() {
  const navigate = useNavigate();
  const [token, setToken] = useState<string | null>(null);
  const [candidates, setCandidates] = useState<ImportCandidate[]>([]);
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [visibility, setVisibility] = useState<Map<number, boolean>>(new Map());
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const uploadMutation = useMutation({
    mutationFn: (file: File) => {
      const form = new FormData();
      form.append("file", file);
      return apiFetch<ImportUploadResponse>("/import", { method: "POST", body: form });
    },
  });

  const confirmMutation = useMutation({
    mutationFn: (payload: { token: string; selection: ImportSelection[] }) =>
      apiFetch<ImportConfirmResponse>("/import/confirm", { method: "POST", body: payload }),
  });

  const allSelected = candidates.length > 0 && selected.size === candidates.length;

  function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) {
      return;
    }

    setError(null);
    setNotice(null);

    if (file.size > MAX_UPLOAD_BYTES) {
      setError("File is too large (max 50 MB).");
      return;
    }

    uploadMutation.mutate(file, {
      onSuccess: (data) => {
        setToken(data.token);
        setCandidates(data.candidates);
        setSelected(new Set(data.candidates.map((candidate) => candidate.index)));
        setVisibility(
          new Map(data.candidates.map((candidate) => [candidate.index, candidate.is_public])),
        );
      },
      onError: (err) =>
        setError(err instanceof ApiError ? err.message : "Could not parse the file."),
    });
  }

  function toggle(index: number) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(index)) {
        next.delete(index);
      } else {
        next.add(index);
      }
      return next;
    });
  }

  function toggleAll() {
    setSelected(allSelected ? new Set() : new Set(candidates.map((candidate) => candidate.index)));
  }

  function setPublic(index: number, isPublic: boolean) {
    setVisibility((prev) => {
      const next = new Map(prev);
      next.set(index, isPublic);
      return next;
    });
  }

  function reset() {
    setToken(null);
    setCandidates([]);
    setSelected(new Set());
    setVisibility(new Map());
    setError(null);
    setNotice(null);
    confirmMutation.reset();
  }

  function handleConfirm() {
    if (!token || selected.size === 0) {
      return;
    }

    setError(null);
    confirmMutation.mutate(
      {
        token,
        selection: [...selected].map((index) => ({
          index,
          is_public: visibility.get(index) ?? false,
        })),
      },
      {
        onSuccess: (data) => {
          setNotice(`Imported ${data.imported} location(s).`);
          window.setTimeout(() => navigate("/"), 1200);
        },
        onError: (err) => setError(err instanceof ApiError ? err.message : "Import failed."),
      },
    );
  }

  return (
    <div className="mx-auto max-w-4xl px-4 py-10">
      {token === null ? (
        <Card>
          <h1 className="text-2xl font-semibold text-slate-900">Import map data</h1>
          <p className="mt-2 text-sm text-slate-500">
            Upload a GPX file or a SQLite database. The file will be parsed and you can choose which
            items to import.
          </p>

          {error && (
            <Alert tone="error" className="mt-4">
              {error}
            </Alert>
          )}

          <label className="mt-6 flex flex-col gap-2 text-sm font-medium text-slate-700">
            File
            <input
              type="file"
              accept=".gpx,.db,.sqlite,.sqlite3"
              onChange={handleFileChange}
              disabled={uploadMutation.isPending}
              className="block w-full cursor-pointer rounded-md border border-slate-300 text-sm text-slate-700 file:mr-4 file:cursor-pointer file:border-0 file:bg-slate-900 file:px-4 file:py-2 file:text-sm file:font-semibold file:text-white hover:file:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60"
            />
          </label>
          <p className="mt-2 text-xs text-slate-400">
            Supported: GPX tracks (.gpx) and SQLite databases (.db / .sqlite) containing a{" "}
            <code>place</code> table. Max 50 MB.
          </p>

          {uploadMutation.isPending && (
            <p className="mt-4 text-sm text-slate-500">Parsing file...</p>
          )}
        </Card>
      ) : (
        <Card>
          <h1 className="text-2xl font-semibold text-slate-900">Review your import</h1>
          <p className="mt-2 text-sm text-slate-500">
            Select the items you want to save ({candidates.length} found).
          </p>

          {notice && (
            <Alert tone="success" className="mt-4">
              {notice}
            </Alert>
          )}
          {error && (
            <Alert tone="error" className="mt-4">
              {error}
            </Alert>
          )}

          <div className="mt-6 overflow-x-auto">
            <table className="w-full border-collapse text-left text-sm">
              <thead>
                <tr className="border-b border-slate-200 text-xs uppercase tracking-wide text-slate-500">
                  <th className="px-2 py-2">
                    <input
                      type="checkbox"
                      checked={allSelected}
                      onChange={toggleAll}
                      aria-label="Select all"
                      className="h-4 w-4 rounded border-slate-300"
                    />
                  </th>
                  <th className="px-2 py-2">Type</th>
                  <th className="px-2 py-2">Title</th>
                  <th className="px-2 py-2">Description</th>
                  <th className="px-2 py-2">Details</th>
                  <th className="px-2 py-2">Public</th>
                </tr>
              </thead>
              <tbody>
                {candidates.map((candidate) => (
                  <tr key={candidate.index} className="border-b border-slate-100">
                    <td className="px-2 py-2">
                      <input
                        type="checkbox"
                        checked={selected.has(candidate.index)}
                        onChange={() => toggle(candidate.index)}
                        aria-label={`Select ${candidate.title}`}
                        className="h-4 w-4 rounded border-slate-300"
                      />
                    </td>
                    <td className="px-2 py-2">
                      <span className="rounded-full border border-slate-300 px-2 py-0.5 text-xs font-medium text-slate-600">
                        {candidate.type === "line" ? "Line" : "Point"}
                      </span>
                    </td>
                    <td className="px-2 py-2 font-medium text-slate-800">{candidate.title}</td>
                    <td className="px-2 py-2 text-slate-500">{candidate.description}</td>
                    <td className="px-2 py-2 text-slate-500">{candidate.details}</td>
                    <td className="px-2 py-2">
                      <input
                        type="checkbox"
                        checked={visibility.get(candidate.index) ?? false}
                        onChange={(event) => setPublic(candidate.index, event.target.checked)}
                        aria-label={`Make ${candidate.title} public`}
                        className="h-4 w-4 rounded border-slate-300"
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="mt-6 flex justify-end gap-2">
            <Button variant="ghost" onClick={reset} disabled={confirmMutation.isPending}>
              Back
            </Button>
            <Button
              onClick={handleConfirm}
              disabled={selected.size === 0 || confirmMutation.isPending}
            >
              {confirmMutation.isPending ? "Importing..." : `Import selected (${selected.size})`}
            </Button>
          </div>
        </Card>
      )}
    </div>
  );
}
