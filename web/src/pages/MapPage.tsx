import { useMemo, useRef, useState } from "react";

import { TerrainMap, type TerrainMapHandle } from "../components/map/TerrainMap";
import { VisibleLocationsList } from "../components/map/VisibleLocationsList";
import { Spinner } from "../components/ui/Spinner";
import {
  useDeleteLocation,
  useLocations,
  usePrivateLocations,
  useUpdateLocation,
} from "../hooks/useLocations";
import { useAuth } from "../providers/authContext";
import type { LocationFeature } from "../types/api";

export function MapPage() {
  const { user } = useAuth();
  const locations = useLocations();
  const privateLocations = usePrivateLocations(user !== null);
  const updateLocation = useUpdateLocation();
  const deleteLocation = useDeleteLocation();

  const features = useMemo(() => {
    const publicFeatures = locations.data?.features ?? [];
    const privateFeatures = user ? (privateLocations.data?.features ?? []) : [];

    const byId = new Map<number, LocationFeature>();
    for (const feature of publicFeatures) {
      byId.set(feature.id, feature);
    }
    for (const feature of privateFeatures) {
      byId.set(feature.id, feature);
    }

    return [...byId.values()];
  }, [locations.data, privateLocations.data, user]);

  const mapRef = useRef<TerrainMapHandle | null>(null);
  const [visibleIds, setVisibleIds] = useState<Set<number>>(new Set());

  const visibleLocations = useMemo(
    () => features.filter((feature) => visibleIds.has(feature.id)),
    [features, visibleIds],
  );

  return (
    <div className="flex h-[calc(100dvh-3.5rem)] w-full flex-col md:flex-row">
      <div className="relative min-h-0 flex-1">
        {locations.isLoading && (
          <div className="absolute inset-0 z-20 flex items-center justify-center bg-white/60">
            <Spinner className="h-8 w-8" />
          </div>
        )}
        <TerrainMap
          ref={mapRef}
          locations={features}
          currentUserId={user?.id ?? null}
          onUpdate={(input) => updateLocation.mutateAsync(input)}
          onDelete={(id) => deleteLocation.mutateAsync(id)}
          onVisibleIdsChange={(ids) => setVisibleIds(new Set(ids))}
        />
      </div>

      <aside className="h-64 shrink-0 overflow-y-auto border-t border-slate-200 bg-white md:h-auto md:w-80 md:border-l md:border-t-0">
        <VisibleLocationsList
          locations={visibleLocations}
          onSelect={(id) => mapRef.current?.focusLocation(id)}
        />
      </aside>
    </div>
  );
}
