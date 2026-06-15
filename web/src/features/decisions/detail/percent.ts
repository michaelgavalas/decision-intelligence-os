/**
 * Helpers for converting between a stored fraction in [0, 1] and the whole
 * percentage (0-100) shown to people in the UI.
 */

/** Formats a fraction in [0, 1] as a whole-number percent string, e.g. "70%". */
export function toPercent(fraction: number): string {
  const clamped = Math.min(1, Math.max(0, fraction));
  return `${Math.round(clamped * 100)}%`;
}

/** The whole-number percent value (0-100) for a fraction in [0, 1]. */
export function toPercentValue(fraction: number): number {
  const clamped = Math.min(1, Math.max(0, fraction));
  return Math.round(clamped * 100);
}

export interface ParsedPercent {
  /** The fraction in [0, 1] when valid, otherwise null. */
  fraction: number | null;
  /** A validation message when the input is not a valid percent. */
  error: string | null;
}

/**
 * Parses a percent text input (0-100) into a fraction in [0, 1]. Returns an
 * error message when the value is empty, non-numeric, or out of range.
 */
export function fromPercent(input: string): ParsedPercent {
  const trimmed = input.trim();
  if (trimmed === "") {
    return { fraction: null, error: "Enter a value between 0 and 100." };
  }

  const value = Number(trimmed);
  if (!Number.isFinite(value)) {
    return { fraction: null, error: "Enter a number between 0 and 100." };
  }

  if (value < 0 || value > 100) {
    return { fraction: null, error: "Must be between 0 and 100." };
  }

  return { fraction: value / 100, error: null };
}
