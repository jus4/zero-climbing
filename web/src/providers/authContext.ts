import { createContext, useContext } from "react";

import type { Me } from "../types/api";

export type AuthContextValue = {
  user: Me | null;
  isAuthenticated: boolean;
  onboardingRequired: boolean;
  isLoading: boolean;
  logout: () => Promise<void>;
  refetch: () => Promise<void>;
};

export const AuthContext = createContext<AuthContextValue | null>(null);

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
