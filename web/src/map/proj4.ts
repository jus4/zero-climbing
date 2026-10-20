import proj4 from "proj4";
import { get } from "ol/proj";
import { register } from "ol/proj/proj4";

proj4.defs("EPSG:3067", "+proj=utm +zone=35 +ellps=GRS80 +units=m +no_defs");
register(proj4);

const projection = get("EPSG:3067");
const extent = [-548576, 6291456, 1548576, 8388608];
const resolutions = [8192, 4096, 2048, 1024, 512, 256, 128, 64, 32, 16, 8, 4, 2, 1, 0.5];
const matrixIds = resolutions.map((_, index) => String(index));

export const proj4Config = { projection, extent, resolutions, matrixIds };
