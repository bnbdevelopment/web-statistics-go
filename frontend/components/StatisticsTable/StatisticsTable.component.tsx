"use client";
import React, { useEffect, useState } from "react";
import { Table } from "antd";
import type { TableProps } from 'antd';

interface LocationData {
  City: string;
  UserCount: number;
}

const columns: TableProps<LocationData>['columns'] = [
    {
        title: 'Város',
        dataIndex: 'City',
        key: 'City',
    },
    {
        title: 'Látogatók',
        dataIndex: 'UserCount',
        key: 'UserCount',
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
      <div>
        <p>Táblázat betöltése...</p>
      </div>
    );
  }

  return (
    <Table columns={columns} dataSource={locations} />
  );
};

export default StatisticsTable;
