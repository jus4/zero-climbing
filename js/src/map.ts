let mapInstance: import("./modules/terrainMap").TerrainMap | null = null;

async function initMap() {
  const mapContainer = document.getElementById("map");
  if (!mapContainer) return;

  if (mapInstance) {
    mapInstance.dispose();
    mapInstance = null;
  }

  const observer = new IntersectionObserver(async (entries) => {
    if (entries[0].isIntersecting) {
      observer.disconnect();
      const { TerrainMap } = await import("./modules/terrainMap");
      mapInstance = new TerrainMap();
      mapInstance.init();
    }
  });
  observer.observe(mapContainer);
}

document.addEventListener("DOMContentLoaded", initMap);
document.addEventListener("htmx:afterSettle", initMap);
