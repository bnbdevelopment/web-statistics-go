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

const MapContent = dynamic(
  () =>
    import("react-leaflet").then((mod) => {
      const { MapContainer, TileLayer, useMap } = mod;
      const L = require("leaflet");
      require("leaflet.heat");

      // Fix for default icon not showing in Leaflet
      delete (L.Icon.Default.prototype as any)._getIconUrl;
      L.Icon.Default.mergeOptions({
        iconRetinaUrl:
          "https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon-2x.png",
        iconUrl: "https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon.png",
        shadowUrl:
          "https://unpkg.com/leaflet@1.7.1/dist/images/marker-shadow.png",
      });

      const HeatmapLayer = ({ positions }: { positions: number[][] }) => {
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

      interface MapProps {
        locations: LocationData[];
      }

      const ClientMap: React.FC<MapProps> = ({ locations }) => {
        // Prepare data for heatmap
        const heatmapData = locations.map((loc) => [
          loc.Latitude,
          loc.Longitude,
          loc.UserCount,
        ]);

        return (
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
            {heatmapData.length > 0 && <HeatmapLayer positions={heatmapData} />}
          </MapContainer>
        );
      };

      return ClientMap;
    }),
  { ssr: false, loading: () => <p>Loading Map...</p> },
);

const MapComponent: React.FC = () => {
  const [locations, setLocations] = useState<LocationData[]>([]);

  useEffect(() => {
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

  return (
    <div style={{ height: "500px", width: "100%" }}>
      <MapContent locations={locations} />
    </div>
  );
};

export default MapComponent;
