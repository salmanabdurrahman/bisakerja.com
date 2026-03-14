import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { NotificationDigestControl } from "@/features/preferences/components/notification-digest-control";

describe("NotificationDigestControl", () => {
  it("submits digest mode update", async () => {
    const submitMock = vi.fn().mockResolvedValue({
      user_id: "user_1",
      alert_mode: "daily_digest",
      digest_hour: 8,
      updated_at: "2030-01-01T00:00:00Z",
    });

    render(
      <NotificationDigestControl
        initialSettings={{
          alert_mode: "instant",
          digest_hour: null,
          updated_at: null,
        }}
        onSubmit={submitMock}
        onUnauthorized={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText("Alert mode"), {
      target: { value: "daily_digest" },
    });
    fireEvent.change(screen.getByRole("spinbutton"), {
      target: { value: "8" },
    });

    fireEvent.click(
      screen.getByRole("button", { name: "Save notification settings" }),
    );

    await waitFor(() => {
      expect(submitMock).toHaveBeenCalledWith({
        alert_mode: "daily_digest",
        digest_hour: 8,
      });
    });
  });

  it("shows validation message when digest hour is invalid", async () => {
    const submitMock = vi.fn();

    render(
      <NotificationDigestControl
        initialSettings={{
          alert_mode: "daily_digest",
          digest_hour: 9,
          updated_at: null,
        }}
        onSubmit={submitMock}
        onUnauthorized={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByRole("spinbutton"), {
      target: { value: "99" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Save notification settings" }),
    );

    expect(
      screen.getByText("Digest hour harus integer 0-23."),
    ).toBeInTheDocument();
    expect(submitMock).not.toHaveBeenCalled();
  });
});
