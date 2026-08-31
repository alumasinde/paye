import { StyleSheet, Text, View } from "react-native";
import { Button } from "./Button";
import { colors, radius } from "../lib/theme";

type Props = { title: string; message?: string; retryLabel?: string; onRetry?: () => void };

export function ErrorState({ title, message, retryLabel = "Try again", onRetry }: Props) {
  return (
    <View style={s.wrap}>
      <Text style={s.title}>{title}</Text>
      {message ? <Text style={s.message}>{message}</Text> : null}
      {onRetry ? (
        <View style={s.action}>
          <Button label={retryLabel} onPress={onRetry} variant="secondary" />
        </View>
      ) : null}
    </View>
  );
}

const s = StyleSheet.create({
  wrap: { backgroundColor: colors.dangerBg, borderWidth: 1, borderColor: colors.dangerBorder, borderRadius: radius.lg, padding: 16 },
  title: { color: colors.danger, fontWeight: "800", fontSize: 14 },
  message: { color: colors.danger, fontSize: 13, lineHeight: 19, marginTop: 4 },
  action: { marginTop: 12 },
});
