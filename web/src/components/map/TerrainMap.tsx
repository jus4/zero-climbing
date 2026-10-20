import GeoJSON from "ol/format/GeoJSON";
import Map from "ol/Map";
import Overlay from "ol/Overlay";
import { getTopLeft } from "ol/extent";
import TileLayer from "ol/layer/Tile";
import VectorLayer from "ol/layer/Vector";
import WMTS from "ol/source/WMTS";
import VectorSource from "ol/source/Vector";
import { Circle as CircleStyle, Fill, Stroke, Style } from "ol/style";
import WMTSTileGrid from "ol/tilegrid/WMTS";
import View from "ol/View";
import type { Coordinate } from "ol/coordinate";
import type Feature from "ol/Feature";
import { useEffect, useRef, useState } from "react";

import "ol/ol.css";

import { proj4Config } from "../../map/proj4";
import type { LocationFeature, LocationFeatureProperties } from "../../types/api";
import { LayerSwitcher, type BaseLayerId } from "./LayerSwitcher";
import { LocationPopup } from "./LocationPopup";

const proj = proj4Config.projection!;

const pointStyle = new Style({
  image: new CircleStyle({
    radius: 6,
    fill: new Fill({ color: "#ef4444" }),
    stroke: new Stroke({ color: "#ffffff", width: 1.5 }),
  }),
});

type SelectedLocation = {
  id: number;
  properties: LocationFeatureProperties;
  coordinate: Coordinate;
};

type TerrainMapProps = {
  locations: LocationFeature[];
  currentUserId: number | null;
  onUpdate: (input: { id: number; title: string; description: string }) => Promise<unknown>;
  onDelete: (id: number) => Promise<unknown>;
};

export function TerrainMap({ locations, currentUserId, onUpdate, onDelete }: TerrainMapProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const popupRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<Map | null>(null);
  const baseLayersRef = useRef<Record<BaseLayerId, TileLayer<WMTS>> | null>(null);
  const overlayRef = useRef<Overlay | null>(null);

  const [activeLayer, setActiveLayer] = useState<BaseLayerId>("maastokartta");
  const [selected, setSelected] = useState<SelectedLocation | null>(null);

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

    map.on("singleclick", (event) => {
      const feature = map.forEachFeatureAtPixel(event.pixel, (candidate) => candidate, {
        hitTolerance: 10,
      }) as Feature | undefined;

      if (feature && feature.getId() !== undefined) {
        setSelected({
          id: Number(feature.getId()),
          properties: feature.getProperties() as LocationFeatureProperties,
          coordinate: event.coordinate,
        });
      } else {
        setSelected(null);
      }
    });

    return () => {
      map.setTarget(undefined);
      mapRef.current = null;
      baseLayersRef.current = null;
      overlayRef.current = null;
    };
  }, []);

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
    const layer = new VectorLayer({ source, style: pointStyle });
    map.addLayer(layer);

    return () => {
      map.removeLayer(layer);
    };
  }, [locations]);

  return (
    <div ref={containerRef} className="relative h-full w-full">
      <LayerSwitcher active={activeLayer} onChange={setActiveLayer} />
      <div ref={popupRef}>
        {selected && (
          <LocationPopup
            key={selected.id}
            id={selected.id}
            properties={selected.properties}
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
