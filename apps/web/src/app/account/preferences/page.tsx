import { AppShell } from "@/components/layout/app-shell";
import { AccountPreferencesClient } from "@/features/preferences/components/account-preferences-client";
import {
  loadPreferencesDraftFromCookie,
  preferencesDraftCookie,
} from "@/features/preferences/draft";
import { buildLoginHref } from "@/lib/auth/redirect-path";
import { resolveServerAccessToken } from "@/lib/auth/server-session";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { getMe, type SubscriptionState } from "@/services/auth";
import { getBillingStatus } from "@/services/billing";
import type { UpdatePreferencesInput } from "@/services/preferences";
import { getPreferences } from "@/services/preferences";
import type { NotificationAlertMode } from "@/services/preferences";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";

export default async function AccountPreferencesPage() {
  const accessToken = await resolveServerAccessToken();
  if (!accessToken) {
    redirect(buildLoginHref("/account/preferences"));
  }

  const cookieStore = await cookies();
  const cookieDraft = loadPreferencesDraftFromCookie(
    cookieStore.get(preferencesDraftCookie)?.value,
  );

  let profileSubscriptionState: SubscriptionState | null = null;
  let fallbackPremium = false;
  try {
    const profileResponse = await getMe(accessToken);
    profileSubscriptionState = profileResponse.data.subscription_state ?? null;
    fallbackPremium = profileResponse.data.is_premium;
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/account/preferences"));
    }
  }

  let subscriptionState: SubscriptionState | "status_unavailable" =
    profileSubscriptionState ?? (fallbackPremium ? "premium_active" : "free");
  let infoMessage: string | null = null;

  try {
    const billingResponse = await getBillingStatus(accessToken);
    subscriptionState = billingResponse.data.subscription_state;
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/account/preferences"));
    }
    infoMessage =
      "Status premium belum bisa diverifikasi dari billing/status. Menggunakan fallback sementara.";
    if (!profileSubscriptionState && !fallbackPremium) {
      subscriptionState = "status_unavailable";
    }
  }

  let initialPreferences: UpdatePreferencesInput | null = cookieDraft;
  let initialUpdatedAt: string | null = null;
  let initialNotificationSettings: {
    alert_mode: NotificationAlertMode;
    digest_hour?: number | null;
    updated_at?: string | null;
  } = {
    alert_mode: "instant",
    digest_hour: null,
    updated_at: null,
  };

  try {
    const preferencesResponse = await getPreferences(accessToken);
    initialNotificationSettings = {
      alert_mode: preferencesResponse.data.alert_mode,
      digest_hour: preferencesResponse.data.digest_hour ?? null,
      updated_at: preferencesResponse.data.updated_at ?? null,
    };

    if (!initialPreferences) {
      initialPreferences = {
        keywords: preferencesResponse.data.keywords,
        locations: preferencesResponse.data.locations,
        job_types: preferencesResponse.data.job_types,
        salary_min: preferencesResponse.data.salary_min,
      };
      initialUpdatedAt = preferencesResponse.data.updated_at ?? null;
    }
  } catch (error) {
    if (error instanceof APIRequestError && error.status === 401) {
      redirect(buildLoginHref("/account/preferences"));
    }
  }

  if (cookieDraft) {
    infoMessage =
      "Draft lokal berhasil dipulihkan. Simpan ulang preferences untuk sinkronisasi ke server.";
  }

  if (!initialPreferences) {
    return (
      <AppShell>
        <main className="grid gap-4" role="main">
          <div className="grid gap-1">
            <h2 className="text-xl font-semibold">Preferences</h2>
            <p className="text-sm text-gray-600">
              Atur keyword, lokasi, dan tipe kerja untuk personalisasi
              notifikasi.
            </p>
          </div>
          <section className="grid gap-3 rounded-lg border border-red-200 bg-red-50 p-4">
            <h3 className="text-lg font-semibold text-red-900">
              Gagal memuat preferences
            </h3>
            <p className="text-sm text-red-800">
              Data preferences belum bisa diambil dari server. Coba refresh
              halaman.
            </p>
            <a
              href="/account/preferences"
              className="w-fit rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:opacity-90"
            >
              Coba lagi
            </a>
          </section>
        </main>
      </AppShell>
    );
  }

  return (
    <AppShell>
      <main className="grid gap-4" role="main">
        <div className="grid gap-1">
          <h2 className="text-xl font-semibold">Preferences</h2>
          <p className="text-sm text-gray-600">
            Atur keyword, lokasi, dan tipe kerja untuk personalisasi notifikasi.
          </p>
        </div>
        <AccountPreferencesClient
          initialPreferences={initialPreferences}
          initialUpdatedAt={initialUpdatedAt}
          initialNotificationSettings={initialNotificationSettings}
          subscriptionState={subscriptionState}
          infoMessage={infoMessage}
        />
      </main>
    </AppShell>
  );
}
