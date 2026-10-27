import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiFetch } from "../lib/api";
import type { LocationCollection } from "../types/api";

export function useLocations() {
  return useQuery({
    queryKey: ["locations", "public"],
    queryFn: () => apiFetch<LocationCollection>("/locations"),
  });
}

export function usePrivateLocations(enabled: boolean) {
  return useQuery({
    queryKey: ["locations", "private"],
    queryFn: () => apiFetch<LocationCollection>("/locations/private"),
    enabled,
  });
}

type UpdateLocationInput = {
  id: number;
  title: string;
  description: string;
  isPublic: boolean;
};

export function useUpdateLocation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, title, description, isPublic }: UpdateLocationInput) =>
      apiFetch(`/locations/${id}`, {
        method: "PATCH",
        body: { title, description, is_public: isPublic },
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["locations"] }),
  });
}

export function useDeleteLocation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => apiFetch(`/locations/${id}`, { method: "DELETE" }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["locations"] }),
  });
}
