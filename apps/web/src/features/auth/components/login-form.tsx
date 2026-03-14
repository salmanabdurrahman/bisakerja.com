"use client";

import { FormEvent, useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import {
  clearBrowserSession,
  writeBrowserSession,
} from "@/lib/auth/session-cookie";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { useAuthSession } from "@/features/auth/session-provider";
import { getMe, loginUser } from "@/services/auth";

interface LoginFormProps {
  initialEmail?: string;
  redirectPath: string;
  registered?: boolean;
}

interface LoginFieldErrors {
  email?: string;
  password?: string;
}

export function LoginForm({
  initialEmail = "",
  redirectPath,
  registered = false,
}: LoginFormProps) {
  const router = useRouter();
  const { markAuthenticated } = useAuthSession();

  const [email, setEmail] = useState(initialEmail);
  const [password, setPassword] = useState("");
  const [fieldErrors, setFieldErrors] = useState<LoginFieldErrors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const successMessage = useMemo(() => {
    if (!registered) {
      return null;
    }
    return "Pendaftaran berhasil. Silakan login untuk melanjutkan.";
  }, [registered]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFieldErrors({});
    setFormError(null);
    setIsSubmitting(true);

    try {
      const response = await loginUser({
        email: email.trim(),
        password,
      });

      writeBrowserSession({
        accessToken: response.data.access_token,
        refreshToken: response.data.refresh_token,
        expiresIn: response.data.expires_in,
      });

      try {
        await getMe(response.data.access_token);
      } catch (error) {
        clearBrowserSession();
        throw error;
      }

      markAuthenticated();
      router.replace(redirectPath);
    } catch (error) {
      if (error instanceof APIRequestError) {
        setFieldErrors(extractFieldErrors(error));
        setFormError(toLoginMessage(error));
      } else {
        setFormError("Login gagal. Coba lagi dalam beberapa saat.");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="grid gap-4 rounded-lg border border-gray-200 p-4"
      aria-label="Login form"
    >
      {successMessage ? (
        <p className="rounded-md border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-800">
          {successMessage}
        </p>
      ) : null}

      <label className="grid gap-1 text-sm">
        <span className="font-medium text-gray-700">Email</span>
        <input
          type="email"
          name="email"
          autoComplete="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="rounded-md border border-gray-300 px-3 py-2"
          required
        />
        {fieldErrors.email ? (
          <span className="text-sm text-red-600">{fieldErrors.email}</span>
        ) : null}
      </label>

      <label className="grid gap-1 text-sm">
        <span className="font-medium text-gray-700">Password</span>
        <input
          type="password"
          name="password"
          autoComplete="current-password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          className="rounded-md border border-gray-300 px-3 py-2"
          required
          minLength={8}
        />
        {fieldErrors.password ? (
          <span className="text-sm text-red-600">{fieldErrors.password}</span>
        ) : null}
      </label>

      {formError ? (
        <p className="text-sm text-red-600" role="alert">
          {formError}
        </p>
      ) : null}

      <button
        type="submit"
        className="rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
        disabled={isSubmitting}
      >
        {isSubmitting ? "Signing in..." : "Sign in"}
      </button>

      <p className="text-sm text-gray-600" aria-live="polite">
        Belum punya akun?{" "}
        <a href="/auth/register" className="text-blue-700 underline">
          Daftar di sini
        </a>
      </p>
    </form>
  );
}

function extractFieldErrors(error: APIRequestError): LoginFieldErrors {
  const byField: LoginFieldErrors = {};
  for (const item of error.errors) {
    if (item.field === "email") {
      byField.email = item.message;
    }
    if (item.field === "password") {
      byField.password = item.message;
    }
  }
  return byField;
}

function toLoginMessage(error: APIRequestError): string {
  if (error.status === 401) {
    return "Email atau password salah.";
  }
  if (error.status === 429) {
    return "Terlalu banyak percobaan login. Tunggu sebentar lalu coba lagi.";
  }
  if (error.status === 400 && error.errors.length > 0) {
    return "Input login belum valid. Periksa lagi email dan password.";
  }
  return error.message;
}
