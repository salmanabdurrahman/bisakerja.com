import type { PropsWithChildren } from "react";

export function AppShell({ children }: PropsWithChildren) {
  return (
    <div className="mx-auto min-h-screen max-w-5xl px-6 py-10">
      <header className="mb-10">
        <h1 className="text-3xl font-semibold">Bisakerja</h1>
        <p className="mt-2 text-sm text-gray-600">
          Phase 0 foundation for Next.js user-facing application.
        </p>
      </header>
      {children}
    </div>
  );
}
