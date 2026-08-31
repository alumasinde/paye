import * as SecureStore from "expo-secure-store";

const KEY = "budget254.auth.v1";

export type Session = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: { id: string; email: string; first_name: string; last_name: string };
};

// SecureStore can throw on some Android keystores (e.g. a corrupted
// keystore, or certain first-run edge cases) - treat that as "no
// session" rather than letting it bubble up and crash whichever screen
// happened to check for a logged-in user.
export async function getSession(): Promise<Session | null> {
  try {
    const raw = await SecureStore.getItemAsync(KEY);
    return raw ? (JSON.parse(raw) as Session) : null;
  } catch {
    return null;
  }
}

export async function setSession(session: Session): Promise<void> {
  await SecureStore.setItemAsync(KEY, JSON.stringify(session));
}

export async function clearSession(): Promise<void> {
  try {
    await SecureStore.deleteItemAsync(KEY);
  } catch {
    // Best effort - if the keystore can't delete it, there's nothing
    // more the caller can do; treat the session as gone either way.
  }
}
