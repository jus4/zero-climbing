import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { apiFetch } from "../lib/api";
import type { LocationCollection } from "../types/api";

export function useLocations() {
  return useQuery({
    queryKey: ["locations"],
    queryFn: () => apiFetch<LocationCollection>("/locations"),
  });
}

type UpdateLocationInput = {
  id: number;
  title: string;
  description: string;
};

export function useUpdateLocation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, title, description }: UpdateLocationInput) =>
      apiFetch(`/locations/${id}`, {
        method: "PATCH",
        body: { title, description },
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
