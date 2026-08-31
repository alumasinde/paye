import { View, ViewProps, StyleSheet } from "react-native";
import { colors, radius } from "../lib/theme";

export function Card({ style, ...rest }: ViewProps) {
  return <View style={[s.card, style]} {...rest} />;
}

const s = StyleSheet.create({
  card: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: radius.xl,
    padding: 16,
    shadowColor: colors.text,
    shadowOpacity: 0.04,
    shadowRadius: 12,
    elevation: 2,
  },
});
