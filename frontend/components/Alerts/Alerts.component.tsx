"use client";
import { useEffect, useState } from "react";
import { Alert, Empty } from "antd";
import { formatLocalDate } from "@/utils/date";

interface ArchetypeAlert {
  level: "info" | "warning" | "critical" | string;
  message: string;
  metric: string;
  value: number;
  threshold: number;
}

interface Props {
  site: string;
  from: Date | null;
  to: Date | null;
}

const Alerts = ({ site, from, to }: Props) => {
  const [alerts, setAlerts] = useState<ArchetypeAlert[]>([]);

  useEffect(() => {
    const fetchAlerts = async () => {
      try {
        const params = new URLSearchParams();
        if (site) params.set("site", site);
        if (from) params.set("from", formatLocalDate(from));
        if (to) params.set("to", formatLocalDate(to));

        const res = await fetch(`/api/v1/statistics/alerts?${params.toString()}`);
        if (!res.ok) throw new Error("Alert fetch failed");
        const data = await res.json();
        setAlerts(data || []);
      } catch (err) {
        console.error("Alert fetch error", err);
        setAlerts([]);
      }
    };

    fetchAlerts();
  }, [site, from, to]);

  if (!alerts.length) {
    return <Empty description="Nincs kritikus riasztás" />;
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
      {alerts.map((alert, idx) => (
        <Alert
          key={idx}
          type={alert.level === "critical" ? "error" : alert.level}
          message={alert.message}
          description={`${alert.metric}: ${alert.value.toFixed(1)} (küszöb ${alert.threshold})`}
          showIcon
        />
      ))}
    </div>
  );
};

export default Alerts;
