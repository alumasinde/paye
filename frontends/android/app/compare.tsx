import { StyleSheet, Text, View } from "react-native";
import { useLocalSearchParams, router } from "expo-router";
import { Screen } from "../components/Screen";
import { Button } from "../components/Button";
import { ComparisonTable } from "../components/ComparisonTable";
import type { SavedCalculation } from "../lib/accountApi";
import { colors } from "../lib/theme";

export default function Compare() {
  const params = useLocalSearchParams<{ items?: string }>();
  let items: SavedCalculation[] = [];
  try {
    items = params.items ? (JSON.parse(params.items) as SavedCalculation[]) : [];
  } catch {
    items = [];
  }

  if (items.length < 2) {
    return (
      <Screen contentContainerStyle={s.emptyPage}>
        <Text style={s.emptyTitle}>Nothing to compare</Text>
        <Text style={s.emptyText}>Go back to your history and select at least two saved calculations.</Text>
        <Button label="Back to history" onPress={() => router.replace("/history")} />
      </Screen>
    );
  }

  return (
    <Screen>
      <Text style={s.back} onPress={() => router.replace("/history")}>
        ← History
      </Text>
      <Text style={s.title}>Compare salaries</Text>
      <Text style={s.subtitle}>Comparing {items.length} saved calculations, side by side.</Text>

      <ComparisonTable items={items} />

      <View style={s.action}>
        <Button label="Back to history" onPress={() => router.replace("/history")} />
      </View>
    </Screen>
  );
}

const s = StyleSheet.create({
  back: { color: colors.accent, fontWeight: "800", fontSize: 14, marginBottom: 20 },
  title: { fontSize: 30, fontWeight: "900", color: colors.text },
  subtitle: { fontSize: 14, color: colors.textMuted, marginTop: 6, marginBottom: 20 },
  action: { marginTop: 24 },
  emptyPage: { flexGrow: 1, justifyContent: "center" },
  emptyTitle: { fontSize: 24, fontWeight: "900", color: colors.text },
  emptyText: { color: colors.textMuted, lineHeight: 21, marginTop: 8, marginBottom: 20 },
});
