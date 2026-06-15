import { Navigate, useLocation } from "react-router-dom";

import { Spinner } from "@/components/ui/spinner";
import { useAuth } from "@/features/auth/use-auth";

interface ProtectedRouteProps {
  children: React.ReactNode;
}

/**
 * Gate for authenticated routes.
 *
 * While the initial session refresh resolves it shows a loading state.
 * Unauthenticated users are redirected to `/login`, preserving the intended
 * destination so they can be returned to it after signing in.
 */
export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { status } = useAuth();
  const location = useLocation();

  if (status === "loading") {
    return (
      <div className="flex min-h-dvh items-center justify-center bg-background">
        <Spinner className="size-8" />
      </div>
    );
  }

  if (status === "unauthenticated") {
    return (
      <Navigate
        to="/login"
        replace
        state={{ from: location.pathname + location.search }}
      />
    );
  }

  return <>{children}</>;
}
