import { useEffect, useState } from "react";
import NetInfo from "@react-native-community/netinfo";

// True/false once NetInfo has reported; null briefly at startup before
// the first event arrives. Treat null as "assume online" everywhere it's
// used for a UI decision (e.g. disabling a button) so a slow first event
// doesn't block the person from trying - the network request itself will
// still fail cleanly and show a retry if there's genuinely no connection.
export function useNetworkStatus(): boolean | null {
  const [online, setOnline] = useState<boolean | null>(null);

  useEffect(() => {
    return NetInfo.addEventListener((state) => {
      setOnline(state.isConnected !== false && state.isInternetReachable !== false);
    });
  }, []);

  return online;
}
