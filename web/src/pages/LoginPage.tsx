import { useSearchParams } from "react-router-dom";

import { Alert } from "../components/ui/Alert";
import { Card } from "../components/ui/Card";
import { apiUrl } from "../lib/api";

export function LoginPage() {
  const [searchParams] = useSearchParams();
  const error = searchParams.get("error");

  return (
    <div className="mx-auto max-w-md px-4 py-16">
      <Card className="text-center">
        <h1 className="text-2xl font-semibold text-slate-900">Sign in</h1>
        <p className="mt-2 text-sm text-slate-500">Use your Google account to continue.</p>

        {error && (
          <Alert tone="error" className="mt-4 text-left">
            {error}
          </Alert>
        )}

        <a
          href={apiUrl("/auth/google")}
          className="mt-6 inline-flex items-center justify-center rounded-md bg-slate-900 px-4 py-2 text-sm font-semibold text-white transition hover:bg-slate-700"
        >
          Sign in with Google
        </a>
      </Card>
    </div>
  );
}
