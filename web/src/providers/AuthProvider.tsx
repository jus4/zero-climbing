import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, type ReactNode } from "react";

import { ApiError, apiFetch } from "../lib/api";
import type { Me } from "../types/api";
import { AuthContext, type AuthContextValue } from "./authContext";

async function fetchMe(): Promise<Me | null> {
  try {
    return await apiFetch<Me>("/auth/me");
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      return null;
    }
    throw error;
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const query = useQuery({ queryKey: ["me"], queryFn: fetchMe });

  const value = useMemo<AuthContextValue>(() => {
    const user = query.data ?? null;

    return {
      user,
      isAuthenticated: Boolean(user),
      onboardingRequired: Boolean(user?.onboardingRequired),
      isLoading: query.isLoading,
      logout: async () => {
        await apiFetch("/auth/logout", { method: "POST" });
        await queryClient.invalidateQueries({ queryKey: ["me"] });
      },
      refetch: async () => {
        await queryClient.invalidateQueries({ queryKey: ["me"] });
      },
    };
  }, [query.data, query.isLoading, queryClient]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
