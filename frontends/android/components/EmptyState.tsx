import { StyleSheet, Text, View } from "react-native";
import { colors } from "../lib/theme";

export function EmptyState({ title, message }: { title: string; message?: string }) {
  return (
    <View style={s.wrap}>
      <Text style={s.title}>{title}</Text>
      {message ? <Text style={s.message}>{message}</Text> : null}
    </View>
  );
}

const s = StyleSheet.create({
  wrap: { paddingVertical: 40, alignItems: "center" },
  title: { fontWeight: "800", fontSize: 15, color: colors.text },
  message: { color: colors.textMuted, fontSize: 13, marginTop: 6, textAlign: "center", paddingHorizontal: 20 },
});
