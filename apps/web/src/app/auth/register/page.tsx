import { AppShell } from "@/components/layout/app-shell";
import { RegisterForm } from "@/features/auth/components/register-form";

export default function RegisterPage() {
  return (
    <AppShell>
      <main className="grid gap-4" role="main">
        <h2 className="text-xl font-semibold">Register</h2>
        <p className="text-sm text-gray-600">
          Buat akun baru untuk menyimpan preferences dan mengakses fitur
          premium.
        </p>
        <RegisterForm />
      </main>
    </AppShell>
  );
}
