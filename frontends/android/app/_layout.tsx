import { Stack } from "expo-router";
import { StatusBar } from "expo-status-bar";
import { ErrorBoundary } from "../components/ErrorBoundary";

export default function Layout() {
  return (
    <ErrorBoundary>
      <StatusBar style="dark" />
      <Stack screenOptions={{ headerShown: false }} />
    </ErrorBoundary>
  );
}
