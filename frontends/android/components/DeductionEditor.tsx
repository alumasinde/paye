import { View, Text, TextInput, Pressable, StyleSheet } from "react-native";
import type { CustomDeduction } from "../lib/types";
import { formatAmountInput } from "../lib/money";

type Props = {
  items: CustomDeduction[];
  onChange: (value: CustomDeduction[]) => void;
  error?: string;
};

export function DeductionEditor({ items, onChange, error }: Props) {
  const add = () => onChange([...items, { name: "", amount: "", type: "NET_PAY" }]);
  const update = (index: number, key: keyof CustomDeduction, value: string) => {
    onChange(items.map((item, current) => current === index ? { ...item, [key]: value } : item));
  };

  return (
    <View style={s.wrap}>
      <View style={s.headingRow}>
        <View style={s.headingCopy}>
          <Text style={s.title}>Custom deductions</Text>
          <Text style={s.sub}>Optional savings, welfare, SACCO, loans and other deductions.</Text>
        </View>
        <Text style={s.optional}>Optional</Text>
      </View>

      {items.map((item, index) => (
        <View key={`${index}-${item.name}`} style={s.card}>
          <TextInput value={item.name} onChangeText={(value) => update(index, "name", value)} placeholder="Name e.g. Savings" placeholderTextColor="#98A2B3" style={s.input} />
          <TextInput value={item.amount} onChangeText={(value) => update(index, "amount", formatAmountInput(value))} keyboardType="decimal-pad" placeholder="Amount" placeholderTextColor="#98A2B3" style={s.input} />
          <View style={s.actionRow}>
            <Pressable onPress={() => update(index, "type", "NET_PAY")} style={[s.type, item.type === "NET_PAY" ? s.active : null]}><Text style={[s.typeText, item.type === "NET_PAY" ? s.activeText : null]}>Net pay</Text></Pressable>
            <Pressable onPress={() => update(index, "type", "TAXABLE_INCOME")} style={[s.type, item.type === "TAXABLE_INCOME" ? s.active : null]}><Text style={[s.typeText, item.type === "TAXABLE_INCOME" ? s.activeText : null]}>Taxable income</Text></Pressable>
          </View>
          <Pressable onPress={() => onChange(items.filter((_, current) => current !== index))} hitSlop={8}><Text style={s.remove}>Remove deduction</Text></Pressable>
        </View>
      ))}

      {error ? <Text style={s.error}>{error}</Text> : null}
      <Pressable style={s.add} onPress={add}><Text style={s.addText}>+ Add deduction</Text></Pressable>
    </View>
  );
}

const s = StyleSheet.create({
  wrap: { marginBottom: 8 }, headingRow: { flexDirection: "row", justifyContent: "space-between", gap: 12 }, headingCopy: { flex: 1 },
  title: { fontSize: 17, fontWeight: "800", color: "#101828" }, sub: { fontSize: 12, lineHeight: 18, color: "#667085", marginTop: 4 }, optional: { fontSize: 11, fontWeight: "700", color: "#475467", backgroundColor: "#F2F4F7", paddingHorizontal: 8, paddingVertical: 4, borderRadius: 999 },
  card: { borderWidth: 1, borderColor: "#EAECF0", borderRadius: 14, padding: 12, marginTop: 12, backgroundColor: "#FFFFFF" }, input: { borderWidth: 1, borderColor: "#D0D5DD", borderRadius: 10, paddingHorizontal: 12, height: 48, fontSize: 15, color: "#101828", marginBottom: 10 },
  actionRow: { flexDirection: "row", flexWrap: "wrap", gap: 8, marginBottom: 12 }, type: { paddingHorizontal: 11, paddingVertical: 9, borderRadius: 9, borderWidth: 1, borderColor: "#D0D5DD" }, active: { backgroundColor: "#101828", borderColor: "#101828" }, typeText: { fontSize: 12, fontWeight: "700", color: "#475467" }, activeText: { color: "#FFFFFF" },
  remove: { color: "#B42318", fontWeight: "700", fontSize: 13 }, add: { alignSelf: "flex-start", paddingVertical: 14 }, addText: { fontWeight: "800", color: "#175CD3" }, error: { fontSize: 12, lineHeight: 17, color: "#B42318", marginTop: 8 },
});
