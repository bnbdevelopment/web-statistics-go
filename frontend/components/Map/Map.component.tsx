"use client";
import React, { useEffect, useMemo, useState } from "react";
import dynamic from "next/dynamic";
import "leaflet/dist/leaflet.css";
import type * as Leaflet from "leaflet";

interface LocationData {
  City: string;
  Latitude: number;
  Longitude: number;
  UserCount: number;
}

const MapContent = dynamic(
  async () => {
    const { MapContainer, TileLayer, useMap } = await import("react-leaflet");
    const L = (await import("leaflet")) as typeof Leaflet;
    if (typeof window !== "undefined") {
      if (!L.heatLayer) {
        await import("leaflet.heat");
      }
    }

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

        const L = (window as any).L;
        if (!L || !L.heatLayer) {
          console.error("L.heatLayer is not available");
          return;
        }

        const layer = L.heatLayer(points, {
          radius: 40,
          blur: 25,
          maxZoom: 17,
          max: 1.0,
          minOpacity: 0.5,
          gradient: {
            0.0: "blue",
            0.3: "cyan",
            0.5: "lime",
            0.7: "yellow",
            0.9: "orange",
            1.0: "red",
          },
        }).addTo(map);

        return () => {
          map.removeLayer(layer);
        };
      }, [map, points]);

      return null;
    };

    const ClientMap = ({ locations }: { locations: LocationData[] }) => {
      const heatmapPoints = useMemo(() => {
        if (!locations.length) return [];
        const maxUsers = Math.max(...locations.map((l) => l.UserCount), 1);

        return locations.map((l) => {
          const normalizedIntensity =
            Math.log(l.UserCount + 1) / Math.log(maxUsers + 1);
          const boostedIntensity = Math.min(normalizedIntensity * 1.5, 1); // 50% erősítés

          return [l.Latitude, l.Longitude, boostedIntensity] as [
            number,
            number,
            number,
          ];
        });
      }, [locations]);

      return (
        <MapContainer
          center={[47.1625, 19.5033]}
          zoom={7}
          style={{ height: "100%", width: "100%", minHeight: "500px" }}
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
    loading: () => (
      <div
        style={{
          height: 500,
          width: "100%",
          minHeight: 500,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <p>Térkép betöltése…</p>
      </div>
    ),
  },
);

const MapComponent = ({
  from,
  to,
  site,
}: {
  from: any;
  to: any;
  site: string;
}) => {
  const [locations, setLocations] = useState<LocationData[]>([]);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (typeof window === "undefined" || !mounted) return;

    const fetchLocations = async () => {
      try {
        const res = await fetch(
          `/api/v1/get-locations?page=${site}${from ? `&from=${from.toISOString().split("T")[0]}` : ""}${
            to ? `&to=${to.toISOString().split("T")[0]}` : ""
          }`,
          {
            method: "POST",
            headers: { "Content-Type": "application/json" },
          },
        );

        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        const data = await res.json();
        setLocations(data.locations ?? []);
      } catch (err) {
        console.error("Location fetch error:", err);
      }
    };

    fetchLocations();
  }, [mounted, from, to, site]);

  if (!mounted) {
    return (
      <div
        style={{
          height: 500,
          width: "100%",
          minHeight: 500,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <p>Térkép betöltése…</p>
      </div>
    );
  }

  return (
    <div style={{ height: 500, width: "100%", minHeight: 500, minWidth: 0 }}>
      <MapContent locations={locations} />
    </div>
  );
};

export default MapComponent;
