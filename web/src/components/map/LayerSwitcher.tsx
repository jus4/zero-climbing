export type BaseLayerId = "maastokartta" | "ortokuva";

const LABELS: Record<BaseLayerId, string> = {
  maastokartta: "Maastokartta",
  ortokuva: "Ortokuva",
};

type LayerSwitcherProps = {
  active: BaseLayerId;
  onChange: (id: BaseLayerId) => void;
};

export function LayerSwitcher({ active, onChange }: LayerSwitcherProps) {
  return (
    <div className="absolute right-3 top-3 z-10 flex overflow-hidden rounded-md border border-slate-300 bg-white shadow">
      {(Object.keys(LABELS) as BaseLayerId[]).map((id) => (
        <button
          key={id}
          type="button"
          onClick={() => onChange(id)}
          className={`px-3 py-1.5 text-xs font-semibold transition ${
            active === id
              ? "bg-primary-600 text-white"
              : "bg-white text-slate-700 hover:bg-primary-50 hover:text-primary-700"
          }`}
        >
          {LABELS[id]}
        </button>
      ))}
    </div>
  );
}
