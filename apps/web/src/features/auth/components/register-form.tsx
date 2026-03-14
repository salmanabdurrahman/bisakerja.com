"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { APIRequestError } from "@/lib/utils/fetch-json";
import { registerUser } from "@/services/auth";

interface RegisterFieldErrors {
  email?: string;
  password?: string;
  name?: string;
}

export function RegisterForm() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [fieldErrors, setFieldErrors] = useState<RegisterFieldErrors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFieldErrors({});
    setFormError(null);
    setIsSubmitting(true);

    try {
      const normalizedEmail = email.trim();
      await registerUser({
        name: name.trim(),
        email: normalizedEmail,
        password,
      });
      router.replace(
        `/auth/login?registered=1&email=${encodeURIComponent(normalizedEmail)}`,
      );
    } catch (error) {
      if (error instanceof APIRequestError) {
        setFieldErrors(extractFieldErrors(error));
        setFormError(toRegisterMessage(error));
      } else {
        setFormError("Registration failed. Please try again shortly.");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="bk-card grid gap-5 p-6 sm:p-8"
      aria-label="Register form"
    >
      <label className="grid gap-2 text-[14px]">
        <span className="font-medium text-black">Name</span>
        <input
          type="text"
          name="name"
          autoComplete="name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          className="bk-input"
          required
        />
        {fieldErrors.name ? (
          <span className="text-sm text-red-600">{fieldErrors.name}</span>
        ) : null}
      </label>

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
          autoComplete="new-password"
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
        {isSubmitting ? "Creating account..." : "Create account"}
      </Button>

      <p className="text-[14px] text-[#666666]">
        Already have an account?{" "}
        <a href="/auth/login" className="bk-link text-black underline">
          Sign in here
        </a>
      </p>
    </form>
  );
}

function extractFieldErrors(error: APIRequestError): RegisterFieldErrors {
  const byField: RegisterFieldErrors = {};
  for (const item of error.errors) {
    if (item.field === "email") {
      byField.email = item.message;
    }
    if (item.field === "password") {
      byField.password = item.message;
    }
    if (item.field === "name") {
      byField.name = item.message;
    }
  }
  return byField;
}

function toRegisterMessage(error: APIRequestError): string {
  if (error.status === 409) {
    return "Email is already registered. Please sign in or use another email.";
  }
  if (error.status === 429) {
    return "Too many requests. Please wait a moment and try again.";
  }
  if (error.status === 400 && error.errors.length > 0) {
    return "Registration data is invalid. Please review and try again.";
  }
  return error.message;
}
