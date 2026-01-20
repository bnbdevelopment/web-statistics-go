"use client";

import React, { useEffect, useState } from "react";
import dynamic from "next/dynamic";
import "leaflet/dist/leaflet.css";

interface LocationData {
  City: string;
  Latitude: number;
  Longitude: number;
  UserCount: number;
}

// Client-side only component for the HeatmapLayer
const Heatmap = dynamic(
  () =>
    import("react-leaflet").then((mod) => {
      const { useMap } = mod;
      const L = require("leaflet");
      require("leaflet.heat");

      const HeatmapLayer = (props: any) => {
        const { positions } = props;
        const map = useMap();

        useEffect(() => {
          if (!map) return;

          const heat = (L.heatLayer as any)(positions, {
            radius: 20,
            blur: 15,
            maxZoom: 17,
          }).addTo(map);

          return () => {
            map.removeLayer(heat);
          };
        }, [positions, map]);

        return null;
      };
      return HeatmapLayer;
    }),
  { ssr: false },
);

const MapComponent: React.FC = () => {
  const [locations, setLocations] = useState<LocationData[]>([]);

  useEffect(() => {
    // Fix for default icon not showing in Leaflet
    const L = require("leaflet");
    delete (L.Icon.Default.prototype as any)._getIconUrl;
    L.Icon.Default.mergeOptions({
      iconRetinaUrl:
        "https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon-2x.png",
      iconUrl: "https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon.png",
      shadowUrl:
        "https://unpkg.com/leaflet@1.7.1/dist/images/marker-shadow.png",
    });

    const fetchLocations = async () => {
      try {
        const response = await fetch("/api/v1/get-locations", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
        });
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        setLocations(data.locations || []);
      } catch (error) {
        console.error("Error fetching location data:", error);
      }
    };

    fetchLocations();
  }, []);

  const MapContainer = dynamic(
    () => import("react-leaflet").then((mod) => mod.MapContainer),
    { ssr: false },
  );
  const TileLayer = dynamic(
    () => import("react-leaflet").then((mod) => mod.TileLayer),
    { ssr: false },
  );

  // Prepare data for heatmap
  const heatmapData = locations.map((loc) => [
    loc.Latitude,
    loc.Longitude,
    loc.UserCount,
  ]);

  return (
    <div style={{ height: "500px", width: "100%" }}>
      <MapContainer
        center={[0, 0]} // Default center, will adjust based on data or user interaction
        zoom={2}
        scrollWheelZoom={false}
        style={{ height: "100%", width: "100%" }}
      >
        <TileLayer
          attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        {heatmapData.length > 0 && <Heatmap positions={heatmapData} />}
      </MapContainer>
    </div>
  );
};

export default MapComponent;
