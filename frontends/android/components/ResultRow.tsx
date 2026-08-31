import { StyleSheet, Text, View } from "react-native";
import { MoneyText } from "./MoneyText";
import { colors } from "../lib/theme";

type Props = {
  name: string | null | undefined;
  value: string | number | null | undefined;
  emphasis?: boolean;
};

export function ResultRow({ name, value, emphasis = false }: Props) {
  const safeName =
    typeof name === "string" && name.trim().length > 0
      ? name
      : "Not available";

  return (
    <View style={s.row}>
      <Text style={[s.name, emphasis ? s.nameStrong : undefined]}>
        {safeName}
      </Text>
      <MoneyText
        value={value}
        style={[s.value, emphasis ? s.valueStrong : undefined]}
      />
    </View>
  );
}

const s = StyleSheet.create({
  row: {
    flexDirection: "row",
    justifyContent: "space-between",
    gap: 12,
    paddingVertical: 14,
    borderBottomWidth: 1,
    borderBottomColor: "#F2F4F7",
  },
  name: { flex: 1, color: colors.textSubtle, fontSize: 14 },
  nameStrong: { color: colors.text, fontWeight: "900" },
  value: { fontWeight: "800", color: colors.text, marginLeft: 8 },
  valueStrong: { fontSize: 16 },
});
