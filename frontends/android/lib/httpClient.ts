import { API_V1, assertAPIConfigured } from "./config";
import { getSession } from "./auth";
import type { APIError } from "./types";

export class APIClientError extends Error {
  constructor(
    public status: number,
    public body: APIError,
  ) {
    super(body.message || "Request failed");
    this.name = "APIClientError";
  }
}

function fallbackError(status: number, message: string): APIError {
  return { code: `HTTP_${status}`, message };
}

type RequestOptions = {
  method?: string;
  body?: unknown;
  /** Attach the signed-in user's bearer token. Throws APIClientError(401) if there is no session. */
  auth?: boolean;
  timeoutMs?: number;
};

// Every network call in the app - the public PAYE calculator and every
// authenticated account endpoint - goes through this one function, so
// timeout handling, JSON parsing, auth headers, and error shape are
// identical everywhere instead of each screen/service reinventing them
// slightly differently (which is what lib/api.ts and lib/accountApi.ts
// used to do independently).
export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const configurationError = assertAPIConfigured();
  if (configurationError) {
    throw new APIClientError(0, { code: "API_NOT_CONFIGURED", message: configurationError });
  }

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), options.timeoutMs ?? 15_000);

  try {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      Accept: "application/json",
    };

    if (options.auth) {
      const session = await getSession();
      if (!session) {
        throw new APIClientError(401, { code: "NOT_AUTHENTICATED", message: "Please log in to continue." });
      }
      headers.Authorization = `Bearer ${session.access_token}`;
    }

    const response = await fetch(`${API_V1}${path}`, {
      method: options.method ?? "GET",
      headers,
      body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
      signal: controller.signal,
    });

    if (response.status === 204) {
      return null as T;
    }

    const raw = await response.text();
    let body: unknown = null;
    if (raw) {
      try {
        body = JSON.parse(raw);
      } catch {
        body = null;
      }
    }

    if (!response.ok) {
      throw new APIClientError(
        response.status,
        (body as APIError) || fallbackError(response.status, "The server returned an invalid response."),
      );
    }

    if (body === null) {
      throw new APIClientError(response.status, fallbackError(response.status, "The server returned an empty response."));
    }

    return body as T;
  } catch (error) {
    if (error instanceof APIClientError) throw error;
    if (error instanceof Error && error.name === "AbortError") {
      throw new APIClientError(0, {
        code: "REQUEST_TIMEOUT",
        message: "The server took too long to respond. Check your connection and try again.",
      });
    }
    throw new APIClientError(0, {
      code: "NETWORK_ERROR",
      message: "Could not reach the Budget254 server. Check your Wi-Fi and API address.",
    });
  } finally {
    clearTimeout(timer);
  }
}
