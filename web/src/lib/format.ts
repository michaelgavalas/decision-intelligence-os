/** A placeholder rendered when a value is absent. */
export const EMPTY_PLACEHOLDER = "-";

const dateFormatter = new Intl.DateTimeFormat("en-US", {
  year: "numeric",
  month: "short",
  day: "numeric",
});

/**
 * Formats an ISO timestamp as a short, human-readable date (e.g. "Jun 14, 2026").
 * Returns an em dash when the value is missing or unparseable.
 */
export function formatDate(value: string | null | undefined): string {
  if (!value) {
    return EMPTY_PLACEHOLDER;
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return EMPTY_PLACEHOLDER;
  }

  return dateFormatter.format(date);
}
