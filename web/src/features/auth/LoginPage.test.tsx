import { ApolloClient, InMemoryCache } from "@apollo/client";
import { MockLink, type MockedResponse } from "@apollo/client/testing";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it } from "vitest";
import { MemoryRouter, Route, Routes } from "react-router-dom";

import { AuthProvider } from "@/features/auth/auth-context";
import { LoginPage } from "@/features/auth/LoginPage";
import { LoginDocument, RefreshDocument } from "@/graphql/generated/graphql";
import { clearAccessToken, getAccessToken } from "@/lib/auth-token";

/** Refresh mock that reports no active session (logged-out start state). */
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

function renderLogin(mocks: MockedResponse[]) {
  const client = new ApolloClient({
    link: new MockLink([noSessionRefresh, ...mocks]),
    cache: new InMemoryCache(),
  });

  return render(
    <AuthProvider client={client}>
      <MemoryRouter initialEntries={["/login"]}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/decisions" element={<div>Decisions home</div>} />
        </Routes>
      </MemoryRouter>
    </AuthProvider>,
  );
}

describe("LoginPage", () => {
  afterEach(() => {
    clearAccessToken();
  });

  it("signs in and navigates on success", async () => {
    const loginMock: MockedResponse = {
      request: {
        query: LoginDocument,
        variables: {
          input: { email: "ada@example.com", password: "supersecret" },
        },
      },
      result: {
        data: {
          login: {
            __typename: "AuthPayload",
            user: {
              __typename: "User",
              id: "u1",
              email: "ada@example.com",
              name: "Ada",
            },
            accessToken: "access-token-123",
            accessExpiresAt: "2026-06-14T00:00:00Z",
            userErrors: [],
          },
        },
      },
    };

    const user = userEvent.setup();
    renderLogin([loginMock]);

    await user.type(screen.getByLabelText(/email/i), "ada@example.com");
    await user.type(
      screen.getByLabelText(/password/i, { selector: "input" }),
      "supersecret",
    );
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    await waitFor(() => {
      expect(screen.getByText("Decisions home")).toBeInTheDocument();
    });
    expect(getAccessToken()).toBe("access-token-123");
  });

  it("shows an error and does not navigate on invalid credentials", async () => {
    const loginMock: MockedResponse = {
      request: {
        query: LoginDocument,
        variables: {
          input: { email: "ada@example.com", password: "wrongpass" },
        },
      },
      result: {
        data: {
          login: {
            __typename: "AuthPayload",
            user: null,
            accessToken: null,
            accessExpiresAt: null,
            userErrors: [
              {
                __typename: "UserError",
                field: null,
                message: "bad credentials",
                code: "INVALID_CREDENTIALS",
              },
            ],
          },
        },
      },
    };

    const user = userEvent.setup();
    renderLogin([loginMock]);

    await user.type(screen.getByLabelText(/email/i), "ada@example.com");
    await user.type(
      screen.getByLabelText(/password/i, { selector: "input" }),
      "wrongpass",
    );
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/incorrect email or password/i),
      ).toBeInTheDocument();
    });
    expect(screen.queryByText("Decisions home")).not.toBeInTheDocument();
    expect(getAccessToken()).toBeNull();
  });
});
