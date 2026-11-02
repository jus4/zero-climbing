import type { LocationFeature } from "../../types/api";

type VisibleLocationsListProps = {
  locations: LocationFeature[];
  onSelect: (id: number) => void;
};

export function VisibleLocationsList({ locations, onSelect }: VisibleLocationsListProps) {
  return (
    <div className="flex h-full flex-col">
      <div className="border-b border-slate-200 px-4 py-3">
        <h2 className="text-sm font-semibold text-slate-900">Näkyvät lokaatiot</h2>
        <p className="mt-0.5 text-xs text-slate-500">
          {locations.length} {locations.length === 1 ? "lokaatio" : "lokaatiota"} näkyvissä kartalla.
        </p>
      </div>

      {locations.length === 0 ? (
        <p className="px-4 py-6 text-sm text-slate-500">
          No locations in this view. Pan or zoom out to find some.
        </p>
      ) : (
        <ul className="flex-1 divide-y divide-slate-100 overflow-y-auto">
          {locations.map((location) => {
            const { id, properties } = location;

            return (
              <li key={id}>
                <button
                  type="button"
                  onClick={() => onSelect(id)}
                  className="flex w-full items-start gap-3 px-4 py-3 text-left transition hover:bg-primary-50"
                >
                  <span
                    aria-hidden="true"
                    className={`mt-1.5 h-2.5 w-2.5 shrink-0 rounded-full ${properties.is_public ? "bg-green-500" : "bg-red-500"
                      }`}
                  />
                  <span className="min-w-0">
                    <span className="block truncate text-sm font-semibold text-slate-900">
                      {properties.title || "Untitled"}
                    </span>
                    <span className="mt-0.5 block truncate text-xs text-slate-500">
                      {properties.description || "No comments."}
                    </span>
                    {properties.user && (
                      <span className="mt-1 block text-[11px] text-slate-400">
                        Added by {properties.user}
                      </span>
                    )}
                  </span>
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
