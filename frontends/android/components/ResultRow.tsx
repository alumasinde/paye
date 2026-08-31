import { StyleSheet, Text, View } from "react-native";
import { MoneyText } from "./MoneyText";
import { colors } from "../lib/theme";

type Props = { name: string; value: string; emphasis?: boolean };

// Extracted from result.tsx, where it used to be a locally-defined
// component redeclared inline - promoted here so any future screen
// showing a labelled money amount reuses it.
export function ResultRow({ name, value, emphasis = false }: Props) {
  return (
    <View style={s.row}>
      <Text style={[s.name, emphasis ? s.nameStrong : null]}>{name}</Text>
      <MoneyText value={value} style={[s.value, emphasis ? s.valueStrong : null]} />
    </View>
  );
}

const s = StyleSheet.create({
  row: { flexDirection: "row", justifyContent: "space-between", gap: 12, paddingVertical: 14, borderBottomWidth: 1, borderBottomColor: "#F2F4F7" },
  name: { flex: 1, color: colors.textSubtle, fontSize: 14 },
  nameStrong: { color: colors.text, fontWeight: "900" },
  value: { fontWeight: "800", color: colors.text, marginLeft: 8 },
  valueStrong: { fontSize: 16 },
});
