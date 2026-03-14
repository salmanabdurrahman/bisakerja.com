import type { SubscriptionState } from "@/services/auth";

type BadgeState = SubscriptionState | "status_unavailable";

interface SubscriptionBadgeProps {
  state: BadgeState;
  source: "billing" | "profile_fallback" | "unavailable";
}

const labels: Record<BadgeState, string> = {
  free: "Free",
  pending_payment: "Pending payment",
  premium_active: "Premium active",
  premium_expired: "Premium expired",
  status_unavailable: "Status unavailable",
};

const toneByState: Record<BadgeState, string> = {
  free: "border-slate-300 bg-slate-100 text-slate-700",
  pending_payment: "border-amber-300 bg-amber-100 text-amber-900",
  premium_active: "border-emerald-300 bg-emerald-100 text-emerald-900",
  premium_expired: "border-orange-300 bg-orange-100 text-orange-900",
  status_unavailable: "border-red-300 bg-red-100 text-red-800",
};

const sourceLabel: Record<SubscriptionBadgeProps["source"], string> = {
  billing: "Live billing sync",
  profile_fallback: "Profile fallback",
  unavailable: "Sync unavailable",
};

export function SubscriptionBadge({ state, source }: SubscriptionBadgeProps) {
  return (
    <div className="grid gap-1">
      <span
        className={`inline-flex w-fit rounded-full border px-3 py-1 text-xs font-semibold uppercase tracking-wide ${toneByState[state]}`}
      >
        {labels[state]}
      </span>
      <span className="text-xs text-slate-500">{sourceLabel[source]}</span>
    </div>
  );
}
