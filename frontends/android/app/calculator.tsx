import { useState } from "react";
import { Alert, Text, View, StyleSheet } from "react-native";
import { router } from "expo-router";
import { Screen } from "../components/Screen";
import { Card } from "../components/Card";
import { Button } from "../components/Button";
import { ErrorState } from "../components/ErrorState";
import { NetworkBanner } from "../components/NetworkBanner";
import { SegmentedControl } from "../components/SegmentedControl";
import { AccountBar } from "../components/AccountBar";
import { AmountInput } from "../components/AmountInput";
import { DateInput } from "../components/DatePicker";
import { DeductionEditor } from "../components/DeductionEditor";
import { calculatePAYE, APIClientError } from "../lib/api";
import type { CustomDeduction } from "../lib/types";
import { normaliseAmount, normaliseDeductions, validateCalculationForm, type CalculationFormErrors } from "../lib/validation";
import { useNetworkStatus } from "../lib/useNetworkStatus";
import { colors } from "../lib/theme";

type Period = "monthly" | "annual";

function todayISO() {
  return new Date().toISOString().slice(0, 10);
}

export default function Calculator() {
  const [period, setPeriod] = useState<Period>("monthly");
  const [gross, setGross] = useState("");
  const [date, setDate] = useState(todayISO());
  const [deductions, setDeductions] = useState<CustomDeduction[]>([]);
  const [errors, setErrors] = useState<CalculationFormErrors>({});
  const [loading, setLoading] = useState(false);
  const [serverError, setServerError] = useState("");
  const online = useNetworkStatus();

  function reset() {
    setPeriod("monthly");
    setGross("");
    setDate(todayISO());
    setDeductions([]);
    setErrors({});
    setServerError("");
  }

  async function submit() {
    setServerError("");
    const nextErrors = validateCalculationForm(gross, date, deductions);
    setErrors(nextErrors);
    if (Object.keys(nextErrors).length > 0) return;

    setLoading(true);
    try {
      const monthlyGross = period === "annual" ? String(Number(normaliseAmount(gross)) / 12) : normaliseAmount(gross);
      const result = await calculatePAYE({
        gross_salary: monthlyGross,
        calculation_date: date,
        explain: true,
        custom_deductions: normaliseDeductions(deductions),
      });
      router.push({ pathname: "/result", params: { data: JSON.stringify(result) } });
    } catch (error) {
      const message = error instanceof APIClientError ? error.message : "Please check your connection and details.";
      setServerError(message);
      Alert.alert("Calculation failed", message);
    } finally {
      setLoading(false);
    }
  }

  const offline = online === false;

  return (
    <Screen>
      <AccountBar />
      <View style={s.top}>
        <View style={s.logo}>
          <Text style={s.logoText}>B254</Text>
        </View>
        <Text style={s.brand}>Budget254</Text>
      </View>
      <Text style={s.title}>Know your take-home pay.</Text>
      <Text style={s.subtitle}>
        Calculate Kenyan PAYE, statutory deductions and net salary using the payroll rules applicable to your selected
        date.
      </Text>

      <View style={s.infoBanner}>
        <Text style={s.infoTitle}>No account required</Text>
        <Text style={s.infoText}>You can calculate as a guest. Historical dates from 2022 onward are supported.</Text>
      </View>

      <NetworkBanner />

      <Card style={s.card}>
        <View style={s.cardHeader}>
          <Text style={s.cardTitle}>Salary details</Text>
          <Text style={s.reset} onPress={reset}>
            Reset
          </Text>
        </View>

        <SegmentedControl
          label="Salary period"
          value={period}
          onChange={setPeriod}
          options={[
            { value: "monthly", label: "Monthly" },
            { value: "annual", label: "Annual" },
          ]}
        />

        <AmountInput
          label={period === "annual" ? "Gross annual salary" : "Gross monthly salary"}
          value={gross}
          onChange={(value) => {
            setGross(value);
            setErrors((current) => ({ ...current, gross: undefined }));
          }}
          error={errors.gross}
        />
        {period === "annual" ? <Text style={s.periodHint}>We'll divide this by 12 to calculate monthly PAYE.</Text> : null}

        <DateInput
          value={date}
          onChange={(value) => {
            setDate(value);
            setErrors((current) => ({ ...current, date: undefined }));
          }}
          error={errors.date}
        />
        <View style={s.divider} />
        <DeductionEditor
          items={deductions}
          onChange={(value) => {
            setDeductions(value);
            setErrors((current) => ({ ...current, deductions: undefined }));
          }}
          error={errors.deductions}
        />

        {serverError ? (
          <View style={s.errorWrap}>
            <ErrorState title="Unable to calculate" message={serverError} retryLabel="Retry" onRetry={submit} />
          </View>
        ) : null}

        <Button
          label={offline ? "Offline" : "Calculate PAYE"}
          loading={loading}
          loadingLabel="Calculating…"
          disabled={offline}
          onPress={submit}
        />
      </Card>

      <Text style={s.note}>Budget254 resolves the applicable payroll rules on the server using your calculation date.</Text>
    </Screen>
  );
}

const s = StyleSheet.create({
  top: { flexDirection: "row", alignItems: "center", gap: 9 },
  logo: { width: 38, height: 38, borderRadius: 12, backgroundColor: colors.primary, alignItems: "center", justifyContent: "center" },
  logoText: { color: colors.white, fontSize: 10, fontWeight: "900" },
  brand: { fontSize: 17, fontWeight: "900", color: colors.text },
  title: { fontSize: 32, lineHeight: 39, fontWeight: "900", color: colors.text, marginTop: 24, letterSpacing: -0.5 },
  subtitle: { fontSize: 15, lineHeight: 23, color: colors.textMuted, marginTop: 10 },
  infoBanner: { backgroundColor: colors.accentBg, borderWidth: 1, borderColor: colors.accentBorder, borderRadius: 16, padding: 14, marginTop: 22 },
  infoTitle: { color: colors.accent, fontWeight: "800", fontSize: 13 },
  infoText: { color: colors.textSubtle, fontSize: 12, lineHeight: 18, marginTop: 4 },
  card: { marginTop: 18 },
  cardHeader: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 16 },
  cardTitle: { fontSize: 19, fontWeight: "900", color: colors.text },
  reset: { fontSize: 13, fontWeight: "700", color: colors.accent },
  periodHint: { fontSize: 12, lineHeight: 17, color: colors.textMuted, marginTop: -10, marginBottom: 18 },
  divider: { height: 1, backgroundColor: colors.border, marginVertical: 8 },
  errorWrap: { marginTop: 8, marginBottom: 12 },
  note: { fontSize: 12, lineHeight: 18, color: colors.textMuted, marginTop: 16, textAlign: "center" },
});
