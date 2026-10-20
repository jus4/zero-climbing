import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";

import { Alert } from "../components/ui/Alert";
import { Button } from "../components/ui/Button";
import { Card } from "../components/ui/Card";
import { Input } from "../components/ui/Input";
import { ApiError, apiFetch } from "../lib/api";
import { useAuth } from "../providers/authContext";

export function OnboardingPage() {
  const { user, refetch } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    setSubmitting(true);

    try {
      await apiFetch("/auth/username", { method: "POST", body: { username } });
      await refetch();
      navigate("/", { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong. Please try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="mx-auto max-w-md px-4 py-16">
      <Card>
        <h1 className="text-2xl font-semibold text-slate-900">Choose a username</h1>
        {user?.email && <p className="mt-1 text-sm text-slate-500">Signed in as {user.email}</p>}

        {error && (
          <Alert tone="error" className="mt-4">
            {error}
          </Alert>
        )}

        <form className="mt-6 flex flex-col gap-4" onSubmit={handleSubmit}>
          <label className="flex flex-col gap-1 text-sm font-medium text-slate-700">
            Username
            <Input
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              minLength={4}
              maxLength={64}
              autoComplete="username"
              required
              autoFocus
            />
          </label>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Saving..." : "Continue"}
          </Button>
        </form>
      </Card>
    </div>
  );
}
