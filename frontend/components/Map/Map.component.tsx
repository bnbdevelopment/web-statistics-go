"use client";

import React, { useEffect, useMemo, useState } from "react";
import dynamic from "next/dynamic";
import "leaflet/dist/leaflet.css";

interface LocationData {
  City: string;
  Latitude: number;
  Longitude: number;
  UserCount: number;
}

const MapContent = dynamic(
  async () => {
    const { MapContainer, TileLayer, useMap } = await import("react-leaflet");
    const L = await import("leaflet");
    await import("leaflet.heat");

    // 🔧 Leaflet default icon fix – csak egyszer
    delete (L.Icon.Default.prototype as any)._getIconUrl;
    L.Icon.Default.mergeOptions({
      iconRetinaUrl:
        "https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon-2x.png",
      iconUrl: "https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon.png",
      shadowUrl:
        "https://unpkg.com/leaflet@1.7.1/dist/images/marker-shadow.png",
    });

    const HeatmapLayer = ({
      points,
    }: {
      points: [number, number, number][];
    }) => {
      const map = useMap();

      useEffect(() => {
        if (!points.length) return;

        const layer = (L as any)
          .heatLayer(points, {
            radius: 25,
            blur: 15,
            maxZoom: 17,
          })
          .addTo(map);

        return () => {
          map.removeLayer(layer);
        };
      }, [map, points]);

      return null;
    };

    const ClientMap = ({ locations }: { locations: LocationData[] }) => {
      const heatmapPoints = useMemo(
        () =>
          locations.map(
            (l) =>
              [l.Latitude, l.Longitude, l.UserCount] as [
                number,
                number,
                number,
              ],
          ),
        [locations],
      );

      return (
        <MapContainer
          center={[47.1625, 19.5033]} // 🇭🇺 Hungary default
          zoom={7}
          style={{ height: "100%", width: "100%" }}
          whenReady={(map) => {
            setTimeout(() => map.target.invalidateSize(), 0);
          }}
        >
          <TileLayer
            attribution="&copy; OpenStreetMap contributors"
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          />
          <HeatmapLayer points={heatmapPoints} />
        </MapContainer>
      );
    };

    return ClientMap;
  },
  {
    ssr: false,
    loading: () => <p>Térkép betöltése…</p>,
  },
);

const MapComponent: React.FC = () => {
  const [locations, setLocations] = useState<LocationData[]>([]);

  useEffect(() => {
    const fetchLocations = async () => {
      try {
        const res = await fetch("/api/v1/get-locations", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
        });

        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`);
        }

        const data = await res.json();
        setLocations(data.locations ?? []);
      } catch (err) {
        console.error("Location fetch error:", err);
      }
    };

    fetchLocations();
  }, []);

  return (
    <div style={{ height: 500, width: "100%" }}>
      <MapContent locations={locations} />
    </div>
  );
};

export default MapComponent;
