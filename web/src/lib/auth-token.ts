/**
 * In-memory access-token store.
 *
 * The access token is deliberately kept in memory (not localStorage) so that it
 * is not readable by injected scripts. The httpOnly refresh cookie is what
 * survives a page reload and is exchanged for a fresh access token on startup.
 */

type Listener = (token: string | null) => void;

let accessToken: string | null = null;
const listeners = new Set<Listener>();

export function getAccessToken(): string | null {
  return accessToken;
}

export function setAccessToken(token: string | null): void {
  accessToken = token;
  for (const listener of listeners) {
    listener(accessToken);
  }
}

export function clearAccessToken(): void {
  setAccessToken(null);
}

/** Subscribe to token changes. Returns an unsubscribe function. */
export function subscribeAccessToken(listener: Listener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}
