import { useCallback, useEffect, useState } from "react";
import { Alert, Pressable, StyleSheet, Text, View } from "react-native";
import { router } from "expo-router";
import { Screen } from "../components/Screen";
import { Card } from "../components/Card";
import { Button } from "../components/Button";
import { TextField } from "../components/TextField";
import { LoadingState } from "../components/LoadingState";
import { ErrorState } from "../components/ErrorState";
import { EmptyState } from "../components/EmptyState";
import { MoneyText } from "../components/MoneyText";
import { AccountBar } from "../components/AccountBar";
import { history, removeCalculation, renameCalculation, type SavedCalculation } from "../lib/accountApi";
import { APIClientError } from "../lib/httpClient";
import { colors, radius } from "../lib/theme";

type State =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ready"; items: SavedCalculation[] };

const MAX_COMPARE = 3;

export default function History() {
  const [state, setState] = useState<State>({ status: "loading" });
  const [compareMode, setCompareMode] = useState(false);
  const [selected, setSelected] = useState<string[]>([]);
  const [renamingId, setRenamingId] = useState<string | null>(null);
  const [renameValue, setRenameValue] = useState("");

  const load = useCallback(async () => {
    setState({ status: "loading" });
    try {
      const result = await history();
      setState({ status: "ready", items: result.items });
    } catch (error) {
      setState({
        status: "error",
        message: error instanceof APIClientError ? error.message : "Could not load your saved calculations.",
      });
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  async function remove(id: string) {
    try {
      await removeCalculation(id);
      load();
    } catch (error) {
      Alert.alert("Couldn't delete", error instanceof APIClientError ? error.message : "Please try again.");
    }
  }

  function startRename(item: SavedCalculation) {
    setRenamingId(item.id);
    setRenameValue(item.label ?? "");
  }

  async function saveRename(id: string) {
    const trimmed = renameValue.trim();
    setRenamingId(null);
    if (state.status !== "ready") return;
    const previous = state.items;
    setState({ status: "ready", items: previous.map((item) => (item.id === id ? { ...item, label: trimmed || null } : item)) });
    try {
      await renameCalculation(id, trimmed || null);
    } catch (error) {
      setState({ status: "ready", items: previous });
      Alert.alert("Couldn't rename", error instanceof APIClientError ? error.message : "Please try again.");
    }
  }

  function toggleSelected(id: string) {
    setSelected((current) => {
      if (current.includes(id)) return current.filter((x) => x !== id);
      if (current.length >= MAX_COMPARE) return current;
      return [...current, id];
    });
  }

  function toggleCompareMode() {
    setCompareMode((v) => !v);
    setSelected([]);
    setRenamingId(null);
  }

  function goToCompare() {
    if (state.status !== "ready") return;
    const items = state.items.filter((item) => selected.includes(item.id));
    router.push({ pathname: "/compare", params: { items: JSON.stringify(items) } });
  }

  return (
    <Screen>
      <AccountBar />
      <View style={s.headerRow}>
        <Text style={s.title}>Saved calculations</Text>
        {state.status === "ready" && state.items.length >= 2 ? (
          <Text style={s.compareToggle} onPress={toggleCompareMode}>
            {compareMode ? "Cancel" : "Compare"}
          </Text>
        ) : null}
      </View>

      {state.status === "loading" ? <LoadingState label="Loading your history…" /> : null}
      {state.status === "error" ? <ErrorState title="Couldn't load history" message={state.message} onRetry={load} /> : null}
      {state.status === "ready" && state.items.length === 0 ? (
        <EmptyState title="No saved calculations yet" message="Save a result from the calculator to see it here." />
      ) : null}

      {state.status === "ready"
        ? state.items.map((item) => {
            const isSelected = selected.includes(item.id);
            const isRenaming = renamingId === item.id;
            return (
              <Pressable key={item.id} onPress={compareMode ? () => toggleSelected(item.id) : undefined} disabled={!compareMode}>
                <Card style={[s.card, compareMode && isSelected ? s.cardSelected : null]}>
                  {compareMode ? (
                    <View style={[s.checkbox, isSelected ? s.checkboxChecked : null]}>
                      {isSelected ? <Text style={s.checkboxMark}>✓</Text> : null}
                    </View>
                  ) : null}

                  {isRenaming ? (
                    <View>
                      <TextField label="Name" value={renameValue} onChangeText={setRenameValue} placeholder="e.g. Job Offer A" autoFocus />
                      <View style={s.renameActions}>
                        <View style={s.renameButton}>
                          <Button label="Cancel" variant="secondary" onPress={() => setRenamingId(null)} />
                        </View>
                        <View style={s.renameButton}>
                          <Button label="Save name" onPress={() => saveRename(item.id)} />
                        </View>
                      </View>
                    </View>
                  ) : (
                    <>
                      <View style={s.titleRow}>
                        <Text style={s.itemTitle}>{item.label ?? `Calculation for ${item.calculation_date}`}</Text>
                        {!compareMode ? (
                          <Text style={s.rename} onPress={() => startRename(item)}>
                            Rename
                          </Text>
                        ) : null}
                      </View>
                      {item.label ? <Text style={s.itemDate}>{item.calculation_date}</Text> : null}
                      <View style={s.row}>
                        <Text style={s.label}>Gross</Text>
                        <MoneyText value={item.gross_salary} style={s.value} />
                      </View>
                      <View style={s.row}>
                        <Text style={s.label}>Net</Text>
                        <MoneyText value={item.net_salary} style={s.valueStrong} />
                      </View>
                      {!compareMode ? (
                        <Pressable onPress={() => remove(item.id)} hitSlop={8}>
                          <Text style={s.remove}>Delete</Text>
                        </Pressable>
                      ) : null}
                    </>
                  )}
                </Card>
              </Pressable>
            );
          })
        : null}

      {compareMode && selected.length >= 2 ? (
        <View style={s.compareBar}>
          <Button label={`Compare (${selected.length})`} onPress={goToCompare} />
        </View>
      ) : null}
    </Screen>
  );
}

const s = StyleSheet.create({
  headerRow: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 20 },
  title: { fontSize: 30, fontWeight: "900", color: colors.text },
  compareToggle: { fontSize: 14, fontWeight: "800", color: colors.accent },
  card: { marginBottom: 12 },
  cardSelected: { borderColor: colors.accent, borderWidth: 2 },
  checkbox: { width: 22, height: 22, borderRadius: radius.sm, borderWidth: 2, borderColor: colors.borderStrong, alignItems: "center", justifyContent: "center", marginBottom: 12 },
  checkboxChecked: { backgroundColor: colors.accent, borderColor: colors.accent },
  checkboxMark: { color: colors.white, fontSize: 13, fontWeight: "900" },
  titleRow: { flexDirection: "row", justifyContent: "space-between", alignItems: "flex-start", gap: 10 },
  itemTitle: { flex: 1, fontWeight: "800", color: colors.text, fontSize: 15 },
  itemDate: { color: colors.textMuted, fontSize: 12, marginTop: 2 },
  rename: { fontSize: 12, fontWeight: "700", color: colors.accent },
  row: { flexDirection: "row", justifyContent: "space-between", marginTop: 8 },
  label: { color: colors.textMuted },
  value: { fontWeight: "700", color: colors.text },
  valueStrong: { fontWeight: "900", color: colors.text, fontSize: 16 },
  remove: { color: colors.danger, marginTop: 12, fontWeight: "700" },
  renameActions: { flexDirection: "row", gap: 10, marginTop: 4 },
  renameButton: { flex: 1 },
  compareBar: { marginTop: 8, marginBottom: 12 },
});
