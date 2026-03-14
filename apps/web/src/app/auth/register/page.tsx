import { AppShell } from "@/components/layout/app-shell";
import { PageHeader } from "@/components/ui/page-header";
import { RegisterForm } from "@/features/auth/components/register-form";

export default function RegisterPage() {
  return (
    <AppShell>
      <main className="mx-auto grid w-full max-w-2xl gap-6" role="main">
        <PageHeader
          eyebrow="Authentication"
          title="Register"
          description="Create your account to save searches, tune notifications, and unlock premium workflows."
        />
        <RegisterForm />
      </main>
    </AppShell>
  );
}
