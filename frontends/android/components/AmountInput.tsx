import { TextInput, View, Text, StyleSheet } from "react-native";
import { formatAmountInput } from "../lib/money";

type Props = {
  label: string;
  value: string;
  onChange: (value: string) => void;
  error?: string;
};

export function AmountInput({ label, value, onChange, error }: Props) {
  return (
    <View style={s.wrap}>
      <Text style={s.label}>{label}</Text>
      <View style={[s.box, error ? s.boxError : null]}>
        <Text style={s.prefix}>KES</Text>
        <TextInput
          value={formatAmountInput(value)}
          onChangeText={(next) => onChange(formatAmountInput(next))}
          keyboardType="decimal-pad"
          placeholder="e.g. 100,000"
          placeholderTextColor="#98A2B3"
          style={s.input}
          accessibilityLabel={label}
          returnKeyType="done"
        />
      </View>
      {error ? <Text style={s.error}>{error}</Text> : <Text style={s.hint}>Enter your total gross salary before deductions.</Text>}
    </View>
  );
}

const s = StyleSheet.create({
  wrap: { marginBottom: 18 },
  label: { fontSize: 14, fontWeight: "700", color: "#101828", marginBottom: 8 },
  box: { flexDirection: "row", alignItems: "center", borderWidth: 1, borderColor: "#D0D5DD", borderRadius: 14, paddingHorizontal: 14, height: 56, backgroundColor: "#FFFFFF" },
  boxError: { borderColor: "#D92D20" },
  prefix: { fontWeight: "800", color: "#344054", marginRight: 10 },
  input: { flex: 1, fontSize: 18, fontWeight: "600", color: "#101828" },
  hint: { fontSize: 12, lineHeight: 17, marginTop: 6, color: "#667085" },
  error: { fontSize: 12, lineHeight: 17, marginTop: 6, color: "#B42318" },
});
