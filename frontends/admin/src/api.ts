const BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";
const TOKEN_KEY = "budget254.admin.access";
const REFRESH_KEY = "budget254.admin.refresh";

let token: string | null = sessionStorage.getItem(TOKEN_KEY);
let refreshToken: string | null = sessionStorage.getItem(REFRESH_KEY);
let refreshing: Promise<boolean> | null = null;

export function setToken(v: string, refresh?: string) {
  token = v;
  sessionStorage.setItem(TOKEN_KEY, v);
  if (refresh) {
    refreshToken = refresh;
    sessionStorage.setItem(REFRESH_KEY, refresh);
  }
}

export function clearToken() {
  token = null;
  refreshToken = null;
  sessionStorage.removeItem(TOKEN_KEY);
  sessionStorage.removeItem(REFRESH_KEY);
}

export function isAuthenticated(): boolean {
  return !!token;
}

async function tryRefresh(): Promise<boolean> {
  if (!refreshToken) return false;
  if (!refreshing) {
    refreshing = fetch(BASE + "/admin/auth/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
      .then(async (r) => {
        if (!r.ok) return false;
        const b = await r.json();
        setToken(b.tokens.access_token, b.tokens.refresh_token);
        return true;
      })
      .catch(() => false)
      .finally(() => {
        refreshing = null;
      });
  }
  return refreshing;
}

export async function api(path: string, init: RequestInit = {}, _retried = false): Promise<any> {
  const r = await fetch(BASE + path, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init.headers || {}),
    },
  });
  if (r.status === 401 && !_retried) {
    // Access token expired or invalid - try the refresh token once before
    // giving up and forcing a re-login.
    const refreshed = await tryRefresh();
    if (refreshed) return api(path, init, true);
    clearToken();
  }
  const b = r.status === 204 ? null : await r.json();
  if (!r.ok) throw new Error(b?.message || "Request failed");
  return b;
}
