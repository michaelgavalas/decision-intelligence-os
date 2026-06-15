/** Helpers for moving between ISO timestamps and `<input type="date">` values. */

/**
 * Formats an ISO timestamp as a `YYYY-MM-DD` value for a date input. Returns an
 * empty string when the value is missing or unparseable.
 */
export function dateInputValue(value: string | null | undefined): string {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return date.toISOString().slice(0, 10);
}

/**
 * Converts a `YYYY-MM-DD` date input value into an ISO timestamp at UTC
 * midnight, or null when empty.
 */
export function dateInputToISO(value: string): string | null {
  const trimmed = value.trim();
  if (!trimmed) {
    return null;
  }
  const date = new Date(`${trimmed}T00:00:00.000Z`);
  if (Number.isNaN(date.getTime())) {
    return null;
  }
  return date.toISOString();
}
