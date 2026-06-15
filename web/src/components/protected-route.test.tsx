import { ApolloClient, InMemoryCache } from "@apollo/client";
import { MockLink, type MockedResponse } from "@apollo/client/testing";
import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { MemoryRouter, Route, Routes } from "react-router-dom";

import { ProtectedRoute } from "@/components/protected-route";
import { AuthProvider } from "@/features/auth/auth-context";
import { RefreshDocument } from "@/graphql/generated/graphql";
import { clearAccessToken } from "@/lib/auth-token";

const noSessionRefresh: MockedResponse = {
  request: { query: RefreshDocument },
  result: {
    data: {
      refreshToken: {
        __typename: "AuthPayload",
        user: null,
        accessToken: null,
        accessExpiresAt: null,
        userErrors: [],
      },
    },
  },
};

describe("ProtectedRoute", () => {
  afterEach(() => {
    clearAccessToken();
  });

  it("redirects to /login when unauthenticated", async () => {
    const client = new ApolloClient({
      link: new MockLink([noSessionRefresh]),
      cache: new InMemoryCache(),
    });

    render(
      <AuthProvider client={client}>
        <MemoryRouter initialEntries={["/decisions"]}>
          <Routes>
            <Route
              path="/decisions"
              element={
                <ProtectedRoute>
                  <div>Protected content</div>
                </ProtectedRoute>
              }
            />
            <Route path="/login" element={<div>Login screen</div>} />
          </Routes>
        </MemoryRouter>
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.getByText("Login screen")).toBeInTheDocument();
    });
    expect(screen.queryByText("Protected content")).not.toBeInTheDocument();
  });
});
