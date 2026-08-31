import { useState } from "react";
import { StyleSheet, Text, View } from "react-native";
import { router } from "expo-router";
import { Screen } from "../components/Screen";
import { Card } from "../components/Card";
import { Button } from "../components/Button";
import { TextField } from "../components/TextField";
import { ErrorState } from "../components/ErrorState";
import { register } from "../lib/accountApi";
import { APIClientError } from "../lib/httpClient";
import { colors } from "../lib/theme";

type FormErrors = { firstName?: string; lastName?: string; email?: string; password?: string };

function validate(firstName: string, lastName: string, email: string, password: string): FormErrors {
  const errors: FormErrors = {};
  if (!firstName.trim()) errors.firstName = "Enter your first name.";
  if (!lastName.trim()) errors.lastName = "Enter your last name.";
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) errors.email = "Enter a valid email address.";
  if (password.length < 12) errors.password = "Use at least 12 characters.";
  return errors;
}

export default function Register() {
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<FormErrors>({});
  const [loading, setLoading] = useState(false);
  const [serverError, setServerError] = useState("");

  async function submit() {
    setServerError("");
    const nextErrors = validate(firstName, lastName, email, password);
    setErrors(nextErrors);
    if (Object.keys(nextErrors).length > 0) return;

    setLoading(true);
    try {
      await register({ first_name: firstName.trim(), last_name: lastName.trim(), email: email.trim(), password });
      router.replace("/calculator");
    } catch (e) {
      setServerError(e instanceof APIClientError ? e.message : "Could not create your account. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Screen>
      <Text style={s.brand}>Budget254</Text>
      <Text style={s.title}>Create account</Text>
      <Text style={s.subtitle}>Save calculations and view your history across devices.</Text>

      <Card style={s.card}>
        <TextField label="First name" value={firstName} onChangeText={setFirstName} error={errors.firstName} returnKeyType="next" />
        <TextField label="Last name" value={lastName} onChangeText={setLastName} error={errors.lastName} returnKeyType="next" />
        <TextField
          label="Email"
          value={email}
          onChangeText={setEmail}
          autoCapitalize="none"
          keyboardType="email-address"
          error={errors.email}
          returnKeyType="next"
        />
        <TextField
          label="Password"
          value={password}
          onChangeText={setPassword}
          secureTextEntry
          error={errors.password}
          hint={errors.password ? undefined : "At least 12 characters."}
          returnKeyType="done"
        />
        {serverError ? <View style={s.errorWrap}><ErrorState title="Registration failed" message={serverError} /></View> : null}
        <Button label="Create account" loading={loading} loadingLabel="Creating account…" onPress={submit} />
      </Card>

      <Text style={s.link} onPress={() => router.back()}>
        Already have an account? Log in
      </Text>
    </Screen>
  );
}

const s = StyleSheet.create({
  brand: { fontSize: 16, fontWeight: "900", color: colors.text },
  title: { fontSize: 30, fontWeight: "900", color: colors.text, marginTop: 8 },
  subtitle: { fontSize: 14, lineHeight: 21, color: colors.textMuted, marginTop: 8 },
  card: { marginTop: 22 },
  errorWrap: { marginBottom: 16 },
  link: { textAlign: "center", marginTop: 22, fontWeight: "700", color: colors.accent },
});
