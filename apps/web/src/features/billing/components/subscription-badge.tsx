import type { SubscriptionState } from "@/services/auth";

type BadgeState = SubscriptionState | "status_unavailable";

interface SubscriptionBadgeProps {
  state: BadgeState;
  source: "billing" | "profile_fallback" | "unavailable";
}

const labels: Record<BadgeState, string> = {
  free: "free",
  pending_payment: "pending_payment",
  premium_active: "premium_active",
  premium_expired: "premium_expired",
  status_unavailable: "status_unavailable",
};

const toneByState: Record<BadgeState, string> = {
  free: "border-gray-300 bg-gray-50 text-gray-700",
  pending_payment: "border-amber-300 bg-amber-50 text-amber-800",
  premium_active: "border-emerald-300 bg-emerald-50 text-emerald-800",
  premium_expired: "border-orange-300 bg-orange-50 text-orange-800",
  status_unavailable: "border-red-300 bg-red-50 text-red-700",
};

const sourceLabel: Record<SubscriptionBadgeProps["source"], string> = {
  billing: "Source: billing/status",
  profile_fallback: "Source: profile fallback",
  unavailable: "Source: unavailable",
};

export function SubscriptionBadge({ state, source }: SubscriptionBadgeProps) {
  return (
    <div className="grid gap-1">
      <span
        className={`inline-flex w-fit rounded-md border px-2 py-1 text-sm font-medium ${toneByState[state]}`}
      >
        {labels[state]}
      </span>
      <span className="text-xs text-gray-500">{sourceLabel[source]}</span>
    </div>
  );
}
