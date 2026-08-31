import { request } from "./httpClient";
import type { CalculateRequest, Calculation } from "./types";

export { APIClientError } from "./httpClient";

export async function calculatePAYE(payload: CalculateRequest): Promise<Calculation> {
  return request<Calculation>("/calculator/paye", { method: "POST", body: payload });
}
