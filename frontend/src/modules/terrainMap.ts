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
import type Feature from "ol/Feature";

const proj = proj4Config.projection!;

type LocationProperties = {
  title?: unknown;
  description?: unknown;
  user?: unknown;
  user_id?: unknown;
};

type EditPayload = {
  title: string;
  description: string;
};

const csrfCookieName = "csrf_token";

export class TerrainMap {
  private map: Map | null = null;
  private layer: VectorLayer<VectorSource> | null = null;
  private popup: Overlay | null = null;
  private popupEl: HTMLDivElement | null = null;
  private titleEl: HTMLDivElement | null = null;
  private descriptionEl: HTMLDivElement | null = null;
  private metaEl: HTMLDivElement | null = null;
  private actionsEl: HTMLDivElement | null = null;
  private editFormEl: HTMLDivElement | null = null;
  private editTitleInput: HTMLInputElement | null = null;
  private editDescriptionInput: HTMLTextAreaElement | null = null;
  private errorEl: HTMLDivElement | null = null;
  private currentUserId: number = 0;
  private activeFeature: Feature | null = null;

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
      const feature = this.map!.forEachFeatureAtPixel(
        evt.pixel,
        (f) => f,
        { hitTolerance: 10 },
      ) as Feature | undefined;
      if (feature) {
        this.showPopup(evt.coordinate, feature);
      } else {
        this.hidePopup();
      }
    });

    this.loadCurrentUser();
    this.loadLocations();
  }

  private async loadCurrentUser() {
    try {
      const res = await fetch("/api/me", { credentials: "same-origin" });
      if (!res.ok) return;
      const me = await res.json();
      this.currentUserId = typeof me.id === "number" ? me.id : 0;
    } catch {
      this.currentUserId = 0;
    }
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

    this.layer = layer;
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

    const title = this.textDiv("600", "15px", "#0f172a", "6px", "16px");
    const description = this.textDiv("400", "13px", "#334155", "8px", "0");
    description.style.whiteSpace = "pre-wrap";
    description.style.wordBreak = "break-word";

    const meta = this.textDiv("400", "12px", "#94a3b8", "8px", "0");
    meta.style.borderTop = "1px solid #e2e8f0";
    meta.style.paddingTop = "8px";

    const actions = document.createElement("div");
    actions.style.display = "none";
    actions.style.marginTop = "10px";
    actions.style.display = "flex";
    actions.style.gap = "8px";
    actions.style.borderTop = "1px solid #e2e8f0";
    actions.style.paddingTop = "10px";

    const editButton = this.button("Edit", "#f97316", "#ffffff");
    editButton.addEventListener("click", () => this.startEdit());
    const deleteButton = this.button("Delete", "#ef4444", "#ffffff");
    deleteButton.addEventListener("click", () => this.confirmDelete());

    actions.appendChild(editButton);
    actions.appendChild(deleteButton);

    const editForm = document.createElement("div");
    editForm.style.display = "none";

    const editTitle = document.createElement("input");
    editTitle.type = "text";
    editTitle.style.boxSizing = "border-box";
    editTitle.style.width = "100%";
    editTitle.style.padding = "6px 8px";
    editTitle.style.marginBottom = "6px";
    editTitle.style.fontSize = "14px";
    editTitle.style.border = "1px solid #cbd5e1";
    editTitle.style.borderRadius = "4px";

    const editDescription = document.createElement("textarea");
    editDescription.rows = 3;
    editDescription.style.boxSizing = "border-box";
    editDescription.style.width = "100%";
    editDescription.style.padding = "6px 8px";
    editDescription.style.marginBottom = "8px";
    editDescription.style.fontSize = "13px";
    editDescription.style.fontFamily = "inherit";
    editDescription.style.border = "1px solid #cbd5e1";
    editDescription.style.borderRadius = "4px";
    editDescription.style.resize = "vertical";

    const editActions = document.createElement("div");
    editActions.style.display = "flex";
    editActions.style.gap = "8px";
    editActions.style.justifyContent = "flex-end";

    const saveButton = this.button("Save", "#0f172a", "#ffffff");
    saveButton.addEventListener("click", () => this.saveEdit());
    const cancelButton = this.button("Cancel", "#ffffff", "#0f172a");
    cancelButton.style.border = "1px solid #cbd5e1";
    cancelButton.addEventListener("click", () => this.stopEdit());

    editActions.appendChild(cancelButton);
    editActions.appendChild(saveButton);

    editForm.appendChild(editTitle);
    editForm.appendChild(editDescription);
    editForm.appendChild(editActions);

    const error = document.createElement("div");
    error.style.display = "none";
    error.style.fontSize = "12px";
    error.style.color = "#dc2626";
    error.style.marginTop = "8px";

    card.appendChild(close);
    card.appendChild(title);
    card.appendChild(description);
    card.appendChild(meta);
    card.appendChild(actions);
    card.appendChild(editForm);
    card.appendChild(error);
    element.appendChild(card);

    this.titleEl = title;
    this.descriptionEl = description;
    this.metaEl = meta;
    this.actionsEl = actions;
    this.editFormEl = editForm;
    this.editTitleInput = editTitle;
    this.editDescriptionInput = editDescription;
    this.errorEl = error;

    const popup = new Overlay({
      element,
      positioning: "bottom-center",
      offset: [0, -12],
      stopEvent: true,
    });
    this.map!.addOverlay(popup);

    return { popup, element };
  }

  private textDiv(
    weight: string,
    size: string,
    color: string,
    marginBottom: string,
    paddingRight: string,
  ): HTMLDivElement {
    const div = document.createElement("div");
    div.style.fontWeight = weight;
    div.style.fontSize = size;
    div.style.color = color;
    div.style.marginBottom = marginBottom;
    div.style.paddingRight = paddingRight;
    return div;
  }

  private button(text: string, bg: string, color: string): HTMLButtonElement {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.textContent = text;
    btn.style.padding = "5px 12px";
    btn.style.fontSize = "13px";
    btn.style.fontWeight = "600";
    btn.style.border = "none";
    btn.style.borderRadius = "4px";
    btn.style.cursor = "pointer";
    btn.style.backgroundColor = bg;
    btn.style.color = color;
    return btn;
  }

  private showPopup(coord: Coordinate, feature: Feature) {
    if (!this.popup || !this.popupEl) return;

    this.activeFeature = feature;
    const props = feature.getProperties() as LocationProperties;
    this.enterViewMode(props);

    this.popup.setPosition(coord);
    this.popupEl.hidden = false;
  }

  private hidePopup() {
    this.activeFeature = null;
    if (this.popup) this.popup.setPosition(undefined);
    if (this.popupEl) this.popupEl.hidden = true;
  }

  private enterViewMode(props: LocationProperties) {
    if (this.actionsEl) {
      const ownerId = typeof props.user_id === "number" ? props.user_id : 0;
      this.actionsEl.style.display =
        this.currentUserId && ownerId === this.currentUserId ? "flex" : "none";
    }
    if (this.editFormEl) this.editFormEl.style.display = "none";
    if (this.titleEl) this.titleEl.style.display = "block";
    if (this.descriptionEl) this.descriptionEl.style.display = "block";
    if (this.metaEl) this.metaEl.style.display = "block";
    this.setError(null);

    const title = typeof props.title === "string" ? props.title : "";
    const description =
      typeof props.description === "string" ? props.description : "";
    const user = typeof props.user === "string" ? props.user : "";

    if (this.titleEl) this.titleEl.textContent = title || "Untitled";
    if (this.descriptionEl)
      this.descriptionEl.textContent = description || "No comments.";
    if (this.metaEl) this.metaEl.textContent = user ? `Added by ${user}` : "";
  }

  private startEdit() {
    if (!this.editFormEl || !this.editTitleInput || !this.editDescriptionInput)
      return;
    const props = this.activeFeature?.getProperties() as
      | LocationProperties
      | undefined;
    if (!props) return;

    this.editTitleInput.value =
      typeof props.title === "string" ? props.title : "";
    this.editDescriptionInput.value =
      typeof props.description === "string" ? props.description : "";

    if (this.titleEl) this.titleEl.style.display = "none";
    if (this.descriptionEl) this.descriptionEl.style.display = "none";
    if (this.metaEl) this.metaEl.style.display = "none";
    if (this.actionsEl) this.actionsEl.style.display = "none";
    this.editFormEl.style.display = "block";
    this.setError(null);
    this.editTitleInput.focus();
  }

  private stopEdit() {
    const props = this.activeFeature?.getProperties() as
      | LocationProperties
      | undefined;
    if (!props) return;
    this.enterViewMode(props);
  }

  private async saveEdit() {
    const feature = this.activeFeature;
    if (!feature) return;
    if (!this.editTitleInput || !this.editDescriptionInput) return;

    const title = this.editTitleInput.value.trim();
    if (!title) {
      this.setError("Title cannot be empty.");
      return;
    }
    this.setError(null);

    const description = this.editDescriptionInput.value.trim();
    const id = feature.getId();
    if (id === undefined) return;

    const body: EditPayload = { title, description };

    const ok = await this.sendJson(`/api/locations/${id}/edit`, "POST", body);
    if (!ok) {
      this.setError("Could not save changes.");
      return;
    }

    feature.set("title", title);
    feature.set("description", description);
    this.enterViewMode(feature.getProperties() as LocationProperties);
  }

  private confirmDelete() {
    const feature = this.activeFeature;
    if (!feature) return;
    const id = feature.getId();
    if (id === undefined) return;

    const swal = (window as unknown as { Swal?: { fire: (options: Record<string, unknown>) => Promise<{ isConfirmed: boolean }> } }).Swal;
    const runDelete = async () => {
      const ok = await this.sendJson(`/api/locations/${id}`, "DELETE");
      if (!ok) {
        this.setError("Could not delete this location.");
        return;
      }
      if (this.layer && id !== undefined) {
        const source = this.layer.getSource();
        const target = source?.getFeatureById(id);
        if (target) source?.removeFeature(target);
      }
      this.hidePopup();
    };

    if (swal && typeof swal.fire === "function") {
      swal.fire({
        title: "Delete this location?",
        text: "This cannot be undone.",
        icon: "warning",
        showCancelButton: true,
        confirmButtonColor: "#ef4444",
        cancelButtonColor: "#64748b",
        confirmButtonText: "Delete",
      }).then((result) => {
        if (result.isConfirmed) runDelete();
      });
    } else if (window.confirm("Delete this location? This cannot be undone.")) {
      void runDelete();
    }
  }

  private async sendJson(
    url: string,
    method: "POST" | "DELETE",
    body?: EditPayload,
  ): Promise<boolean> {
    try {
      const res = await fetch(url, {
        method,
        credentials: "same-origin",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": getCookie(csrfCookieName),
        },
        body: body ? JSON.stringify(body) : undefined,
      });
      return res.ok;
    } catch {
      return false;
    }
  }

  private setError(message: string | null) {
    if (!this.errorEl) return;
    this.errorEl.style.display = message ? "block" : "none";
    if (message) this.errorEl.textContent = message;
  }
}

function getCookie(name: string): string {
  const match = document.cookie.match(`(?:^|;\\s*)${name}=([^;]*)`);
  return match ? decodeURIComponent(match[1]) : "";
}