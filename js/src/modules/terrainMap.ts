import { Map, View } from "ol";
import TileLayer from "ol/layer/Tile";
import WMTS from "ol/source/WMTS";
import WMTSTileGrid from "ol/tilegrid/WMTS";
import { getTopLeft } from "ol/extent";
import { proj4Config } from "../config/proj4";
import GeoJSON from "ol/format/GeoJSON";
import Overlay from "ol/Overlay";
import VectorLayer from "ol/layer/Vector";
import VectorSource from "ol/source/Vector";
import { Circle as CircleStyle, Fill, Stroke, Style } from "ol/style";
import type { Coordinate } from "ol/coordinate";

const proj = proj4Config.projection!;

type LocationProperties = {
  title?: unknown;
  description?: unknown;
  user?: unknown;
};

export class TerrainMap {
  private map: Map | null = null;
  private popup: Overlay | null = null;
  private popupEl: HTMLDivElement | null = null;
  private titleEl: HTMLDivElement | null = null;
  private descriptionEl: HTMLDivElement | null = null;
  private metaEl: HTMLDivElement | null = null;

  dispose() {
    if (this.map) {
      this.map.setTarget(undefined);
      this.map = null;
    }
  }

  init(containerId: string = "map") {
    const tileGrid = new WMTSTileGrid({
      origin: getTopLeft(proj4Config.extent),
      resolutions: proj4Config.resolutions,
      matrixIds: proj4Config.matrixIds,
    });

    const makeWmtsSource = (layer: string) =>
      new WMTS({
        requestEncoding: "REST" as const,
        format: "image/png",
        projection: proj,
        tileGrid,
        wrapX: false,
        style: "default",
        url: `/proxy/wmts/1.0.0/${layer}/default/ETRS-TM35FIN/{TileMatrix}/{TileRow}/{TileCol}.png`,
        layer,
        matrixSet: "ETRS-TM35FIN",
      });

    const maastokartta = new TileLayer({
      visible: true,
      source: makeWmtsSource("maastokartta"),
    });
    maastokartta.set("title", "Maastokartta");

    const ortokuva = new TileLayer({
      visible: false,
      source: makeWmtsSource("ortokuva"),
    });
    ortokuva.set("title", "Ortokuva");

    this.map = new Map({
      target: containerId,
      layers: [maastokartta, ortokuva],
      view: new View({
        projection: proj,
        center: [500000, 7200000],
        zoom: 10,
      }),
    });

    const { popup, element } = this.createPopup();
    this.popup = popup;
    this.popupEl = element;

    this.map.on("singleclick", (evt) => {
      const feature = this.map!.forEachFeatureAtPixel(evt.pixel, (f) => f, {
        hitTolerance: 10,
      });
      if (feature) {
        this.showPopup(evt.coordinate, feature.getProperties() as LocationProperties);
      } else {
        this.hidePopup();
      }
    });

    this.loadLocations();
  }

  private async loadLocations() {
    if (!this.map) return;

    let res: Response;
    try {
      res = await fetch("/api/locations", { credentials: "same-origin" });
    } catch {
      return;
    }
    if (!res.ok) return;

    const geoJSON = await res.json();

    const source = new VectorSource({
      features: new GeoJSON().readFeatures(geoJSON, { featureProjection: proj }),
    });

    const layer = new VectorLayer({
      source,
      style: new Style({
        image: new CircleStyle({
          radius: 6,
          fill: new Fill({ color: "#ef4444" }),
          stroke: new Stroke({ color: "#ffffff", width: 1.5 }),
        }),
      }),
    });

    this.map.addLayer(layer);
  }

  private createPopup() {
    const element = document.createElement("div");
    element.hidden = true;

    const card = document.createElement("div");
    card.style.position = "relative";
    card.style.minWidth = "220px";
    card.style.maxWidth = "320px";
    card.style.padding = "14px 16px";
    card.style.background = "#ffffff";
    card.style.border = "1px solid #e2e8f0";
    card.style.borderRadius = "8px";
    card.style.boxShadow = "0 4px 16px rgba(0, 0, 0, 0.12)";
    card.style.boxSizing = "border-box";
    card.style.fontFamily =
      "system-ui, -apple-system, 'Segoe UI', sans-serif";

    const close = document.createElement("button");
    close.type = "button";
    close.setAttribute("aria-label", "Close");
    close.textContent = "×";
    close.style.position = "absolute";
    close.style.top = "4px";
    close.style.right = "8px";
    close.style.border = "none";
    close.style.background = "transparent";
    close.style.fontSize = "20px";
    close.style.lineHeight = "1";
    close.style.color = "#64748b";
    close.style.cursor = "pointer";
    close.addEventListener("click", () => this.hidePopup());

    const title = document.createElement("div");
    title.style.fontWeight = "600";
    title.style.fontSize = "15px";
    title.style.color = "#0f172a";
    title.style.paddingRight = "16px";
    title.style.marginBottom = "6px";

    const description = document.createElement("div");
    description.style.fontSize = "13px";
    description.style.color = "#334155";
    description.style.whiteSpace = "pre-wrap";
    description.style.wordBreak = "break-word";
    description.style.marginBottom = "8px";

    const meta = document.createElement("div");
    meta.style.fontSize = "12px";
    meta.style.color = "#94a3b8";
    meta.style.borderTop = "1px solid #e2e8f0";
    meta.style.paddingTop = "8px";

    card.appendChild(close);
    card.appendChild(title);
    card.appendChild(description);
    card.appendChild(meta);
    element.appendChild(card);

    this.titleEl = title;
    this.descriptionEl = description;
    this.metaEl = meta;

    const popup = new Overlay({
      element,
      positioning: "bottom-center",
      offset: [0, -12],
      stopEvent: true,
    });
    this.map!.addOverlay(popup);

    return { popup, element };
  }

  private showPopup(coord: Coordinate, props: LocationProperties) {
    if (!this.popup || !this.popupEl) return;

    const title = typeof props.title === "string" ? props.title : "";
    const description = typeof props.description === "string" ? props.description : "";
    const user = typeof props.user === "string" ? props.user : "";

    if (this.titleEl) this.titleEl.textContent = title || "Untitled";
    if (this.descriptionEl) this.descriptionEl.textContent = description || "No comments.";
    if (this.metaEl) this.metaEl.textContent = user ? `Added by ${user}` : "";

    this.popup.setPosition(coord);
    this.popupEl.hidden = false;
  }

  private hidePopup() {
    if (this.popup) this.popup.setPosition(undefined);
    if (this.popupEl) this.popupEl.hidden = true;
  }
}
