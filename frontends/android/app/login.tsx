import { useState } from "react";
import { StyleSheet, Text, View } from "react-native";
import { router } from "expo-router";
import { Screen } from "../components/Screen";
import { Card } from "../components/Card";
import { Button } from "../components/Button";
import { TextField } from "../components/TextField";
import { ErrorState } from "../components/ErrorState";
import { login } from "../lib/accountApi";
import { APIClientError } from "../lib/httpClient";
import { colors } from "../lib/theme";

export default function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function submit() {
    setError("");
    if (!email.trim() || !password) {
      setError("Enter your email and password.");
      return;
    }
    setLoading(true);
    try {
      await login({ email: email.trim(), password });
      router.replace("/calculator");
    } catch (e) {
      setError(e instanceof APIClientError ? e.message : "Could not log in. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Screen>
      <Text style={s.brand}>Budget254</Text>
      <Text style={s.title}>Welcome back</Text>
      <Text style={s.subtitle}>Log in to save calculations and view your history.</Text>

      <Card style={s.card}>
        <TextField
          label="Email"
          value={email}
          onChangeText={setEmail}
          autoCapitalize="none"
          keyboardType="email-address"
          placeholder="you@example.com"
          returnKeyType="next"
        />
        <TextField
          label="Password"
          value={password}
          onChangeText={setPassword}
          secureTextEntry
          placeholder="Your password"
          returnKeyType="done"
        />
        {error ? <View style={s.errorWrap}><ErrorState title="Login failed" message={error} /></View> : null}
        <Button label="Log in" loading={loading} loadingLabel="Logging in…" onPress={submit} />
      </Card>

      <Text style={s.link} onPress={() => router.push("/register")}>
        Don't have an account? Create one
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
