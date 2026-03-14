import type { SubscriptionState } from "@/services/auth";

interface NotificationEntitlementBannerProps {
  subscriptionState: SubscriptionState | "status_unavailable";
}

export function NotificationEntitlementBanner({
  subscriptionState,
}: NotificationEntitlementBannerProps) {
  if (subscriptionState === "premium_active") {
    return (
      <section className="rounded-lg border border-emerald-200 bg-emerald-50 p-4">
        <h3 className="text-base font-semibold text-emerald-900">
          Notifikasi premium aktif
        </h3>
        <p className="mt-1 text-sm text-emerald-800">
          Preferensi yang disimpan akan dipakai untuk matching lowongan
          otomatis.
        </p>
      </section>
    );
  }

  if (subscriptionState === "pending_payment") {
    return (
      <section className="rounded-lg border border-amber-200 bg-amber-50 p-4">
        <h3 className="text-base font-semibold text-amber-900">
          Pembayaran masih diproses
        </h3>
        <p className="mt-1 text-sm text-amber-800">
          Notifikasi akan aktif setelah pembayaran berhasil.
        </p>
        <a
          href="/pricing"
          className="mt-3 inline-flex text-sm text-blue-700 underline"
        >
          Lanjutkan pembayaran
        </a>
      </section>
    );
  }

  if (subscriptionState === "premium_expired") {
    return (
      <section className="rounded-lg border border-orange-200 bg-orange-50 p-4">
        <h3 className="text-base font-semibold text-orange-900">
          Premium sudah berakhir
        </h3>
        <p className="mt-1 text-sm text-orange-800">
          Perbarui langganan untuk mengaktifkan kembali notifikasi matching.
        </p>
        <a
          href="/pricing"
          className="mt-3 inline-flex text-sm text-blue-700 underline"
        >
          Upgrade kembali
        </a>
      </section>
    );
  }

  if (subscriptionState === "status_unavailable") {
    return (
      <section className="rounded-lg border border-red-200 bg-red-50 p-4">
        <h3 className="text-base font-semibold text-red-900">
          Status premium belum tersedia
        </h3>
        <p className="mt-1 text-sm text-red-800">
          Kami belum bisa mengambil status billing terbaru. Coba refresh
          halaman.
        </p>
      </section>
    );
  }

  return (
    <section className="rounded-lg border border-gray-200 bg-gray-50 p-4">
      <h3 className="text-base font-semibold text-gray-900">
        Notifikasi untuk premium
      </h3>
      <p className="mt-1 text-sm text-gray-700">
        User free tetap bisa menyimpan preferensi, tetapi notifikasi matching
        hanya aktif untuk premium.
      </p>
      <a
        href="/pricing"
        className="mt-3 inline-flex text-sm text-blue-700 underline"
      >
        Lihat paket premium
      </a>
    </section>
  );
}
