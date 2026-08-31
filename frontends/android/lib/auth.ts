import { Platform } from "react-native";
import * as SecureStore from "expo-secure-store";

const KEY = "budget254.auth.v1";

export type Session = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: { id: string; email: string; first_name: string; last_name: string };
};

function isWeb(): boolean {
  return Platform.OS === "web";
}

function getWebStorage(): Storage | null {
  if (typeof window === "undefined") return null;

  try {
    return window.localStorage;
  } catch {
    return null;
  }
}

// expo-secure-store provides encrypted storage on Android/iOS. The web
// implementation does not provide the same SecureStore API, so browser
// testing uses localStorage while the installed mobile app keeps using
// SecureStore.
export async function getSession(): Promise<Session | null> {
  try {
    const raw = isWeb()
      ? getWebStorage()?.getItem(KEY) ?? null
      : await SecureStore.getItemAsync(KEY);

    return raw ? (JSON.parse(raw) as Session) : null;
  } catch {
    return null;
  }
}

export async function setSession(session: Session): Promise<void> {
  const value = JSON.stringify(session);

  if (isWeb()) {
    const storage = getWebStorage();
    if (!storage) {
      throw new Error("Browser storage is unavailable.");
    }
    storage.setItem(KEY, value);
    return;
  }

  await SecureStore.setItemAsync(KEY, value);
}

export async function clearSession(): Promise<void> {
  try {
    if (isWeb()) {
      getWebStorage()?.removeItem(KEY);
      return;
    }

    await SecureStore.deleteItemAsync(KEY);
  } catch {
    // Best effort cleanup.
  }
}
