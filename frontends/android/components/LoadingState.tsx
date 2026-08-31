import { ActivityIndicator, StyleSheet, Text, View } from "react-native";
import { colors } from "../lib/theme";

export function LoadingState({ label = "Loading…" }: { label?: string }) {
  return (
    <View style={s.wrap}>
      <ActivityIndicator color={colors.primary} />
      <Text style={s.label}>{label}</Text>
    </View>
  );
}

const s = StyleSheet.create({
  wrap: { paddingVertical: 40, alignItems: "center", gap: 10 },
  label: { color: colors.textMuted, fontSize: 13 },
});
