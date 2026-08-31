import { Component, ReactNode } from "react";
import { StyleSheet, Text, View } from "react-native";
import { Button } from "./Button";
import { colors } from "../lib/theme";

type Props = { children: ReactNode };
type State = { error: Error | null };

// Catches any render-time JavaScript error anywhere below it and shows a
// recovery screen instead of a blank white screen (Expo/React Native's
// default when an uncaught render error occurs). This does NOT catch
// errors in event handlers or async code - those are already handled
// locally (e.g. calculatePAYE's try/catch in calculator.tsx) - it's
// specifically the last line of defence against an unexpected render
// crash taking down the whole app.
export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: { componentStack: string }) {
    // eslint-disable-next-line no-console
    console.error("Unhandled render error", error, info.componentStack);
  }

  reset = () => this.setState({ error: null });

  render() {
    if (this.state.error) {
      return (
        <View style={s.wrap}>
          <Text style={s.title}>Something went wrong</Text>
          <Text style={s.message}>Budget254 ran into an unexpected error. Your saved data is safe - try again.</Text>
          <Button label="Try again" onPress={this.reset} />
        </View>
      );
    }
    return this.props.children;
  }
}

const s = StyleSheet.create({
  wrap: { flex: 1, backgroundColor: colors.bg, justifyContent: "center", padding: 24, gap: 14 },
  title: { fontSize: 24, fontWeight: "900", color: colors.text },
  message: { color: colors.textMuted, lineHeight: 21, marginBottom: 6 },
});
