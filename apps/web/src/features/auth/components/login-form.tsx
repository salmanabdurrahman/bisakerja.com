"use client";

import { FormEvent, useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
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
    return "Registration successful. Please sign in to continue.";
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
        setFormError("Login failed. Please try again shortly.");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="bk-card grid gap-5 p-6 sm:p-8"
      aria-label="Login form"
    >
      {successMessage ? (
        <p className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-[14px] text-emerald-800">
          {successMessage}
        </p>
      ) : null}

      <label className="grid gap-2 text-[14px]">
        <span className="font-medium text-black">Email</span>
        <input
          type="email"
          name="email"
          autoComplete="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="bk-input"
          required
        />
        {fieldErrors.email ? (
          <span className="text-sm text-red-600">{fieldErrors.email}</span>
        ) : null}
      </label>

      <label className="grid gap-2 text-[14px]">
        <span className="font-medium text-black">Password</span>
        <input
          type="password"
          name="password"
          autoComplete="current-password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          className="bk-input"
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

      <Button type="submit" variant="primary" size="lg" disabled={isSubmitting}>
        {isSubmitting ? "Signing in..." : "Sign in"}
      </Button>

      <p className="text-[14px] text-[#666666]" aria-live="polite">
        Don&apos;t have an account?{" "}
        <a href="/auth/register" className="bk-link text-black underline">
          Register here
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
    return "Email or password is incorrect.";
  }
  if (error.status === 429) {
    return "Too many login attempts. Please wait a moment and try again.";
  }
  if (error.status === 400 && error.errors.length > 0) {
    return "Login input is invalid. Please check your email and password.";
  }
  return error.message;
}
