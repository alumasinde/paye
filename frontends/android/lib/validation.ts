import type { CustomDeduction } from "./types";

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;
const MIN_SUPPORTED_DATE = "2022-01-01";

export type CalculationFormErrors = {
  gross?: string;
  date?: string;
  deductions?: string;
};

export function normaliseAmount(value: string): string {
  return value.replace(/,/g, "").trim();
}

export function isPositiveAmount(value: string): boolean {
  const parsed = Number(normaliseAmount(value));
  return Number.isFinite(parsed) && parsed > 0;
}

export function normaliseDeductions(items: CustomDeduction[]): CustomDeduction[] {
  return items
    .map((item) => ({
      ...item,
      name: item.name.trim(),
      amount: normaliseAmount(item.amount),
    }))
    .filter((item) => item.name.length > 0 || item.amount.length > 0);
}

export function validateCalculationForm(
  gross: string,
  date: string,
  deductions: CustomDeduction[],
): CalculationFormErrors {
  const errors: CalculationFormErrors = {};

  if (!isPositiveAmount(gross)) {
    errors.gross = "Enter a gross monthly salary greater than zero.";
  }

  if (!DATE_PATTERN.test(date) || Number.isNaN(Date.parse(`${date}T00:00:00Z`))) {
    errors.date = "Use a valid date in YYYY-MM-DD format.";
  } else if (date < MIN_SUPPORTED_DATE) {
    errors.date = "Historical calculations are supported from 2022-01-01.";
  }

  const cleaned = normaliseDeductions(deductions);
  if (cleaned.length !== deductions.length) {
    errors.deductions = "Complete or remove empty custom deductions.";
  } else if (cleaned.some((item) => !item.name || !isPositiveAmount(item.amount))) {
    errors.deductions = "Each custom deduction needs a name and an amount greater than zero.";
  }

  return errors;
}
