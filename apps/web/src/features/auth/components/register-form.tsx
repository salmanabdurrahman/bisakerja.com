"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

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
      className="grid gap-4 rounded-lg border border-gray-200 p-4"
      aria-label="Register form"
    >
      <label className="grid gap-1 text-sm">
        <span className="font-medium text-gray-700">Name</span>
        <input
          type="text"
          name="name"
          autoComplete="name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          className="rounded-md border border-gray-300 px-3 py-2"
          required
          minLength={2}
          maxLength={100}
        />
        {fieldErrors.name ? (
          <span className="text-sm text-red-600">{fieldErrors.name}</span>
        ) : null}
      </label>

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
          autoComplete="new-password"
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
        {isSubmitting ? "Creating account..." : "Create account"}
      </button>

      <p className="text-sm text-gray-600">
        Already have an account?{" "}
        <a href="/auth/login" className="text-blue-700 underline">
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
