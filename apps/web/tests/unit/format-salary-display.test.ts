import { describe, expect, it } from "vitest";

import { formatSalaryDisplay } from "@/lib/utils/format-salary-display";

describe("formatSalaryDisplay", () => {
  it("formats comparator salary fallback values into friendly labels", () => {
    expect(formatSalaryDisplay("<= 2999998")).toBe("Up to Rp 2.999.998");
    expect(formatSalaryDisplay("<= 15000000")).toBe("Up to Rp 15.000.000");
    expect(formatSalaryDisplay("<= Rp 9.000.000")).toBe("Up to Rp 9.000.000");
    expect(formatSalaryDisplay("<= Rp 3.500.000")).toBe("Up to Rp 3.500.000");
    expect(formatSalaryDisplay(">= Rp 7.500.000")).toBe("From Rp 7.500.000");
  });

  it("normalizes shorthand monthly ranges from source labels", () => {
    expect(formatSalaryDisplay("Rp 8 – Rp 12 per month")).toBe(
      "Rp 8.000.000 - Rp 12.000.000 / month",
    );
  });

  it("formats numeric ranges and exact values", () => {
    expect(formatSalaryDisplay("10000000 - 15000000")).toBe(
      "Rp 10.000.000 - Rp 15.000.000",
    );
    expect(formatSalaryDisplay("15000000")).toBe("Rp 15.000.000");
  });

  it("keeps non numeric labels untouched", () => {
    expect(formatSalaryDisplay("Competitive salary")).toBe(
      "Competitive salary",
    );
  });
});
