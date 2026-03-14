import type { Metadata } from "next";
import { Montserrat, Roboto } from "next/font/google";

import { WebVitalsObserver } from "@/features/observability/components/web-vitals-observer";
import "./globals.css";
import { AuthSessionProvider } from "@/features/auth/session-provider";

const primaryFont = Roboto({
  subsets: ["latin"],
  variable: "--font-primary",
  weight: ["300", "400", "500", "700"],
});

const displayFont = Montserrat({
  subsets: ["latin"],
  variable: "--font-display",
  weight: ["500", "600", "700", "800"],
});

export const metadata: Metadata = {
  title: "Bisakerja | Smart Job Discovery",
  description:
    "A modern SaaS-style experience for job discovery, account growth, and premium upgrades.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${primaryFont.variable} ${displayFont.variable}`}>
        <AuthSessionProvider>
          <WebVitalsObserver />
          {children}
        </AuthSessionProvider>
      </body>
    </html>
  );
}
