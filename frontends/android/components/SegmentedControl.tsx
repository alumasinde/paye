import { Pressable, StyleSheet, Text, View } from "react-native";
import { colors, radius } from "../lib/theme";

type Option<T extends string> = { value: T; label: string };

type Props<T extends string> = {
  label?: string;
  options: Option<T>[];
  value: T;
  onChange: (value: T) => void;
};

// Small two-or-more-way toggle, used for the monthly/annual salary
// switch. Generic over the option value type so it can be reused for any
// future mutually-exclusive choice without a new component.
export function SegmentedControl<T extends string>({ label, options, value, onChange }: Props<T>) {
  return (
    <View style={s.wrap}>
      {label ? <Text style={s.label}>{label}</Text> : null}
      <View style={s.track}>
        {options.map((option) => {
          const active = option.value === value;
          return (
            <Pressable
              key={option.value}
              onPress={() => onChange(option.value)}
              style={[s.segment, active ? s.segmentActive : null]}
              accessibilityRole="button"
              accessibilityState={{ selected: active }}
            >
              <Text style={[s.segmentText, active ? s.segmentTextActive : null]}>{option.label}</Text>
            </Pressable>
          );
        })}
      </View>
    </View>
  );
}

const s = StyleSheet.create({
  wrap: { marginBottom: 18 },
  label: { fontSize: 14, fontWeight: "700", color: colors.text, marginBottom: 8 },
  track: { flexDirection: "row", backgroundColor: "#F2F4F7", borderRadius: radius.md, padding: 4, gap: 4 },
  segment: { flex: 1, paddingVertical: 10, borderRadius: radius.sm, alignItems: "center" },
  segmentActive: { backgroundColor: colors.white, shadowColor: colors.text, shadowOpacity: 0.08, shadowRadius: 4, elevation: 1 },
  segmentText: { fontSize: 13, fontWeight: "700", color: colors.textMuted },
  segmentTextActive: { color: colors.text },
});
