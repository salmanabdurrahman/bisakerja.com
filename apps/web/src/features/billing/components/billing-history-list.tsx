import type { BillingTransaction } from "@/services/billing";

interface BillingHistoryListProps {
  transactions: BillingTransaction[];
  warningMessage?: string | null;
}

export function BillingHistoryList({
  transactions,
  warningMessage = null,
}: BillingHistoryListProps) {
  if (transactions.length === 0) {
    return (
      <section className="bk-card p-5">
        <h3 className="bk-heading-card">Billing history</h3>
        {warningMessage ? (
          <p className="mt-2 rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 bk-body text-amber-800">
            {warningMessage}
          </p>
        ) : (
          <p className="mt-2 bk-body">No transactions yet.</p>
        )}
      </section>
    );
  }

  return (
    <section className="bk-card grid gap-3 p-5">
      <h3 className="bk-heading-card">Billing history</h3>
      {warningMessage ? (
        <p className="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 bk-body text-amber-800">
          {warningMessage}
        </p>
      ) : null}
      <ul className="grid gap-2">
        {transactions.map((transaction) => (
          <li
            key={transaction.id}
            className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2"
          >
            <p className="bk-body font-medium text-black">
              {transaction.status} - {formatIDRCurrency(transaction.amount)}
            </p>
            <p className="bk-body-sm text-[#777777]">
              {new Date(transaction.created_at).toLocaleString("en-US")}
            </p>
          </li>
        ))}
      </ul>
    </section>
  );
}

function formatIDRCurrency(value: number): string {
  return `IDR ${value.toLocaleString("en-US")}`;
}
