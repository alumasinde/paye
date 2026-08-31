import { StyleSheet, Text, View } from "react-native";
import { useNetworkStatus } from "../lib/useNetworkStatus";
import { colors, radius } from "../lib/theme";

// Shows a small banner when the device has no usable internet
// connection. Only renders once we're confident the device is actually
// offline (online === false), not during the brief unknown state at
// startup (online === null).
export function NetworkBanner() {
  const online = useNetworkStatus();
  if (online !== false) return null;

  return (
    <View style={s.wrap}>
      <Text style={s.text}>You're offline. Calculations need an internet connection.</Text>
    </View>
  );
}

const s = StyleSheet.create({
  wrap: { backgroundColor: colors.dangerBg, borderWidth: 1, borderColor: colors.dangerBorder, borderRadius: radius.md, padding: 10, marginBottom: 16 },
  text: { color: colors.danger, fontSize: 12, fontWeight: "700", textAlign: "center" },
});
