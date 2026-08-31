export function kes(value: string | number) {
  const n = Number(value);
  return new Intl.NumberFormat("en-KE", { style: "currency", currency: "KES", minimumFractionDigits: 2 }).format(
    Number.isFinite(n) ? n : 0,
  );
}

// Formats a raw amount string with thousands separators as the person
// types (e.g. "100000" -> "100,000", "100000.5" -> "100,000.5"). Keeps at
// most one decimal point and up to 2 decimal digits; anything that isn't
// a digit or the first decimal point is dropped. This only touches
// presentation - the underlying numeric value (via normaliseAmount in
// validation.ts) is unaffected by the commas.
export function formatAmountInput(raw: string): string {
  let cleaned = raw.replace(/[^0-9.]/g, "");
  const firstDot = cleaned.indexOf(".");
  if (firstDot !== -1) {
    cleaned = cleaned.slice(0, firstDot + 1) + cleaned.slice(firstDot + 1).replace(/\./g, "");
  }
  const [whole, decimal] = cleaned.split(".");
  const trimmedWhole = whole.replace(/^0+(?=\d)/, "");
  const withCommas = trimmedWhole.replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  if (decimal === undefined) return withCommas;
  return `${withCommas}.${decimal.slice(0, 2)}`;
}
