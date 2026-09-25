import { Map, View } from "ol";
import TileLayer from "ol/layer/Tile";
import WMTS from "ol/source/WMTS";
import WMTSTileGrid from "ol/tilegrid/WMTS";
import { getTopLeft } from "ol/extent";
import { proj4Config } from "../config/proj4";

const proj = proj4Config.projection!;

export class TerrainMap {
  private map: Map | null = null;

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
  }
}
