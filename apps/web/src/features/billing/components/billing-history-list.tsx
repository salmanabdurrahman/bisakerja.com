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
      <section className="rounded-lg border border-gray-200 p-4">
        <h3 className="text-lg font-semibold text-gray-900">Billing history</h3>
        {warningMessage ? (
          <p className="mt-2 text-sm text-amber-700">{warningMessage}</p>
        ) : (
          <p className="mt-2 text-sm text-gray-600">Belum ada transaksi.</p>
        )}
      </section>
    );
  }

  return (
    <section className="grid gap-3 rounded-lg border border-gray-200 p-4">
      <h3 className="text-lg font-semibold text-gray-900">Billing history</h3>
      {warningMessage ? (
        <p className="text-sm text-amber-700">{warningMessage}</p>
      ) : null}
      <ul className="grid gap-2">
        {transactions.map((transaction) => (
          <li
            key={transaction.id}
            className="rounded-md border border-gray-200 px-3 py-2 text-sm text-gray-700"
          >
            <p className="font-medium">
              {transaction.status} - Rp
              {transaction.amount.toLocaleString("id-ID")}
            </p>
            <p className="text-xs text-gray-500">
              {new Date(transaction.created_at).toLocaleString("id-ID")}
            </p>
          </li>
        ))}
      </ul>
    </section>
  );
}
