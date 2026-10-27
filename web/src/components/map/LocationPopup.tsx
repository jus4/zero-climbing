import { useState } from "react";

import { ApiError } from "../../lib/api";
import type { LocationFeatureProperties } from "../../types/api";
import { Alert } from "../ui/Alert";
import { ConfirmDialog } from "../ui/ConfirmDialog";
import { Input } from "../ui/Input";

type LocationPopupProps = {
  id: number;
  properties: LocationFeatureProperties;
  currentUserId: number | null;
  onUpdate: (input: {
    id: number;
    title: string;
    description: string;
    isPublic: boolean;
  }) => Promise<unknown>;
  onDelete: (id: number) => Promise<unknown>;
  onClose: () => void;
};

export function LocationPopup({
  id,
  properties,
  currentUserId,
  onUpdate,
  onDelete,
  onClose,
}: LocationPopupProps) {
  const [editing, setEditing] = useState(false);
  const [title, setTitle] = useState(properties.title);
  const [description, setDescription] = useState(properties.description);
  const [isPublic, setIsPublic] = useState(properties.is_public);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const isOwner = currentUserId !== null && currentUserId === properties.user_id;

  function startEdit() {
    setError(null);
    setTitle(properties.title);
    setDescription(properties.description);
    setIsPublic(properties.is_public);
    setEditing(true);
  }

  async function handleSave() {
    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      setError("Title cannot be empty.");
      return;
    }

    setError(null);
    setSaving(true);
    try {
      await onUpdate({ id, title: trimmedTitle, description: description.trim(), isPublic });
      setEditing(false);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save changes.");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    setError(null);
    setDeleting(true);
    try {
      await onDelete(id);
      onClose();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not delete this location.");
      setDeleting(false);
      setConfirmOpen(false);
    }
  }

  return (
    <div className="relative min-w-[220px] max-w-[320px] rounded-lg border border-slate-200 bg-white p-4 shadow-lg">
      <button
        type="button"
        aria-label="Close"
        onClick={onClose}
        className="absolute right-2 top-1 text-lg leading-none text-slate-400 transition hover:text-slate-600"
      >
        ×
      </button>

      {error && (
        <Alert tone="error" className="mb-3 mr-4">
          {error}
        </Alert>
      )}

      {editing ? (
        <div className="flex flex-col gap-2">
          <Input
            value={title}
            onChange={(event) => setTitle(event.target.value)}
            placeholder="Title"
          />
          <textarea
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            rows={3}
            placeholder="Description"
            className="w-full resize-y rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-900 outline-none transition focus:border-slate-500 focus:ring-2 focus:ring-slate-200"
          />
          <label className="flex items-center gap-2 text-xs font-medium text-slate-600">
            <input
              type="checkbox"
              checked={isPublic}
              onChange={(event) => setIsPublic(event.target.checked)}
              className="h-4 w-4 rounded border-slate-300"
            />
            Public location
          </label>
          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => setEditing(false)}
              className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-xs font-semibold text-slate-900 transition hover:bg-slate-50"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={() => void handleSave()}
              disabled={saving}
              className="rounded-md bg-slate-900 px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {saving ? "Saving..." : "Save"}
            </button>
          </div>
        </div>
      ) : (
        <div>
          <div className="pr-4 text-[15px] font-semibold text-slate-900">
            {properties.title || "Untitled"}
          </div>
          <div className="mt-1 whitespace-pre-wrap break-words text-[13px] text-slate-600">
            {properties.description || "No comments."}
          </div>
          {properties.user && (
            <div className="mt-2 border-t border-slate-200 pt-2 text-xs text-slate-400">
              Added by {properties.user}
            </div>
          )}

          {isOwner && (
            <div className="mt-3 flex gap-2 border-t border-slate-200 pt-3">
              <button
                type="button"
                onClick={startEdit}
                className="rounded-md bg-orange-500 px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-orange-400"
              >
                Edit
              </button>
              <button
                type="button"
                onClick={() => setConfirmOpen(true)}
                className="rounded-md bg-red-600 px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-red-500"
              >
                Delete
              </button>
            </div>
          )}
        </div>
      )}

      <ConfirmDialog
        open={confirmOpen}
        title="Delete this location?"
        message="This cannot be undone."
        confirmLabel="Delete"
        busy={deleting}
        onConfirm={() => void handleDelete()}
        onCancel={() => setConfirmOpen(false)}
      />
    </div>
  );
}
