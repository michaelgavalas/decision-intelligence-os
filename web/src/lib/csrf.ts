/**
 * Reads the `csrf_token` cookie set by the backend. The token must be echoed in
 * the `X-CSRF-Token` header on state-changing requests that rely on the
 * httpOnly refresh cookie (e.g. token refresh and logout).
 */
export function readCsrfToken(): string | null {
  if (typeof document === "undefined") {
    return null;
  }

  const match = document.cookie
    .split("; ")
    .find((row) => row.startsWith("csrf_token="));

  if (!match) {
    return null;
  }

  return decodeURIComponent(match.slice("csrf_token=".length));
}
