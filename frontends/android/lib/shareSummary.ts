import type { Calculation } from "./types";
import { kes } from "./money";

// Plain-text summary used for both the native share sheet and the
// clipboard copy - one function so the two stay identical.
export function buildShareSummary(data: Calculation): string {
  return [
    "Budget254 PAYE Estimate",
    `Calculated for ${data.calculation_date}`,
    "",
    `Gross salary: ${kes(data.gross_salary)}`,
    `PAYE: ${kes(data.paye)}`,
    `Total deductions: ${kes(data.total_deductions)}`,
    `Estimated take-home: ${kes(data.net_salary)}`,
    "",
    "Calculated with the Budget254 PAYE calculator.",
  ].join("\n");
}
