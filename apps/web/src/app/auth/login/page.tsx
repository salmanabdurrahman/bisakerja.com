import { AppShell } from "@/components/layout/app-shell";
import { PageHeader } from "@/components/ui/page-header";
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
      <main className="mx-auto grid w-full max-w-2xl gap-6" role="main">
        <PageHeader
          eyebrow="Authentication"
          title="Login"
          description="Sign in to continue managing your account, growth preferences, and premium billing."
        />
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
