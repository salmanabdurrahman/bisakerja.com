import { AppShell } from "@/components/layout/app-shell";
import { LoginForm } from "@/features/auth/components/login-form";
import { normalizeRedirectPath } from "@/lib/auth/redirect-path";

interface LoginPageProps {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}

export default async function LoginPage({ searchParams }: LoginPageProps) {
  const params = await searchParams;
  const redirectPath = normalizeRedirectPath(getSingleParam(params.redirect));
  const initialEmail = getSingleParam(params.email) ?? "";
  const registered = getSingleParam(params.registered) === "1";

  return (
    <AppShell>
      <main className="grid gap-4" role="main">
        <h2 className="text-xl font-semibold">Login</h2>
        <p className="text-sm text-gray-600">
          Masuk untuk mengelola akun dan preferences notifikasi lowongan.
        </p>
        <LoginForm
          redirectPath={redirectPath}
          initialEmail={initialEmail}
          registered={registered}
        />
      </main>
    </AppShell>
  );
}

function getSingleParam(value: string | string[] | undefined): string | null {
  if (Array.isArray(value)) {
    return value[0] ?? null;
  }
  return value ?? null;
}
