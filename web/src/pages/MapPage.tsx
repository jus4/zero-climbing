import { useMemo } from "react";

import { TerrainMap } from "../components/map/TerrainMap";
import { Spinner } from "../components/ui/Spinner";
import { useDeleteLocation, useLocations, useUpdateLocation } from "../hooks/useLocations";
import { useAuth } from "../providers/authContext";

export function MapPage() {
  const { user } = useAuth();
  const locations = useLocations();
  const updateLocation = useUpdateLocation();
  const deleteLocation = useDeleteLocation();

  const features = useMemo(() => locations.data?.features ?? [], [locations.data]);

  return (
    <div className="relative h-[calc(100dvh-3.5rem)] w-full">
      {locations.isLoading && (
        <div className="absolute inset-0 z-20 flex items-center justify-center bg-white/60">
          <Spinner className="h-8 w-8" />
        </div>
      )}
      <TerrainMap
        locations={features}
        currentUserId={user?.id ?? null}
        onUpdate={(input) => updateLocation.mutateAsync(input)}
        onDelete={(id) => deleteLocation.mutateAsync(id)}
      />
    </div>
  );
}
