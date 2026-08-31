import { TextInput, View, Text, StyleSheet } from "react-native";

type Props = { value: string; onChange: (value: string) => void; error?: string };

export function DateInput({ value, onChange, error }: Props) {
  return (
    <View style={s.wrap}>
      <Text style={s.label}>Calculation date</Text>
      <TextInput
        value={value}
        onChangeText={onChange}
        placeholder="YYYY-MM-DD"
        placeholderTextColor="#98A2B3"
        autoCapitalize="none"
        autoCorrect={false}
        maxLength={10}
        style={[s.input, error ? s.inputError : null]}
        accessibilityLabel="Calculation date"
      />
      {error ? <Text style={s.error}>{error}</Text> : <Text style={s.hint}>Use any supported historical date from 2022 onward.</Text>}
    </View>
  );
}

const s = StyleSheet.create({
  wrap: { marginBottom: 18 },
  label: { fontSize: 14, fontWeight: "700", color: "#101828", marginBottom: 8 },
  input: { borderWidth: 1, borderColor: "#D0D5DD", borderRadius: 14, paddingHorizontal: 14, height: 56, fontSize: 16, color: "#101828", backgroundColor: "#FFFFFF" },
  inputError: { borderColor: "#D92D20" },
  hint: { fontSize: 12, lineHeight: 17, marginTop: 6, color: "#667085" },
  error: { fontSize: 12, lineHeight: 17, marginTop: 6, color: "#B42318" },
});
