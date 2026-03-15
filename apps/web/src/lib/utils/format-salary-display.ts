const comparatorPattern = /^(<=|>=|<|>)\s*(?:rp|idr)?\s*([0-9][0-9.,]*)$/i;
const currencyRangePattern =
  /^(?:rp|idr)\s*([0-9][0-9.,]*)\s*[-–]\s*(?:rp|idr)?\s*([0-9][0-9.,]*)(?:\s*(?:\/|per)\s*(month|bulan))?$/i;
const plainRangePattern = /^([0-9][0-9.,]*)\s*[-–]\s*([0-9][0-9.,]*)$/;
const plainNumberPattern = /^(?:rp|idr)?\s*([0-9][0-9.,]*)$/i;

/**
 * formatSalaryDisplay formats heterogeneous salary labels into a readable UI string.
 */
export function formatSalaryDisplay(rawSalary: string): string {
  const normalized = rawSalary.trim();
  if (!normalized) {
    return "";
  }

  const comparatorMatch = normalized.match(comparatorPattern);
  if (comparatorMatch) {
    const amount = parseFlexibleAmount(comparatorMatch[2]);
    if (amount !== null) {
      return formatComparatorDisplay(comparatorMatch[1], amount);
    }
  }

  const currencyRangeMatch = normalized.match(currencyRangePattern);
  if (currencyRangeMatch) {
    const minimum = parseFlexibleAmount(currencyRangeMatch[1]);
    const maximum = parseFlexibleAmount(currencyRangeMatch[2]);
    if (minimum !== null && maximum !== null) {
      const monthlyRange = Boolean(currencyRangeMatch[3]);
      const adjustedMinimum = normalizeMonthlyShorthand(minimum, monthlyRange);
      const adjustedMaximum = normalizeMonthlyShorthand(maximum, monthlyRange);
      const [orderedMinimum, orderedMaximum] =
        adjustedMinimum <= adjustedMaximum
          ? [adjustedMinimum, adjustedMaximum]
          : [adjustedMaximum, adjustedMinimum];
      const suffix = monthlyRange ? " / month" : "";
      return `Rp ${formatIDR(orderedMinimum)} - Rp ${formatIDR(orderedMaximum)}${suffix}`;
    }
  }

  const plainRangeMatch = normalized.match(plainRangePattern);
  if (plainRangeMatch) {
    const minimum = parseFlexibleAmount(plainRangeMatch[1]);
    const maximum = parseFlexibleAmount(plainRangeMatch[2]);
    if (minimum !== null && maximum !== null) {
      const [orderedMinimum, orderedMaximum] =
        minimum <= maximum ? [minimum, maximum] : [maximum, minimum];
      return `Rp ${formatIDR(orderedMinimum)} - Rp ${formatIDR(orderedMaximum)}`;
    }
  }

  const plainNumberMatch = normalized.match(plainNumberPattern);
  if (plainNumberMatch) {
    const amount = parseFlexibleAmount(plainNumberMatch[1]);
    if (amount !== null) {
      return `Rp ${formatIDR(amount)}`;
    }
  }

  return normalized;
}

function normalizeMonthlyShorthand(
  amount: number,
  monthlyRange: boolean,
): number {
  if (!monthlyRange) {
    return amount;
  }
  if (amount > 0 && amount < 1000) {
    return amount * 1_000_000;
  }
  return amount;
}

function formatIDR(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 0,
  }).format(amount);
}

function formatComparatorDisplay(operator: string, amount: number): string {
  const normalizedOperator = operator.trim();
  if (normalizedOperator === "<=" || normalizedOperator === "<") {
    return `Up to Rp ${formatIDR(amount)}`;
  }
  if (normalizedOperator === ">=" || normalizedOperator === ">") {
    return `From Rp ${formatIDR(amount)}`;
  }
  return `${normalizedOperator} Rp ${formatIDR(amount)}`;
}

function parseFlexibleAmount(raw: string): number | null {
  let value = raw.trim().replace(/\s+/g, "");
  if (!value) {
    return null;
  }

  const dotCount = (value.match(/\./g) ?? []).length;
  const commaCount = (value.match(/,/g) ?? []).length;

  if (dotCount > 0 && commaCount > 0) {
    if (value.lastIndexOf(".") > value.lastIndexOf(",")) {
      value = value.replace(/,/g, "");
    } else {
      value = value.replace(/\./g, "").replace(/,/g, ".");
    }
  } else if (commaCount > 0) {
    if (commaCount === 1 && digitsAfterSeparator(value, ",") <= 2) {
      value = value.replace(/,/g, ".");
    } else {
      value = value.replace(/,/g, "");
    }
  } else if (dotCount > 1 || digitsAfterSeparator(value, ".") > 2) {
    value = value.replace(/\./g, "");
  }

  const parsed = Number.parseFloat(value);
  if (!Number.isFinite(parsed) || parsed < 0) {
    return null;
  }

  return Math.floor(parsed);
}

function digitsAfterSeparator(value: string, separator: string): number {
  const chunks = value.split(separator);
  if (chunks.length !== 2) {
    return 0;
  }
  return chunks[1]?.length ?? 0;
}
