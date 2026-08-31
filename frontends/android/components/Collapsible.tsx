import { ReactNode, useState } from "react";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { colors } from "../lib/theme";

type Props = { title: string; subtitle?: string; children: ReactNode; defaultOpen?: boolean };

// Generic expand/collapse section. Used for the PAYE band explanation and
// the "what is NSSF/SHIF/AHL" educational blurbs - content that's useful
// but shouldn't be dumped in front of the person by default.
export function Collapsible({ title, subtitle, children, defaultOpen = false }: Props) {
  const [open, setOpen] = useState(defaultOpen);
  return (
    <View>
      <Pressable style={s.header} onPress={() => setOpen((v) => !v)} accessibilityRole="button">
        <View style={s.headerText}>
          <Text style={s.title}>{title}</Text>
          {subtitle ? <Text style={s.subtitle}>{subtitle}</Text> : null}
        </View>
        <Text style={s.chevron}>{open ? "−" : "+"}</Text>
      </Pressable>
      {open ? <View style={s.body}>{children}</View> : null}
    </View>
  );
}

const s = StyleSheet.create({
  header: { flexDirection: "row", alignItems: "center", justifyContent: "space-between", paddingVertical: 14 },
  headerText: { flex: 1, paddingRight: 12 },
  title: { fontSize: 14, fontWeight: "800", color: colors.text },
  subtitle: { fontSize: 12, lineHeight: 17, color: colors.textMuted, marginTop: 2 },
  chevron: { fontSize: 20, fontWeight: "700", color: colors.accent, width: 24, textAlign: "center" },
  body: { paddingBottom: 14 },
});
