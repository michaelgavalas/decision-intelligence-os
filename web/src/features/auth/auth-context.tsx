import {
  type ApolloClient,
  type NormalizedCacheObject,
} from "@apollo/client";
import { createContext, useCallback, useEffect, useMemo, useState } from "react";

import {
  LoginDocument,
  LogoutDocument,
  RefreshDocument,
  RegisterDocument,
  type AuthPayloadFieldsFragment,
  type LoginMutation,
  type LoginMutationVariables,
  type LogoutMutation,
  type LogoutMutationVariables,
  type RefreshMutation,
  type RefreshMutationVariables,
  type RegisterMutation,
  type RegisterMutationVariables,
} from "@/graphql/generated/graphql";
import { apolloClient } from "@/lib/apollo";
import {
  clearAccessToken,
  setAccessToken,
} from "@/lib/auth-token";

export type AuthStatus = "loading" | "authenticated" | "unauthenticated";

export interface AuthUser {
  id: string;
  email: string;
  name: string;
}

export interface UserError {
  field?: string | null;
  message: string;
  code: string;
}

export interface AuthResult {
  ok: boolean;
  userErrors: UserError[];
}

export interface AuthContextValue {
  user: AuthUser | null;
  status: AuthStatus;
  login(email: string, password: string): Promise<AuthResult>;
  register(input: {
    email: string;
    name: string;
    password: string;
  }): Promise<AuthResult>;
  logout(): Promise<void>;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

/** A generic error surfaced when a transport/network failure occurs. */
const NETWORK_ERROR: UserError = {
  message: "Something went wrong. Please try again.",
  code: "NETWORK_ERROR",
};

/**
 * Applies a successful auth payload to in-memory state: stores the access token
 * and returns the authenticated user. Returns null when the payload carries no
 * token (e.g. validation errors or a missing session).
 */
function applyPayload(payload: AuthPayloadFieldsFragment): AuthUser | null {
  if (!payload.accessToken || !payload.user) {
    return null;
  }

  setAccessToken(payload.accessToken);
  return {
    id: payload.user.id,
    email: payload.user.email,
    name: payload.user.name,
  };
}

interface AuthProviderProps {
  children: React.ReactNode;
  /**
   * The Apollo client used for auth operations. Defaults to the shared
   * application client; tests may inject a client backed by a mock link.
   */
  client?: ApolloClient<NormalizedCacheObject>;
}

export function AuthProvider({ children, client }: AuthProviderProps) {
  const apollo = client ?? apolloClient;
  const [user, setUser] = useState<AuthUser | null>(null);
  const [status, setStatus] = useState<AuthStatus>("loading");

  // Attempt a silent session refresh on mount using the httpOnly cookie. A
  // missing session is a normal state, so failures here are not surfaced.
  useEffect(() => {
    let cancelled = false;

    async function restoreSession() {
      try {
        const { data } = await apollo.mutate<
          RefreshMutation,
          RefreshMutationVariables
        >({ mutation: RefreshDocument });

        if (cancelled) {
          return;
        }

        const payload = data?.refreshToken;
        const restored = payload ? applyPayload(payload) : null;

        if (restored) {
          setUser(restored);
          setStatus("authenticated");
        } else {
          setStatus("unauthenticated");
        }
      } catch {
        if (!cancelled) {
          clearAccessToken();
          setStatus("unauthenticated");
        }
      }
    }

    void restoreSession();

    return () => {
      cancelled = true;
    };
  }, [apollo]);

  const login = useCallback<AuthContextValue["login"]>(
    async (email, password) => {
      try {
        const { data } = await apollo.mutate<
          LoginMutation,
          LoginMutationVariables
        >({
          mutation: LoginDocument,
          variables: { input: { email, password } },
        });

        const payload = data?.login;
        if (!payload) {
          return { ok: false, userErrors: [NETWORK_ERROR] };
        }

        if (payload.userErrors.length > 0) {
          return { ok: false, userErrors: payload.userErrors };
        }

        const authenticated = applyPayload(payload);
        if (!authenticated) {
          return { ok: false, userErrors: [NETWORK_ERROR] };
        }

        setUser(authenticated);
        setStatus("authenticated");
        return { ok: true, userErrors: [] };
      } catch {
        return { ok: false, userErrors: [NETWORK_ERROR] };
      }
    },
    [apollo],
  );

  const register = useCallback<AuthContextValue["register"]>(async (input) => {
    try {
      const { data } = await apollo.mutate<
        RegisterMutation,
        RegisterMutationVariables
      >({
        mutation: RegisterDocument,
        variables: { input },
      });

      const payload = data?.register;
      if (!payload) {
        return { ok: false, userErrors: [NETWORK_ERROR] };
      }

      if (payload.userErrors.length > 0) {
        return { ok: false, userErrors: payload.userErrors };
      }

      const authenticated = applyPayload(payload);
      if (!authenticated) {
        return { ok: false, userErrors: [NETWORK_ERROR] };
      }

      setUser(authenticated);
      setStatus("authenticated");
      return { ok: true, userErrors: [] };
    } catch {
      return { ok: false, userErrors: [NETWORK_ERROR] };
    }
  }, [apollo]);

  const logout = useCallback<AuthContextValue["logout"]>(async () => {
    try {
      await apollo.mutate<LogoutMutation, LogoutMutationVariables>({
        mutation: LogoutDocument,
      });
    } catch {
      // Logout is best-effort: clear local state regardless of the result.
    }

    clearAccessToken();
    setUser(null);
    setStatus("unauthenticated");
    await apollo.clearStore();
  }, [apollo]);

  const value = useMemo<AuthContextValue>(
    () => ({ user, status, login, register, logout }),
    [user, status, login, register, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
