import { useContext } from "react";

import { AuthContext, type AuthContextValue } from "./auth-context";

/** Access the authentication context. Must be used within an AuthProvider. */
export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
