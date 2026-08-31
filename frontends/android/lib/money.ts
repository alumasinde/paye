export function kes(value: string | number | null | undefined): string {
  const n = Number(value);
  return new Intl.NumberFormat("en-KE", {
    style: "currency",
    currency: "KES",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(Number.isFinite(n) ? n : 0);
}

// Formats a raw amount string with thousands separators as the person
// types (e.g. "100000" -> "100,000", "100000.5" -> "100,000.5").
// Runtime values are guarded because text input state should never crash
// the UI if a value is temporarily null or undefined.
export function formatAmountInput(raw: string | null | undefined): string {
  let cleaned = String(raw ?? "").replace(/[^0-9.]/g, "");
  const firstDot = cleaned.indexOf(".");

  if (firstDot !== -1) {
    cleaned =
      cleaned.slice(0, firstDot + 1) +
      cleaned.slice(firstDot + 1).replace(/\\./g, "");
  }

  const [whole = "", decimal] = cleaned.split(".");
  const trimmedWhole = whole.replace(/^0+(?=\d)/, "");
  const withCommas = trimmedWhole.replace(/\B(?=(\d{3})+(?!\d))/g, ",");

  if (decimal === undefined) return withCommas;
  return `${withCommas}.${decimal.slice(0, 2)}`;
}
