"use client";

import { FormEvent, useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { APIRequestError } from "@/lib/utils/fetch-json";
import type {
  NotificationAlertMode,
  NotificationPreferences,
  UpdateNotificationPreferencesInput,
} from "@/services/preferences";

interface NotificationDigestControlProps {
  initialSettings: {
    alert_mode: NotificationAlertMode;
    digest_hour?: number | null;
    updated_at?: string | null;
  };
  onSubmit: (
    payload: UpdateNotificationPreferencesInput,
  ) => Promise<NotificationPreferences>;
  onUnauthorized: () => void;
}

const digestModes: NotificationAlertMode[] = ["daily_digest", "weekly_digest"];

export function NotificationDigestControl({
  initialSettings,
  onSubmit,
  onUnauthorized,
}: NotificationDigestControlProps) {
  const [alertMode, setAlertMode] = useState<NotificationAlertMode>(
    initialSettings.alert_mode,
  );
  const [digestHourInput, setDigestHourInput] = useState(
    initialSettings.digest_hour?.toString() ?? "",
  );
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formMessage, setFormMessage] = useState<string | null>(null);
  const [baseline, setBaseline] = useState({
    alert_mode: initialSettings.alert_mode,
    digest_hour: initialSettings.digest_hour ?? null,
  });
  const [lastUpdatedAt, setLastUpdatedAt] = useState<string | null>(
    initialSettings.updated_at ?? null,
  );

  const isDigestMode = digestModes.includes(alertMode);
  const normalizedDigestHour = useMemo(() => {
    if (!isDigestMode) {
      return null;
    }
    const trimmed = digestHourInput.trim();
    if (trimmed.length === 0) {
      return null;
    }
    const parsed = Number.parseInt(trimmed, 10);
    if (!Number.isInteger(parsed) || parsed < 0 || parsed > 23) {
      return Number.NaN;
    }
    return parsed;
  }, [digestHourInput, isDigestMode]);

  const digestHourError = Number.isNaN(normalizedDigestHour)
    ? "Digest hour must be an integer from 0 to 23."
    : null;

  const isDirty =
    alertMode !== baseline.alert_mode ||
    (isDigestMode ? normalizedDigestHour : null) !== baseline.digest_hour;

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormMessage(null);

    if (digestHourError) {
      setFormMessage(digestHourError);
      return;
    }

    if (!isDirty) {
      setFormMessage("No changes to save.");
      return;
    }

    const payload: UpdateNotificationPreferencesInput = {
      alert_mode: alertMode,
      ...(isDigestMode && normalizedDigestHour !== null
        ? { digest_hour: normalizedDigestHour }
        : {}),
    };

    setIsSubmitting(true);
    try {
      const response = await onSubmit(payload);
      setBaseline({
        alert_mode: response.alert_mode,
        digest_hour: response.digest_hour ?? null,
      });
      setAlertMode(response.alert_mode);
      setDigestHourInput(
        response.digest_hour !== null && response.digest_hour !== undefined
          ? String(response.digest_hour)
          : "",
      );
      setLastUpdatedAt(response.updated_at ?? null);
      setFormMessage("Notification settings saved successfully.");
    } catch (error) {
      if (error instanceof APIRequestError) {
        if (error.status === 401) {
          onUnauthorized();
          return;
        }
        if (error.status === 429) {
          setFormMessage(
            "Too many save attempts. Please wait a moment and try again.",
          );
          return;
        }
        setFormMessage(error.message);
        return;
      }
      setFormMessage("Failed to save notification settings.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="bk-card grid gap-4 p-5"
      aria-label="Notification digest control form"
    >
      <h3 className="text-lg font-semibold text-slate-900">
        Notification frequency settings
      </h3>

      <label className="grid gap-1 text-sm">
        <span className="font-medium text-slate-700">Alert mode</span>
        <select
          value={alertMode}
          onChange={(event) =>
            setAlertMode(event.target.value as NotificationAlertMode)
          }
          className="bk-select"
        >
          <option value="instant">Instant</option>
          <option value="daily_digest">Daily Digest</option>
          <option value="weekly_digest">Weekly Digest</option>
        </select>
      </label>

      {isDigestMode ? (
        <label className="grid gap-1 text-sm">
          <span className="font-medium text-slate-700">
            Digest hour (optional, 0-23)
          </span>
          <input
            type="number"
            min={0}
            max={23}
            value={digestHourInput}
            onChange={(event) => setDigestHourInput(event.target.value)}
            className="bk-input"
            placeholder="9"
          />
          {digestHourError ? (
            <span className="text-sm text-red-600">{digestHourError}</span>
          ) : (
            <span className="text-xs text-gray-500">
              If left empty, uses the default hour `9`.
            </span>
          )}
        </label>
      ) : null}

      <div className="flex flex-wrap items-center gap-3">
        <Button
          type="submit"
          disabled={isSubmitting || !isDirty || Boolean(digestHourError)}
          variant="secondary"
        >
          {isSubmitting ? "Saving..." : "Save notification settings"}
        </Button>
        {lastUpdatedAt ? (
          <span className="text-sm text-slate-500">
            Updated at: {new Date(lastUpdatedAt).toLocaleString("en-US")}
          </span>
        ) : null}
      </div>

      {formMessage ? (
        <p className="text-sm text-gray-700" role="status" aria-live="polite">
          {formMessage}
        </p>
      ) : null}
    </form>
  );
}
