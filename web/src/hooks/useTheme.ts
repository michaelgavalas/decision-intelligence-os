import { useCallback, useEffect, useState } from "react";

export type Theme = "light" | "dark" | "system";
export type ResolvedTheme = "light" | "dark";

const STORAGE_KEY = "dios-theme";

function getStoredTheme(): Theme {
  if (typeof window === "undefined") {
    return "system";
  }
  const stored = window.localStorage.getItem(STORAGE_KEY);
  if (stored === "light" || stored === "dark" || stored === "system") {
    return stored;
  }
  return "system";
}

function prefersDark(): boolean {
  return (
    typeof window !== "undefined" &&
    window.matchMedia("(prefers-color-scheme: dark)").matches
  );
}

function resolve(theme: Theme): ResolvedTheme {
  if (theme === "system") {
    return prefersDark() ? "dark" : "light";
  }
  return theme;
}

function applyTheme(resolved: ResolvedTheme): void {
  if (typeof document === "undefined") {
    return;
  }
  document.documentElement.classList.toggle("dark", resolved === "dark");
}

/**
 * Theme hook. Persists the user's preference to localStorage, defaults to the
 * OS preference, and toggles the `.dark` class on `<html>`.
 */
export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(getStoredTheme);
  const [resolvedTheme, setResolvedTheme] = useState<ResolvedTheme>(() =>
    resolve(getStoredTheme()),
  );

  useEffect(() => {
    const resolved = resolve(theme);
    setResolvedTheme(resolved);
    applyTheme(resolved);
  }, [theme]);

  useEffect(() => {
    if (theme !== "system") {
      return;
    }
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = () => {
      const resolved = resolve("system");
      setResolvedTheme(resolved);
      applyTheme(resolved);
    };
    media.addEventListener("change", handler);
    return () => media.removeEventListener("change", handler);
  }, [theme]);

  const setTheme = useCallback((next: Theme) => {
    window.localStorage.setItem(STORAGE_KEY, next);
    setThemeState(next);
  }, []);

  const toggleTheme = useCallback(() => {
    setTheme(resolve(getStoredTheme()) === "dark" ? "light" : "dark");
  }, [setTheme]);

  return { theme, resolvedTheme, setTheme, toggleTheme };
}
