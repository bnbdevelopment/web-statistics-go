declare module "leaflet.heat" {
  import * as L from "leaflet";

  interface HeatLayerOptions {
    radius?: number;
    blur?: number;
    maxZoom?: number;
    minOpacity?: number;
    gradient?: { [key: number]: string };
  }

  type HeatLatLngTuple = [number, number, number?];

  function heatLayer(
    latlngs: HeatLatLngTuple[],
    options?: HeatLayerOptions,
  ): L.Layer;

  export = heatLayer;
}
