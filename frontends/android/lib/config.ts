/**
 * API configuration.
 *
 * For physical phone testing, EXPO_PUBLIC_API_BASE_URL must point to the
 * laptop's LAN IPv4 address, not localhost or 10.0.2.2.
 */
const rawBaseURL = process.env.EXPO_PUBLIC_API_BASE_URL?.trim() || "";

export const API_BASE_URL = rawBaseURL.replace(/\/$/, "");
export const API_V1 = API_BASE_URL ? `${API_BASE_URL}/api/v1` : "";

export function assertAPIConfigured(): string | null {
  if (!API_BASE_URL) {
    return "API URL is not configured. Set EXPO_PUBLIC_API_BASE_URL in your .env file.";
  }
  if (!/^https?:\/\//i.test(API_BASE_URL)) {
    return "API URL must start with http:// or https://.";
  }
  return null;
}
