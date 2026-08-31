import { ReactNode } from "react";
import { KeyboardAvoidingView, Platform, ScrollView, ScrollViewProps, StyleSheet, StyleProp, ViewStyle } from "react-native";
import { colors } from "../lib/theme";

type Props = Omit<ScrollViewProps, "contentContainerStyle"> & {
  children: ReactNode;
  contentContainerStyle?: StyleProp<ViewStyle>;
};

// Standard page scaffold (keyboard-avoiding scroll view with the app's
// background + padding) used by every screen. Replaces the copy-pasted
// KeyboardAvoidingView/ScrollView block that used to open each screen file.
export function Screen({ children, contentContainerStyle, ...rest }: Props) {
  return (
    <KeyboardAvoidingView style={s.flex} behavior={Platform.OS === "ios" ? "padding" : undefined}>
      <ScrollView
        contentContainerStyle={[s.page, contentContainerStyle]}
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
        {...rest}
      >
        {children}
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const s = StyleSheet.create({
  flex: { flex: 1, backgroundColor: colors.bg },
  page: { paddingHorizontal: 20, paddingTop: 60, paddingBottom: 44 },
});
