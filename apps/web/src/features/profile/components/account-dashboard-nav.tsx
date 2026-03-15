"use client";

import { useRouter } from "next/navigation";

import { Button, ButtonLink } from "@/components/ui/button";
import { useAuthSession } from "@/features/auth/session-provider";
import { clearBrowserSession } from "@/lib/auth/session-cookie";

const accountNavigationLinks = [
  { href: "/account", label: "Account overview" },
  { href: "/account/preferences", label: "Manage preferences" },
  { href: "/account/saved-searches", label: "Saved searches" },
  { href: "/account/tracker", label: "Application tracker" },
  { href: "/account/notifications", label: "Notification center" },
  { href: "/account/subscription", label: "Subscription" },
  { href: "/account/ai-tools", label: "AI tools" },
];

export function AccountDashboardNav() {
  const router = useRouter();
  const { markAnonymous } = useAuthSession();

  function handleLogout() {
    clearBrowserSession();
    markAnonymous();
    router.replace("/auth/login");
  }

  return (
    <nav
      aria-label="Account dashboard navigation"
      className="flex flex-wrap gap-3"
    >
      {accountNavigationLinks.map((link) => (
        <ButtonLink key={link.href} href={link.href} variant="outline">
          {link.label}
        </ButtonLink>
      ))}
      <Button type="button" onClick={handleLogout} variant="primary">
        Logout
      </Button>
    </nav>
  );
}
