import { lazy } from "react";
import { createBrowserRouter, Navigate, Outlet } from "react-router-dom";

import { ErrorFallback } from "@/components/error-boundary";
import { ProtectedRoute } from "@/components/protected-route";
import { AuthProvider } from "@/features/auth/auth-context";
import { CurrentTeamProvider } from "@/features/teams/current-team";
import { AppShell } from "@/layouts/AppShell";
import { AuthLayout } from "@/layouts/AuthLayout";
import { LoginPage } from "@/pages/login-page";
import { NotFoundPage } from "@/pages/not-found-page";
import { RegisterPage } from "@/pages/register-page";

/**
 * Heavy authenticated pages are split into their own chunks so the initial
 * load (login/register) stays small. The charts-heavy analytics page in
 * particular keeps recharts out of the main bundle.
 */
const DecisionsPage = lazy(() =>
  import("@/pages/decisions-page").then((m) => ({ default: m.DecisionsPage })),
);
const DecisionDetailPage = lazy(() =>
  import("@/pages/decision-detail-page").then((m) => ({
    default: m.DecisionDetailPage,
  })),
);
const AnalyticsPage = lazy(() =>
  import("@/pages/analytics-page").then((m) => ({ default: m.AnalyticsPage })),
);
const TeamsPage = lazy(() =>
  import("@/pages/teams-page").then((m) => ({ default: m.TeamsPage })),
);

/**
 * A pathless root route that makes the auth context available to every route,
 * including the public login and register pages.
 */
function AuthRoot() {
  return (
    <AuthProvider>
      <Outlet />
    </AuthProvider>
  );
}

export const router = createBrowserRouter([
  {
    element: <AuthRoot />,
    errorElement: <ErrorFallback />,
    children: [
      {
        element: <AuthLayout />,
        children: [
          { path: "/login", element: <LoginPage /> },
          { path: "/register", element: <RegisterPage /> },
        ],
      },
      {
        element: (
          <ProtectedRoute>
            <CurrentTeamProvider>
              <AppShell />
            </CurrentTeamProvider>
          </ProtectedRoute>
        ),
        children: [
          { index: true, element: <Navigate to="/decisions" replace /> },
          { path: "/decisions", element: <DecisionsPage /> },
          { path: "/decisions/:id", element: <DecisionDetailPage /> },
          { path: "/analytics", element: <AnalyticsPage /> },
          { path: "/teams", element: <TeamsPage /> },
        ],
      },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
]);
