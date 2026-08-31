import { useCallback, useEffect, useState } from "react";
import { clearSession, getSession, type Session } from "./auth";

// undefined = not checked yet, null = checked and signed out.
export function useSession() {
  const [session, setSessionState] = useState<Session | null | undefined>(undefined);

  const refresh = useCallback(async () => {
    setSessionState(await getSession());
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const signOut = useCallback(async () => {
    await clearSession();
    setSessionState(null);
  }, []);

  return { session: session ?? null, loading: session === undefined, refresh, signOut };
}
