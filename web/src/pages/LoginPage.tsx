import { useSearchParams } from "react-router-dom";

import { Alert } from "../components/ui/Alert";
import { apiUrl } from "../lib/api";
import { Paragraph } from "../components/ui/Paragraph";
import { Heading } from "../components/ui/Heading";
import { Divider } from "../components/ui/Divider";

export function LoginPage() {
  const [searchParams] = useSearchParams();
  const error = searchParams.get("error");

  return (
    <div className="mx-auto px-4 py-16 container text-center">
      <Heading as="h1" size="xl" align="center">Kirjaudu</Heading>
      <Divider />
      <Paragraph>Tällä hetkellä ainoastaan Google gmail kirjautuminen on tuettuna.</Paragraph>

      {error && (
        <Alert tone="error" className="mt-4 text-left">
          {error}
        </Alert>
      )}

      <a
        href={apiUrl("/auth/google")}
        className="mt-6 inline-flex items-center justify-center rounded-md bg-primary-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-primary-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2"
      >
        Kirjaudu
      </a>
    </div>
  );
}
