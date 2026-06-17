import { useState } from "react";

import { Slider } from "@/components/ui/slider";

interface PercentSliderProps {
  /** Current value, 0-100. */
  value: number;
  onChange(value: number): void;
  id?: string;
  /** Used for accessible labels, e.g. "Confidence". */
  label: string;
  ariaDescribedby?: string;
}

function clampPercent(n: number): number {
  if (!Number.isFinite(n)) {
    return 0;
  }
  return Math.min(100, Math.max(0, Math.round(n)));
}

/**
 * A draggable percentage bar with an inline-editable readout. The number reads
 * as plain text until you hover or focus it; typing a value moves the bar.
 */
export function PercentSlider({
  value,
  onChange,
  id,
  label,
  ariaDescribedby,
}: PercentSliderProps) {
  // Local text buffer so partial edits (empty, "7" before "70") feel natural.
  const [text, setText] = useState(String(value));

  return (
    <div className="flex items-center gap-4 pt-1">
      <Slider
        id={id}
        aria-describedby={ariaDescribedby}
        thumbLabel={`${label} percent`}
        min={0}
        max={100}
        step={1}
        value={[value]}
        onValueChange={([v]) => {
          setText(String(v));
          onChange(v);
        }}
        className="flex-1"
      />
      <div className="flex items-baseline text-base font-semibold tabular-nums text-foreground">
        <input
          type="text"
          inputMode="numeric"
          aria-label={`${label} percent`}
          value={text}
          onChange={(event) => {
            const digits = event.target.value.replace(/[^0-9]/g, "").slice(0, 3);
            setText(digits);
            onChange(digits === "" ? 0 : clampPercent(Number(digits)));
          }}
          onFocus={(event) => event.currentTarget.select()}
          onBlur={() => {
            const next = clampPercent(Number(text || 0));
            setText(String(next));
            onChange(next);
          }}
          className="w-8 rounded bg-transparent text-right tabular-nums text-foreground outline-none transition-colors duration-150 hover:bg-muted focus:bg-muted focus:ring-1 focus:ring-ring"
        />
        <span aria-hidden="true">%</span>
      </div>
    </div>
  );
}
