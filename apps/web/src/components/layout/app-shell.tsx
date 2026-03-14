"use client";

import type { PropsWithChildren } from "react";
import { useEffect, useState } from "react";
import Link from "next/link";
import { HiBars3, HiXMark } from "react-icons/hi2";

import { ButtonLink } from "@/components/ui/button";
import {
  hasBrowserSession,
  subscribeSessionChanged,
} from "@/lib/auth/session-cookie";

const primaryLinks = [
  { href: "/", label: "Home" },
  { href: "/jobs", label: "Jobs" },
  { href: "/pricing", label: "Pricing" },
];

const accountLinks = [
  { href: "/account", label: "Account" },
  { href: "/account/preferences", label: "Preferences" },
  { href: "/account/saved-searches", label: "Saved searches" },
  { href: "/account/ai-tools", label: "AI tools" },
];

export function AppShell({ children }: PropsWithChildren) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(() =>
    hasBrowserSession(),
  );

  useEffect(() => {
    return subscribeSessionChanged(() => {
      setIsAuthenticated(hasBrowserSession());
    });
  }, []);

  return (
    <div className="min-h-screen flex flex-col">
      <header className="sticky top-0 z-40 bg-[#F9F9F9]/80 backdrop-blur-md">
        <div className="mx-auto flex h-18 max-w-6xl items-center justify-between px-4 sm:px-6 lg:px-8">
          <Link
            href="/"
            className="inline-flex items-center gap-2.5 rounded-full px-2 py-1.5 transition hover:bg-black/5"
          >
            <h1 className="text-[18px] sm:text-xl font-normal tracking-tight text-black font-display">
              Bisakerja
            </h1>
          </Link>
          <nav
            className="hidden items-center gap-2 md:flex"
            aria-label="Primary navigation"
          >
            {primaryLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className="rounded-full px-4 py-2 text-[14px] font-normal text-black transition hover:bg-black/5"
              >
                {link.label}
              </Link>
            ))}
          </nav>

          <div className="flex items-center gap-2">
            <div className="hidden sm:flex items-center gap-2">
              {isAuthenticated ? (
                <ButtonLink href="/account" size="md" variant="primary">
                  Dashboard
                </ButtonLink>
              ) : (
                <>
                  <ButtonLink
                    href="/auth/login"
                    size="md"
                    variant="ghost"
                    className="font-normal text-black"
                  >
                    Log in
                  </ButtonLink>
                  <ButtonLink href="/auth/register" size="md" variant="primary">
                    Get Started
                  </ButtonLink>
                </>
              )}
            </div>

            <button
              className="md:hidden p-2 rounded-full hover:bg-black/5 text-black"
              aria-label="Menu"
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            >
              {isMobileMenuOpen ? (
                <HiXMark className="h-6 w-6" />
              ) : (
                <HiBars3 className="h-6 w-6" />
              )}
            </button>
          </div>
        </div>

        {/* Mobile Navigation Menu */}
        {isMobileMenuOpen && (
          <div className="md:hidden bg-white border-t border-gray-100 px-4 py-4 space-y-4 shadow-lg absolute w-full left-0 z-50">
            <nav className="flex flex-col gap-2">
              {primaryLinks.map((link) => (
                <Link
                  key={link.href}
                  href={link.href}
                  className="rounded-lg px-4 py-3 text-[15px] font-medium text-black hover:bg-black/5 transition"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  {link.label}
                </Link>
              ))}
            </nav>
            <div className="flex flex-col gap-3 pt-4 border-t border-gray-100">
              {isAuthenticated ? (
                <ButtonLink
                  href="/account"
                  variant="primary"
                  className="w-full justify-center"
                >
                  Dashboard
                </ButtonLink>
              ) : (
                <>
                  <ButtonLink
                    href="/auth/login"
                    variant="ghost"
                    className="w-full justify-center text-black font-medium"
                  >
                    Log in
                  </ButtonLink>
                  <ButtonLink
                    href="/auth/register"
                    variant="primary"
                    className="w-full justify-center"
                  >
                    Get Started
                  </ButtonLink>
                </>
              )}
            </div>
          </div>
        )}
      </header>

      <div className="mx-auto w-full max-w-6xl px-4 py-12 sm:px-6 sm:py-16 lg:px-8 flex-1">
        {children}
      </div>
      <footer className="bg-[#1A1A1A] text-[#E0E0E0] rounded-t-4xl sm:rounded-t-[48px] mt-16 sm:mt-24">
        <div className="mx-auto flex flex-col lg:grid max-w-6xl gap-12 px-6 py-16 sm:py-20 sm:px-10 lg:grid-cols-[2fr_1fr_1fr_1fr] lg:px-12">
          <div className="grid gap-4 sm:gap-6 text-center sm:text-left items-center sm:items-start">
            <h2 className="text-2xl sm:text-3xl font-normal text-white font-display tracking-tight">
              Bisakerja
            </h2>
            <p className="max-w-sm text-[14px] leading-relaxed text-[#888888] mx-auto sm:mx-0">
              Modern workspace for professionals who want a cleaner flow,
              sharper alerts, and premium velocity.
            </p>
          </div>
          <FooterLinks title="Product" links={primaryLinks} />
          <FooterLinks
            title="Resources"
            links={[
              { href: "/blog", label: "Blog" },
              { href: "/templates", label: "Templates" },
            ]}
          />
          <FooterLinks title="Account" links={accountLinks} />
        </div>
        <div className="mx-auto max-w-6xl px-6 pb-12 pt-8 sm:px-10 lg:px-12 border-t border-[#333333] flex flex-col md:flex-row items-center justify-between gap-4">
          <p className="text-[14px] text-[#555555] text-center md:text-left">
            © {new Date().getFullYear()} Bisakerja. All rights reserved.
          </p>
          <div className="flex items-center gap-6 mt-4 md:mt-0 text-[14px] text-[#555555]">
            <Link href="/privacy" className="hover:text-white transition">
              Privacy Policy
            </Link>
            <Link href="/terms" className="hover:text-white transition">
              Terms of Service
            </Link>
          </div>
        </div>
      </footer>
    </div>
  );
}

interface FooterLinksProps {
  title: string;
  links: Array<{ href: string; label: string }>;
}

function FooterLinks({ title, links }: FooterLinksProps) {
  return (
    <section className="grid gap-6">
      <h3 className="text-[14px] font-medium text-[#888888]">{title}</h3>
      <ul className="grid gap-3">
        {links.map((link) => (
          <li key={link.href}>
            <Link
              href={link.href}
              className="text-[14px] font-normal text-[#E0E0E0] transition hover:text-white"
            >
              {link.label}
            </Link>
          </li>
        ))}
      </ul>
    </section>
  );
}
