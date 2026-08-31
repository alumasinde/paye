// Single source of truth for colors/spacing/radius. Every screen and
// component in the app pulls from here instead of hardcoding hex values,
// so the palette can change in one place instead of N StyleSheet.create
// calls scattered across screens.
export const colors = {
  bg: "#F8FAFC",
  surface: "#FFFFFF",
  border: "#EAECF0",
  borderStrong: "#D0D5DD",
  text: "#101828",
  textMuted: "#667085",
  textSubtle: "#475467",
  primary: "#101828",
  accent: "#175CD3",
  accentBg: "#EFF8FF",
  accentBorder: "#B2DDFF",
  danger: "#B42318",
  dangerBg: "#FEF3F2",
  dangerBorder: "#FECDCA",
  white: "#FFFFFF",
  placeholder: "#98A2B3",
};

export const radius = { sm: 10, md: 14, lg: 16, xl: 20, pill: 999 };
export const spacing = { xs: 4, sm: 8, md: 12, lg: 16, xl: 20, xxl: 24 };
