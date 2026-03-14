import type { Metadata } from "next";
import "./globals.css";
import { AuthSessionProvider } from "@/features/auth/session-provider";

export const metadata: Metadata = {
  title: "Bisakerja",
  description: "Bisakerja frontend app",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id">
      <body>
        <AuthSessionProvider>{children}</AuthSessionProvider>
      </body>
    </html>
  );
}
