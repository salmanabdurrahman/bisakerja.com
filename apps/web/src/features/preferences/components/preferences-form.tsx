"use client";

import { FormEvent, useMemo, useState } from "react";

import { APIRequestError } from "@/lib/utils/fetch-json";
import type {
  PreferredJobType,
  UpdatePreferencesInput,
} from "@/services/preferences";

const maxSalary = 999_000_000;

export const preferenceJobTypeOptions: PreferredJobType[] = [
  "fulltime",
  "parttime",
  "contract",
  "internship",
];

interface PreferencesFormProps {
  initial: UpdatePreferencesInput;
  initialUpdatedAt?: string | null;
  onSubmit: (
    payload: UpdatePreferencesInput,
  ) => Promise<{ updated_at?: string | null }>;
  onUnauthorized: (draft: UpdatePreferencesInput) => void;
}

interface FieldErrors {
  keywords?: string;
  locations?: string;
  job_types?: string;
  salary_min?: string;
}

export function PreferencesForm({
  initial,
  initialUpdatedAt = null,
  onSubmit,
  onUnauthorized,
}: PreferencesFormProps) {
  const [keywordsInput, setKeywordsInput] = useState(
    initial.keywords.join(", "),
  );
  const [locationsInput, setLocationsInput] = useState(
    initial.locations.join(", "),
  );
  const [jobTypesInput, setJobTypesInput] = useState<PreferredJobType[]>(
    initial.job_types,
  );
  const [salaryMinInput, setSalaryMinInput] = useState(
    String(initial.salary_min),
  );
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [formMessage, setFormMessage] = useState<string | null>(null);
  const [lastUpdatedAt, setLastUpdatedAt] = useState<string | null>(
    initialUpdatedAt,
  );
  const [baseline, setBaseline] = useState<UpdatePreferencesInput>(initial);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const normalizedInput = useMemo(
    () =>
      buildNormalizedPreferencesInput({
        keywordsInput,
        locationsInput,
        jobTypesInput,
        salaryMinInput,
      }),
    [keywordsInput, locationsInput, jobTypesInput, salaryMinInput],
  );

  const validationErrors = useMemo(
    () => validateNormalizedInput(normalizedInput),
    [normalizedInput],
  );

  const isValid = Object.keys(validationErrors).length === 0;
  const isDirty = !isSamePreferences(normalizedInput, baseline);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFieldErrors(validationErrors);
    setFormMessage(null);

    if (!isValid) {
      setFormMessage("Input preferences belum valid.");
      return;
    }

    if (!isDirty) {
      setFormMessage("Belum ada perubahan untuk disimpan.");
      return;
    }

    setIsSubmitting(true);
    try {
      const response = await onSubmit(normalizedInput);
      setBaseline(normalizedInput);
      setLastUpdatedAt(response.updated_at ?? null);
      setFieldErrors({});
      setFormMessage("Preferences berhasil disimpan.");
    } catch (error) {
      if (error instanceof APIRequestError) {
        if (error.status === 401) {
          onUnauthorized(normalizedInput);
          setFormMessage(
            "Sesi berakhir. Silakan login ulang untuk melanjutkan.",
          );
          return;
        }
        if (error.status === 429) {
          setFormMessage(
            "Terlalu banyak percobaan simpan. Tunggu sebentar lalu coba lagi.",
          );
          return;
        }
        if (error.status === 400) {
          setFieldErrors(toFieldErrors(error));
          setFormMessage("Data preferences belum valid.");
          return;
        }
        setFormMessage(error.message);
        return;
      }

      setFormMessage("Gagal menyimpan preferences. Coba lagi beberapa saat.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="grid gap-4 rounded-lg border border-gray-200 p-4"
      aria-label="Preferences form"
    >
      <label className="grid gap-1 text-sm">
        <span className="font-medium text-gray-700">Keywords</span>
        <textarea
          value={keywordsInput}
          onChange={(event) => setKeywordsInput(event.target.value)}
          rows={3}
          className="rounded-md border border-gray-300 px-3 py-2"
          placeholder="golang, backend, software engineer"
        />
        {fieldErrors.keywords ? (
          <span className="text-sm text-red-600">{fieldErrors.keywords}</span>
        ) : null}
      </label>

      <label className="grid gap-1 text-sm">
        <span className="font-medium text-gray-700">Locations</span>
        <textarea
          value={locationsInput}
          onChange={(event) => setLocationsInput(event.target.value)}
          rows={2}
          className="rounded-md border border-gray-300 px-3 py-2"
          placeholder="jakarta, remote"
        />
        {fieldErrors.locations ? (
          <span className="text-sm text-red-600">{fieldErrors.locations}</span>
        ) : null}
      </label>

      <fieldset className="grid gap-2">
        <legend className="text-sm font-medium text-gray-700">Job types</legend>
        <div className="flex flex-wrap gap-3">
          {preferenceJobTypeOptions.map((jobType) => (
            <label
              key={jobType}
              className="inline-flex items-center gap-2 text-sm"
            >
              <input
                type="checkbox"
                checked={jobTypesInput.includes(jobType)}
                onChange={(event) =>
                  setJobTypesInput((previous) =>
                    event.target.checked
                      ? dedupeJobTypes([...previous, jobType])
                      : previous.filter((item) => item !== jobType),
                  )
                }
              />
              <span>{jobType}</span>
            </label>
          ))}
        </div>
        {fieldErrors.job_types ? (
          <span className="text-sm text-red-600">{fieldErrors.job_types}</span>
        ) : null}
      </fieldset>

      <label className="grid gap-1 text-sm">
        <span className="font-medium text-gray-700">Salary minimum</span>
        <input
          type="number"
          min={0}
          max={maxSalary}
          value={salaryMinInput}
          onChange={(event) => setSalaryMinInput(event.target.value)}
          className="rounded-md border border-gray-300 px-3 py-2"
        />
        {fieldErrors.salary_min ? (
          <span className="text-sm text-red-600">{fieldErrors.salary_min}</span>
        ) : null}
      </label>

      <div className="flex flex-wrap items-center gap-3">
        <button
          type="submit"
          disabled={isSubmitting || !isValid || !isDirty}
          className="rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {isSubmitting ? "Saving..." : "Save preferences"}
        </button>
        {lastUpdatedAt ? (
          <span className="text-sm text-gray-500">
            Updated at: {new Date(lastUpdatedAt).toLocaleString("id-ID")}
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

interface BuildInput {
  keywordsInput: string;
  locationsInput: string;
  jobTypesInput: PreferredJobType[];
  salaryMinInput: string;
}

function buildNormalizedPreferencesInput(
  input: BuildInput,
): UpdatePreferencesInput {
  return {
    keywords: normalizeList(input.keywordsInput),
    locations: normalizeList(input.locationsInput),
    job_types: dedupeJobTypes(input.jobTypesInput),
    salary_min: normalizeSalary(input.salaryMinInput),
  };
}

function normalizeList(rawValue: string): string[] {
  return Array.from(
    new Set(
      rawValue
        .split(/[,\n]/)
        .map((value) => value.trim().toLowerCase())
        .filter((value) => value.length > 0),
    ),
  );
}

function dedupeJobTypes(values: PreferredJobType[]): PreferredJobType[] {
  return Array.from(new Set(values));
}

function normalizeSalary(rawValue: string): number {
  const value = rawValue.trim();
  if (value.length === 0) {
    return 0;
  }
  const parsed = Number.parseInt(value, 10);
  if (!Number.isFinite(parsed) || Number.isNaN(parsed)) {
    return -1;
  }
  return parsed;
}

function validateNormalizedInput(input: UpdatePreferencesInput): FieldErrors {
  const errors: FieldErrors = {};

  if (input.keywords.length < 1 || input.keywords.length > 10) {
    errors.keywords = "Keywords wajib 1-10 item.";
  } else if (
    input.keywords.some((keyword) => keyword.length < 2 || keyword.length > 50)
  ) {
    errors.keywords = "Panjang keyword harus 2-50 karakter.";
  }

  if (input.locations.length > 5) {
    errors.locations = "Locations maksimal 5 item.";
  } else if (
    input.locations.some(
      (location) => location.length < 2 || location.length > 100,
    )
  ) {
    errors.locations = "Panjang location harus 2-100 karakter.";
  }

  if (input.job_types.length > 4) {
    errors.job_types = "Job types maksimal 4 item.";
  } else if (
    input.job_types.some(
      (jobType) => !preferenceJobTypeOptions.includes(jobType),
    )
  ) {
    errors.job_types = "Job types tidak valid.";
  }

  if (
    !Number.isInteger(input.salary_min) ||
    input.salary_min < 0 ||
    input.salary_min > maxSalary
  ) {
    errors.salary_min = "Salary minimum harus 0 sampai 999000000.";
  }

  return errors;
}

function toFieldErrors(error: APIRequestError): FieldErrors {
  const output: FieldErrors = {};
  for (const item of error.errors) {
    if (item.field === "keywords") {
      output.keywords = item.message;
    }
    if (item.field === "locations") {
      output.locations = item.message;
    }
    if (item.field === "job_types") {
      output.job_types = item.message;
    }
    if (item.field === "salary_min") {
      output.salary_min = item.message;
    }
  }
  return output;
}

function isSamePreferences(
  left: UpdatePreferencesInput,
  right: UpdatePreferencesInput,
): boolean {
  return (
    equalStringArray(left.keywords, right.keywords) &&
    equalStringArray(left.locations, right.locations) &&
    equalStringArray(left.job_types, right.job_types) &&
    left.salary_min === right.salary_min
  );
}

function equalStringArray(left: string[], right: string[]): boolean {
  if (left.length !== right.length) {
    return false;
  }
  return left.every((value, index) => value === right[index]);
}
