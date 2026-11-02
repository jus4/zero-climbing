import GeoJSON from "ol/format/GeoJSON";
import Map from "ol/Map";
import Overlay from "ol/Overlay";
import { getTopLeft, getCenter } from "ol/extent";
import TileLayer from "ol/layer/Tile";
import VectorLayer from "ol/layer/Vector";
import WMTS from "ol/source/WMTS";
import VectorSource from "ol/source/Vector";
import { Icon, Style } from "ol/style";
import WMTSTileGrid from "ol/tilegrid/WMTS";
import View from "ol/View";
import type { Coordinate } from "ol/coordinate";
import type Feature from "ol/Feature";
import type { FeatureLike } from "ol/Feature";
import { useEffect, useRef, useState, useCallback, type Ref, useImperativeHandle } from "react";

import "ol/ol.css";

import { proj4Config } from "../../map/proj4";
import type { LocationFeature } from "../../types/api";
import locationIconSvg from "../../../public/location_icon.svg?raw";
import { LayerSwitcher, type BaseLayerId } from "./LayerSwitcher";
import { LocationPopup } from "./LocationPopup";

const proj = proj4Config.projection!;

function createLocationIconSrc(color: string) {
  const svg = locationIconSvg
    .replace('fill="#ffffff"', `fill="${color}"`)
    .replace(
      "<g>",
      '<defs><filter id="location-shadow" x="-30%" y="-30%" width="160%" height="160%"><feDropShadow dx="0" dy="24" stdDeviation="18" flood-color="#000000" flood-opacity="0.35" /></filter></defs><g>',
    )
    .replace("<path ", '<path filter="url(#location-shadow)" stroke="#ffffff" stroke-width="14" ')
    .replace("</svg>", '<circle cx="197.849" cy="137.329" r="49.191" fill="#ffffff" /></svg>');

  return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`;
}

function createLocationPointStyle(color: string) {
  return new Style({
    image: new Icon({
      src: createLocationIconSrc(color),
      width: 32,
      height: 32,
      anchor: [0.5, 1],
    }),
  });
}

const publicPointStyle = createLocationPointStyle("#22c55e");
const privatePointStyle = createLocationPointStyle("#ef4444");

function locationStyle(feature: FeatureLike): Style {
  return feature.get("is_public") === true ? publicPointStyle : privatePointStyle;
}

type SelectedLocation = {
  id: number;
  coordinate: Coordinate;
};

export type TerrainMapHandle = {
  focusLocation: (id: number) => void;
};

type TerrainMapProps = {
  locations: LocationFeature[];
  currentUserId: number | null;
  onUpdate: (input: {
    id: number;
    title: string;
    description: string;
    isPublic: boolean;
  }) => Promise<unknown>;
  ref?: Ref<TerrainMapHandle>;
  onVisibleIdsChange: (ids: number[]) => void;
  onDelete: (id: number) => Promise<unknown>;
};

export function TerrainMap({
  locations,
  currentUserId,
  onUpdate,
  onDelete,
  onVisibleIdsChange,
  ref,
}: TerrainMapProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const popupRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<Map | null>(null);
  const vectorSourceRef = useRef<VectorSource | null>(null);
  const onVisibleIdsChangeRef = useRef(onVisibleIdsChange);
  const lastVisibleKeyRef = useRef<string | null>(null);
  const baseLayersRef = useRef<Record<BaseLayerId, TileLayer<WMTS>> | null>(null);
  const overlayRef = useRef<Overlay | null>(null);

  const [activeLayer, setActiveLayer] = useState<BaseLayerId>("maastokartta");
  const [selected, setSelected] = useState<SelectedLocation | null>(null);

  const selectedFeature = selected
    ? locations.find((location) => location.id === selected.id)
    : undefined;

  useEffect(() => {
    onVisibleIdsChangeRef.current = onVisibleIdsChange;
  }, [onVisibleIdsChange]);
  const updateVisibleFeatures = useCallback(() => {
    const map = mapRef.current;
    const source = vectorSourceRef.current;
    const size = map?.getSize();
    if (!map || !source || !size) {
      return;
    }
    const extent = map.getView().calculateExtent(size);
    const ids = source
      .getFeaturesInExtent(extent)
      .map((feature) => Number(feature.getId()))
      .filter((id) => Number.isFinite(id));
    // Avoid notifying the parent when the visible set did not change.
    const key = ids.sort((a, b) => a - b).join(",");
    if (key === lastVisibleKeyRef.current) {
      return;
    }
    lastVisibleKeyRef.current = key;
    onVisibleIdsChangeRef.current?.(ids);
  }, []);

  useImperativeHandle(
    ref,
    () => ({
      focusLocation(id: number) {
        const map = mapRef.current;
        const source = vectorSourceRef.current;
        if (!map || !source) {
          return;
        }
        const feature = source.getFeatureById(id);
        if (!feature || Array.isArray(feature)) {
          return;
        }
        const geometry = feature.getGeometry();
        if (!geometry) {
          return;
        }
        const coordinate = getCenter(geometry.getExtent());
        const view = map.getView();
        view.animate({
          center: coordinate,
          zoom: Math.max(view.getZoom() ?? 12, 14),
          duration: 300,
        });
        // Reuse the existing popup state.
        setSelected({ id, coordinate });
      },
    }),
    [],
  );

  useEffect(() => {
    if (!containerRef.current || !popupRef.current) {
      return;
    }

    const tileGrid = new WMTSTileGrid({
      origin: getTopLeft(proj4Config.extent),
      resolutions: proj4Config.resolutions,
      matrixIds: proj4Config.matrixIds,
    });

    const makeSource = (layer: string) =>
      new WMTS({
        requestEncoding: "REST",
        format: "image/png",
        projection: proj,
        tileGrid,
        wrapX: false,
        style: "default",
        url: `/proxy/wmts/1.0.0/${layer}/default/ETRS-TM35FIN/{TileMatrix}/{TileRow}/{TileCol}.png`,
        layer,
        matrixSet: "ETRS-TM35FIN",
      });

    const maastokartta = new TileLayer({ visible: true, source: makeSource("maastokartta") });
    const ortokuva = new TileLayer({ visible: false, source: makeSource("ortokuva") });
    baseLayersRef.current = { maastokartta, ortokuva };

    const overlay = new Overlay({
      element: popupRef.current,
      positioning: "bottom-center",
      offset: [0, -12],
      stopEvent: true,
    });
    overlayRef.current = overlay;

    const map = new Map({
      target: containerRef.current,
      layers: [maastokartta, ortokuva],
      overlays: [overlay],
      view: new View({
        projection: proj,
        center: [500000, 7200000],
        zoom: 10,
      }),
    });
    mapRef.current = map;

    map.on("moveend", updateVisibleFeatures);
    map.on("change:size", updateVisibleFeatures);

    map.on("singleclick", (event) => {
      const feature = map.forEachFeatureAtPixel(event.pixel, (candidate) => candidate, {
        hitTolerance: 10,
      }) as Feature | undefined;

      if (feature && feature.getId() !== undefined) {
        setSelected({
          id: Number(feature.getId()),
          coordinate: event.coordinate,
        });
      } else {
        setSelected(null);
      }
    });

    return () => {
      map.un("moveend", updateVisibleFeatures);
      map.un("change:size", updateVisibleFeatures);
      map.setTarget(undefined);
      mapRef.current = null;
      baseLayersRef.current = null;
      overlayRef.current = null;
    };
  }, [updateVisibleFeatures]);

  useEffect(() => {
    const overlay = overlayRef.current;
    if (!overlay) {
      return;
    }
    overlay.setPosition(selected ? selected.coordinate : undefined);
  }, [selected]);

  useEffect(() => {
    const layers = baseLayersRef.current;
    if (!layers) {
      return;
    }
    layers.maastokartta.setVisible(activeLayer === "maastokartta");
    layers.ortokuva.setVisible(activeLayer === "ortokuva");
  }, [activeLayer]);

  useEffect(() => {
    const map = mapRef.current;
    if (!map) {
      return;
    }

    const source = new VectorSource({
      features: new GeoJSON().readFeatures(
        { type: "FeatureCollection", features: locations },
        { featureProjection: proj },
      ),
    });
    vectorSourceRef.current = source;
    const layer = new VectorLayer({ source, style: locationStyle });
    map.addLayer(layer);

    updateVisibleFeatures();
    if (!map.getSize()) {
      map.once("postrender", updateVisibleFeatures);
    }

    return () => {
      map.removeLayer(layer);
      vectorSourceRef.current = null;
    };
  }, [locations, updateVisibleFeatures]);

  return (
    <div ref={containerRef} className="relative h-full w-full">
      <LayerSwitcher active={activeLayer} onChange={setActiveLayer} />
      <div ref={popupRef}>
        {selected && selectedFeature && (
          <LocationPopup
            key={selected.id}
            id={selected.id}
            properties={selectedFeature.properties}
            currentUserId={currentUserId}
            onUpdate={onUpdate}
            onDelete={onDelete}
            onClose={() => setSelected(null)}
          />
        )}
      </div>
    </div>
  );
}
