import { request } from "./httpClient";
import { setSession } from "./auth";
import type { Calculation } from "./types";

type AuthResponse = {
  user: { id: string; email: string; first_name: string; last_name: string };
  tokens: { access_token: string; refresh_token: string; expires_in: number };
};

export type RegisterInput = { first_name: string; last_name: string; email: string; password: string };
export type LoginInput = { email: string; password: string };

// Matches saved.repository.Snapshot's JSON shape exactly: the full
// calculation is nested under `payload` (it's stored as one JSON blob
// server-side), not flattened onto the saved-calculation record. gross/net
// are duplicated at the top level for cheap sorting/display without
// parsing payload.
export type SavedCalculation = {
  id: string;
  label: string | null;
  calculation_date: string;
  gross_salary: string;
  net_salary: string;
  payload: Calculation;
  created_at: string;
};

async function authenticate(path: string, payload: RegisterInput | LoginInput): Promise<AuthResponse> {
  const body = await request<AuthResponse>(path, { method: "POST", body: payload });
  await setSession({ ...body.tokens, user: body.user });
  return body;
}

export function register(payload: RegisterInput) {
  return authenticate("/auth/register", payload);
}

export function login(payload: LoginInput) {
  return authenticate("/auth/login", payload);
}

export function saveCalculation(calculation: Calculation, label?: string) {
  const body = label && label.trim() ? { ...calculation, label: label.trim() } : calculation;
  return request<{ id: string }>("/calculations", { method: "POST", body, auth: true });
}

export function history() {
  return request<{ items: SavedCalculation[] }>("/calculations", { auth: true });
}

export function renameCalculation(id: string, label: string | null) {
  return request<null>(`/calculations/${id}`, { method: "PATCH", body: { label }, auth: true });
}

export function removeCalculation(id: string) {
  return request<null>(`/calculations/${id}`, { method: "DELETE", auth: true });
}
