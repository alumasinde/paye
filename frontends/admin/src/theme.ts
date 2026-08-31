// Applies colors from /config.txt (served as a static file from
// public/config.txt) onto the CSS custom properties defined in
// src/styles/theme.css. Editing that file and refreshing the browser is
// enough to reskin the app - no rebuild needed. Anything config.txt
// doesn't specify (or a value it can't parse) just keeps its default
// from theme.css.

const CONFIG_KEY_TO_CSS_VAR: Record<string, string> = {
  primary: "--color-primary",
  primaryText: "--color-primary-text",
  accent: "--color-accent",
  background: "--color-background",
  surface: "--color-surface",
  text: "--color-text",
  mutedText: "--color-muted-text",
  border: "--color-border",
  danger: "--color-danger",
  success: "--color-success",
};

export async function loadTheme(): Promise<void> {
  try {
    const res = await fetch("/config.txt", { cache: "no-store" });
    if (!res.ok) return;
    const text = await res.text();
    for (const rawLine of text.split("\n")) {
      const line = rawLine.trim();
      if (!line || line.startsWith("#")) continue;
      const eq = line.indexOf("=");
      if (eq === -1) continue;
      const key = line.slice(0, eq).trim();
      const value = line.slice(eq + 1).trim();
      const cssVar = CONFIG_KEY_TO_CSS_VAR[key];
      if (cssVar && value) {
        document.documentElement.style.setProperty(cssVar, value);
      }
    }
  } catch {
    // config.txt is optional network access at startup - if it's missing
    // or fetch fails for any reason, the app just keeps theme.css's
    // built-in defaults.
  }
}
