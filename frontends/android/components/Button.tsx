import { ActivityIndicator, GestureResponderEvent, Pressable, StyleSheet, Text } from "react-native";
import { colors, radius } from "../lib/theme";

type Props = {
  label: string;
  onPress: (event: GestureResponderEvent) => void;
  loading?: boolean;
  loadingLabel?: string;
  disabled?: boolean;
  variant?: "primary" | "secondary";
};

export function Button({ label, onPress, loading, loadingLabel, disabled, variant = "primary" }: Props) {
  const isDisabled = Boolean(disabled || loading);
  return (
    <Pressable
      accessibilityRole="button"
      onPress={onPress}
      disabled={isDisabled}
      style={({ pressed }) => [
        s.base,
        variant === "secondary" ? s.secondary : s.primary,
        isDisabled ? s.disabled : null,
        pressed && !isDisabled ? s.pressed : null,
      ]}
    >
      {loading ? <ActivityIndicator color={variant === "secondary" ? colors.primary : colors.white} /> : null}
      <Text style={[s.text, variant === "secondary" ? s.textSecondary : null]}>{loading ? (loadingLabel ?? label) : label}</Text>
    </Pressable>
  );
}

const s = StyleSheet.create({
  base: { height: 56, borderRadius: radius.md, alignItems: "center", justifyContent: "center", flexDirection: "row", gap: 10 },
  primary: { backgroundColor: colors.primary },
  secondary: { backgroundColor: colors.white, borderWidth: 1, borderColor: colors.borderStrong },
  disabled: { opacity: 0.72 },
  pressed: { opacity: 0.9 },
  text: { color: colors.white, fontWeight: "900", fontSize: 16 },
  textSecondary: { color: colors.primary },
});
