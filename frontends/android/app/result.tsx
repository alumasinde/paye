import { useState } from "react";
import { Alert, Share, StyleSheet, Text, View } from "react-native";
import * as Clipboard from "expo-clipboard";
import { useLocalSearchParams, router } from "expo-router";
import { Screen } from "../components/Screen";
import { Card } from "../components/Card";
import { Button } from "../components/Button";
import { TextField } from "../components/TextField";
import { ResultRow } from "../components/ResultRow";
import { MoneyText } from "../components/MoneyText";
import { Collapsible } from "../components/Collapsible";
import { AccountBar } from "../components/AccountBar";
import type { Calculation } from "../lib/types";
import { colors } from "../lib/theme";
import { buildShareSummary } from "../lib/shareSummary";
import { deductionInfoFor } from "../lib/deductionInfo";
import { useSession } from "../lib/useSession";
import { saveCalculation } from "../lib/accountApi";
import { APIClientError } from "../lib/httpClient";

type SaveState = "idle" | "saving" | "saved" | "error";

export default function Result() {
  const params = useLocalSearchParams<{ data?: string }>();
  const [copied, setCopied] = useState(false);
  const { session, loading: sessionLoading } = useSession();
  const [label, setLabel] = useState("");
  const [saveState, setSaveState] = useState<SaveState>("idle");
  const [saveError, setSaveError] = useState("");

  let data: Calculation | null = null;
  try {
    data = params.data ? (JSON.parse(params.data) as Calculation) : null;
  } catch {
    data = null;
  }

  if (!data) {
    return (
      <Screen contentContainerStyle={s.emptyPage}>
        <Text style={s.emptyTitle}>Calculation unavailable</Text>
        <Text style={s.emptyText}>Please return to the calculator and try again.</Text>
        <Button label="Back to calculator" onPress={() => router.replace("/calculator")} />
      </Screen>
    );
  }

  const summary = buildShareSummary(data);
  const calculation = data;

  async function share() {
    try {
      await Share.share({ message: summary });
    } catch {
      Alert.alert("Couldn't share", "Please try again.");
    }
  }

  async function copy() {
    await Clipboard.setStringAsync(summary);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  async function save() {
    setSaveState("saving");
    setSaveError("");
    try {
      await saveCalculation(calculation, label);
      setSaveState("saved");
    } catch (error) {
      setSaveState("error");
      setSaveError(error instanceof APIClientError ? error.message : "Could not save this calculation.");
    }
  }

  return (
    <Screen>
      <AccountBar />
      <Text style={s.back} onPress={() => router.replace("/calculator")}>
        ← New calculation
      </Text>
      <Text style={s.brand}>Budget254</Text>
      <Text style={s.title}>Your salary result</Text>
      <Text style={s.date}>Calculated for {data.calculation_date}</Text>

      <View style={s.hero}>
        <Text style={s.heroLabel}>Estimated take-home pay</Text>
        <MoneyText value={data.net_salary} style={s.heroValue} />
        <Text style={s.heroSub}>After statutory and selected custom deductions</Text>
        <View style={s.heroActions}>
          <View style={s.heroAction}>
            <Button label="Share" onPress={share} variant="secondary" />
          </View>
          <View style={s.heroAction}>
            <Button label={copied ? "Copied ✓" : "Copy"} onPress={copy} variant="secondary" />
          </View>
        </View>
      </View>

      {!sessionLoading ? (
        <Card style={s.saveCard}>
          {session ? (
            saveState === "saved" ? (
              <View style={s.savedRow}>
                <Text style={s.savedText}>Saved to your history ✓</Text>
                <Text style={s.link} onPress={() => router.push("/history")}>
                  View history
                </Text>
              </View>
            ) : (
              <>
                <Text style={s.saveTitle}>Save this calculation</Text>
                <Text style={s.saveSubtitle}>Give it a name so it's easy to find and compare later.</Text>
                <TextField
                  label="Name (optional)"
                  value={label}
                  onChangeText={setLabel}
                  placeholder="e.g. Job Offer A"
                  returnKeyType="done"
                />
                {saveState === "error" ? <Text style={s.saveError}>{saveError}</Text> : null}
                <Button label="Save" loading={saveState === "saving"} loadingLabel="Saving…" onPress={save} />
              </>
            )
          ) : (
            <>
              <Text style={s.saveTitle}>Save this calculation</Text>
              <Text style={s.saveSubtitle}>Log in to save this result and compare it against other offers later.</Text>
              <Button label="Log in to save" onPress={() => router.push("/login")} variant="secondary" />
            </>
          )}
        </Card>
      ) : null}

      <Text style={s.section}>Summary</Text>
      <Card style={s.rowCard}>
        <ResultRow name="Gross salary" value={data.gross_salary} />
        <ResultRow name="Taxable income" value={data.taxable_income} />
        <ResultRow name="PAYE before relief" value={data.paye_before_relief} />
        <ResultRow name="Tax relief" value={data.relief} />
        <ResultRow name="Final PAYE" value={data.paye} />
        <ResultRow name="Total deductions" value={data.total_deductions} />
        <ResultRow name="Net salary" value={data.net_salary} emphasis />
      </Card>

      <Text style={s.section}>Statutory deductions</Text>
      <Card style={s.rowCard}>
        {data.statutory_deductions.length ? (
          data.statutory_deductions.map((item, index) => {
            const info = deductionInfoFor(item.code);
            return (
              <View key={`${item.code}-${index}`}>
                <ResultRow name={item.name} value={item.amount} />
                {info ? (
                  <Collapsible title={`What is ${item.name}?`}>
                    <Text style={s.infoText}>{info}</Text>
                  </Collapsible>
                ) : null}
              </View>
            );
          })
        ) : (
          <Text style={s.emptyList}>No statutory deductions returned for this calculation.</Text>
        )}
      </Card>

      {data.custom_deductions.length ? (
        <>
          <Text style={s.section}>Custom deductions</Text>
          <Card style={s.rowCard}>
            {data.custom_deductions.map((item, index) => (
              <ResultRow key={`${item.name}-${index}`} name={item.name} value={item.amount} />
            ))}
          </Card>
        </>
      ) : null}

      {data.trace?.length ? (
        <>
          <Text style={s.section}>PAYE explanation</Text>
          <Card style={s.rowCard}>
            <Collapsible title="How your PAYE was calculated" subtitle={`${data.trace.length} tax band${data.trace.length === 1 ? "" : "s"} applied`}>
              {data.trace.map((item, index) => (
                <ResultRow key={index} name={`Band ${index + 1} · ${item.rate}%`} value={item.tax} />
              ))}
            </Collapsible>
          </Card>
        </>
      ) : null}

      <View style={s.action}>
        <Button label="Calculate again" onPress={() => router.replace("/calculator")} />
      </View>
    </Screen>
  );
}

const s = StyleSheet.create({
  back: { color: colors.accent, fontWeight: "800", fontSize: 14, marginBottom: 20 },
  brand: { fontSize: 16, fontWeight: "900", color: colors.text },
  title: { fontSize: 30, fontWeight: "900", color: colors.text, marginTop: 8 },
  date: { color: colors.textMuted, marginTop: 5 },
  hero: { backgroundColor: colors.primary, borderRadius: 22, padding: 22, marginTop: 20 },
  heroLabel: { color: "#D0D5DD", fontSize: 13, fontWeight: "700" },
  heroValue: { color: colors.white, fontSize: 34, fontWeight: "900", marginTop: 8 },
  heroSub: { color: "#98A2B3", fontSize: 12, lineHeight: 18, marginTop: 8 },
  heroActions: { flexDirection: "row", gap: 10, marginTop: 16 },
  heroAction: { flex: 1 },
  saveCard: { marginTop: 16 },
  saveTitle: { fontSize: 16, fontWeight: "900", color: colors.text },
  saveSubtitle: { fontSize: 13, lineHeight: 19, color: colors.textMuted, marginTop: 4, marginBottom: 14 },
  saveError: { fontSize: 12, color: colors.danger, marginBottom: 10 },
  savedRow: { flexDirection: "row", justifyContent: "space-between", alignItems: "center" },
  savedText: { fontSize: 14, fontWeight: "800", color: colors.text },
  link: { fontSize: 13, fontWeight: "700", color: colors.accent },
  section: { fontSize: 18, fontWeight: "900", color: colors.text, marginTop: 24, marginBottom: 8 },
  rowCard: { paddingHorizontal: 14, paddingVertical: 0 },
  infoText: { fontSize: 13, lineHeight: 19, color: colors.textSubtle },
  emptyList: { color: colors.textMuted, paddingVertical: 14, fontSize: 13 },
  action: { marginTop: 28 },
  emptyPage: { flexGrow: 1, justifyContent: "center" },
  emptyTitle: { fontSize: 24, fontWeight: "900", color: colors.text },
  emptyText: { color: colors.textMuted, lineHeight: 21, marginTop: 8, marginBottom: 20 },
});
