import {
  ApolloClient,
  InMemoryCache,
  Observable,
  type FetchResult,
  type NormalizedCacheObject,
} from "@apollo/client";
import { setContext } from "@apollo/client/link/context";
import { onError } from "@apollo/client/link/error";
import { HttpLink } from "@apollo/client/link/http";
import type { ApolloLink } from "@apollo/client/link/core";
import { relayStylePagination } from "@apollo/client/utilities";

import {
  clearAccessToken,
  getAccessToken,
  setAccessToken,
} from "@/lib/auth-token";
import { readCsrfToken } from "@/lib/csrf";
import { GRAPHQL_URL } from "@/lib/env";

/** Operations that must never trigger a refresh-and-retry loop. */
const AUTH_OPERATION_NAMES = new Set(["Refresh", "Login", "Register"]);

const REFRESH_MUTATION = /* GraphQL */ `
  mutation Refresh {
    refreshToken {
      accessToken
      user {
        id
      }
      userErrors {
        code
      }
    }
  }
`;

interface RefreshResponse {
  data?: {
    refreshToken?: {
      accessToken?: string | null;
      userErrors?: Array<{ code: string }>;
    } | null;
  };
}

/**
 * Exchanges the httpOnly refresh cookie for a new access token.
 *
 * Concurrent failed requests share a single in-flight refresh so the backend is
 * only asked once. Returns the new access token, or null when refresh fails.
 */
let pendingRefresh: Promise<string | null> | null = null;

function refreshAccessToken(): Promise<string | null> {
  if (pendingRefresh) {
    return pendingRefresh;
  }

  pendingRefresh = (async () => {
    try {
      const csrfToken = readCsrfToken();
      const response = await fetch(GRAPHQL_URL, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          ...(csrfToken ? { "X-CSRF-Token": csrfToken } : {}),
        },
        body: JSON.stringify({ query: REFRESH_MUTATION }),
      });

      if (!response.ok) {
        return null;
      }

      const body = (await response.json()) as RefreshResponse;
      const token = body.data?.refreshToken?.accessToken ?? null;

      if (token) {
        setAccessToken(token);
        return token;
      }

      return null;
    } catch {
      return null;
    } finally {
      pendingRefresh = null;
    }
  })();

  return pendingRefresh;
}

const httpLink = new HttpLink({
  uri: GRAPHQL_URL,
  credentials: "include",
});

const authLink = setContext((_operation, { headers }) => {
  const token = getAccessToken();
  const csrfToken = readCsrfToken();
  return {
    headers: {
      ...headers,
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(csrfToken ? { "X-CSRF-Token": csrfToken } : {}),
    },
  };
});

const errorLink = onError(({ graphQLErrors, operation, forward }) => {
  if (!graphQLErrors) {
    return;
  }

  const isUnauthenticated = graphQLErrors.some(
    (error) => error.extensions?.code === "UNAUTHENTICATED",
  );

  if (!isUnauthenticated) {
    return;
  }

  const operationName = operation.operationName;
  if (operationName && AUTH_OPERATION_NAMES.has(operationName)) {
    return;
  }

  // Guard against retrying the same operation more than once.
  const context = operation.getContext();
  if (context.hasRetriedAfterRefresh) {
    clearAccessToken();
    return;
  }

  return new Observable<FetchResult>((observer) => {
    refreshAccessToken()
      .then((token) => {
        if (!token) {
          clearAccessToken();
          observer.error(graphQLErrors[0]);
          return;
        }

        operation.setContext({
          ...context,
          hasRetriedAfterRefresh: true,
          headers: {
            ...context.headers,
            Authorization: `Bearer ${token}`,
          },
        });

        forward(operation).subscribe({
          next: observer.next.bind(observer),
          error: observer.error.bind(observer),
          complete: observer.complete.bind(observer),
        });
      })
      .catch((error) => {
        clearAccessToken();
        observer.error(error);
      });
  });
});

function createCache(): InMemoryCache {
  return new InMemoryCache({
    typePolicies: {
      Query: {
        fields: {
          decisions: relayStylePagination(["teamId"]),
        },
      },
    },
  });
}

export function createApolloClient(): ApolloClient<NormalizedCacheObject> {
  const link: ApolloLink = errorLink.concat(authLink).concat(httpLink);

  return new ApolloClient({
    link,
    cache: createCache(),
    connectToDevTools: import.meta.env.DEV,
  });
}

/** Shared client instance used by the application. */
export const apolloClient = createApolloClient();
