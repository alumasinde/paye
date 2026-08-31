import { StyleSheet, Text, View } from "react-native";
import { router } from "expo-router";
import { useSession } from "../lib/useSession";
import { colors } from "../lib/theme";

// The entry point into everything account-related (login, history,
// logout). Previously login.tsx and history.tsx existed but nothing in
// the app ever linked to them - this is what makes them reachable.
export function AccountBar() {
  const { session, loading, signOut } = useSession();

  if (loading) return <View style={s.row} />;

  if (!session) {
    return (
      <View style={s.row}>
        <Text style={s.link} onPress={() => router.push("/login")}>
          Log in
        </Text>
        <Text style={s.link} onPress={() => router.push("/register")}>
          Create account
        </Text>
      </View>
    );
  }

  async function handleSignOut() {
    await signOut();
    router.replace("/calculator");
  }

  return (
    <View style={s.row}>
      <Text style={s.greeting} numberOfLines={1}>
        Hi, {session.user.first_name}
      </Text>
      <View style={s.actions}>
        <Text style={s.link} onPress={() => router.push("/history")}>
          History
        </Text>
        <Text style={s.link} onPress={handleSignOut}>
          Log out
        </Text>
      </View>
    </View>
  );
}

const s = StyleSheet.create({
  row: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 6, minHeight: 20 },
  greeting: { fontSize: 13, fontWeight: "700", color: colors.text, flexShrink: 1, marginRight: 12 },
  actions: { flexDirection: "row", gap: 16 },
  link: { fontSize: 13, fontWeight: "700", color: colors.accent },
});
