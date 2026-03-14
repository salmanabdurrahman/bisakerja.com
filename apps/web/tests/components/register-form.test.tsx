import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { RegisterForm } from "@/features/auth/components/register-form";
import { registerUser } from "@/services/auth";

const replaceMock = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    replace: replaceMock,
  }),
}));

vi.mock("@/services/auth", async () => {
  const actual =
    await vi.importActual<typeof import("@/services/auth")>("@/services/auth");
  return {
    ...actual,
    registerUser: vi.fn(),
  };
});

describe("RegisterForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("redirects to login page after successful register", async () => {
    vi.mocked(registerUser).mockResolvedValueOnce({
      meta: { code: 201, status: "success", message: "User registered" },
      data: {
        id: "user_1",
        email: "new@example.com",
        name: "New User",
        role: "user",
        created_at: "2026-03-14T00:00:00Z",
      },
    });

    render(<RegisterForm />);

    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "New User" },
    });
    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "new@example.com" },
    });
    fireEvent.change(screen.getByLabelText("Password"), {
      target: { value: "StrongPass1" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create account" }));

    await waitFor(() => {
      expect(registerUser).toHaveBeenCalledWith({
        name: "New User",
        email: "new@example.com",
        password: "StrongPass1",
      });
    });

    expect(replaceMock).toHaveBeenCalledWith(
      "/auth/login?registered=1&email=new%40example.com",
    );
  });
});
