"use client";
import React, { useEffect, useState } from "react";
import { Button, Space, Table, Empty } from "antd";
import type { TableProps } from "antd";
import { DownloadOutlined } from "@ant-design/icons";
import { formatLocalDate } from "@/utils/date";

interface LocationData {
  City: string;
  UserCount: number;
}

const columns: TableProps<LocationData>["columns"] = [
  {
    title: "Város",
    dataIndex: "City",
    key: "City",
  },
  {
    title: "Látogatók",
    dataIndex: "UserCount",
    key: "UserCount",
    sorter: (a, b) => a.UserCount - b.UserCount,
  },
];

const StatisticsTable = ({
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
        const params = new URLSearchParams();
        if (site) params.set("site", site);
        if (from) params.set("from", formatLocalDate(from));
        if (to) params.set("to", formatLocalDate(to));

        const res = await fetch(`/api/v1/get-locations?${params.toString()}`, {
          method: "GET",
          headers: { "Content-Type": "application/json" },
        });

        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        const data = await res.json();
        setLocations(data.locations ?? []);
      } catch (err) {
        console.error("Location fetch error:", err);
      }
    };

    fetchLocations();
  }, [mounted, from, to, site]);

  const exportCSV = () => {
    if (!locations.length) return;
    const headers = "Város,Látogatók";
    const rows = locations.map((loc) => `"${loc.City || "Ismeretlen"}",${loc.UserCount}`).join("\n");
    const blob = new Blob([`${headers}\n${rows}`], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", `locations_${site || "all"}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (!mounted) {
    return (
      <div>
        <p>Táblázat betöltése...</p>
      </div>
    );
  }

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "flex-end", marginBottom: 12 }}>
        <Button
          size="small"
          icon={<DownloadOutlined />}
          onClick={exportCSV}
          disabled={!locations.length}
        >
          CSV Export
        </Button>
      </div>
      {locations.length === 0 ? (
        <Empty description="Nincs megjeleníthető földrajzi adat" />
      ) : (
        <Table
          columns={columns}
          dataSource={locations}
          rowKey={(record) => record.City || "unknown"}
          pagination={{ pageSize: 8 }}
          size="small"
        />
      )}
    </div>
  );
};

export default StatisticsTable;
