import { Monitor, Moon, Sun } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useTheme, type Theme } from "@/hooks/useTheme";

const ORDER: Theme[] = ["system", "light", "dark"];
const ICONS = { system: Monitor, light: Sun, dark: Moon } as const;
const LABELS = { system: "System", light: "Light", dark: "Dark" } as const;

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const next = ORDER[(ORDER.indexOf(theme) + 1) % ORDER.length];
  const Icon = ICONS[theme];

  return (
    <Button
      variant="ghost"
      size="iconSm"
      onClick={() => setTheme(next)}
      title={`Theme: ${LABELS[theme]}`}
      aria-label={`Theme: ${LABELS[theme]}. Switch to ${LABELS[next]}.`}
    >
      <Icon className="size-4" />
    </Button>
  );
}
