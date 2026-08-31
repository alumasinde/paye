// Short plain-language explanations of Kenya's statutory payroll
// deductions, keyed by the codes the backend returns in
// statutory_deductions[].code (see migrations/002_seed_kenya_payroll_rules.sql).
export const DEDUCTION_INFO: Record<string, string> = {
  PAYE: "Pay As You Earn - income tax deducted directly from your salary and remitted to KRA by your employer.",
  NSSF: "National Social Security Fund - a mandatory pension contribution that builds your retirement savings.",
  SHIF: "Social Health Insurance Fund - Kenya's mandatory health cover contribution, replacing the old NHIF.",
  NHIF: "National Hospital Insurance Fund - Kenya's former mandatory health cover contribution, now replaced by SHIF.",
  AHL: "Affordable Housing Levy - a mandatory contribution toward Kenya's affordable housing program.",
};

export function deductionInfoFor(code: string | null | undefined): string | null {
  const normalizedCode = String(code ?? "").trim().toUpperCase();
  return normalizedCode ? DEDUCTION_INFO[normalizedCode] ?? null : null;
}
