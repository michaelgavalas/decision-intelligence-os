/**
 * Resolves a design-token CSS variable to its current computed value so that
 * Recharts (which needs concrete color strings) stays in sync with the theme.
 * Falls back to a neutral value during SSR / tests where `getComputedStyle`
 * is unavailable.
 */
export function resolveColor(token: string, fallback = "#64748b"): string {
  if (typeof window === "undefined" || typeof document === "undefined") {
    return fallback;
  }
  const value = getComputedStyle(document.documentElement)
    .getPropertyValue(token)
    .trim();
  return value || fallback;
}

export const CHART_TOKENS = {
  primary: "--primary",
  secondary: "--secondary",
  accent: "--accent",
  success: "--success",
  border: "--border",
  mutedForeground: "--muted-foreground",
  surface: "--surface",
} as const;
