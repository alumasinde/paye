import { Text, TextInput, TextInputProps, View, StyleSheet } from "react-native";
import { colors, radius } from "../lib/theme";

type Props = TextInputProps & { label: string; error?: string; hint?: string };

// Generic labelled text input with error/hint text underneath, matching
// AmountInput's visual style. Used by login/register instead of each
// screen redeclaring its own TextInput + label styles.
export function TextField({ label, error, hint, style, ...rest }: Props) {
  return (
    <View style={s.wrap}>
      <Text style={s.label}>{label}</Text>
      <TextInput
        placeholderTextColor={colors.placeholder}
        style={[s.input, error ? s.inputError : null, style]}
        accessibilityLabel={label}
        {...rest}
      />
      {error ? <Text style={s.error}>{error}</Text> : hint ? <Text style={s.hint}>{hint}</Text> : null}
    </View>
  );
}

const s = StyleSheet.create({
  wrap: { marginBottom: 18 },
  label: { fontSize: 14, fontWeight: "700", color: colors.text, marginBottom: 8 },
  input: {
    borderWidth: 1,
    borderColor: colors.borderStrong,
    borderRadius: radius.md,
    paddingHorizontal: 14,
    height: 56,
    fontSize: 16,
    color: colors.text,
    backgroundColor: colors.surface,
  },
  inputError: { borderColor: colors.danger },
  hint: { fontSize: 12, lineHeight: 17, marginTop: 6, color: colors.textMuted },
  error: { fontSize: 12, lineHeight: 17, marginTop: 6, color: colors.danger },
});
