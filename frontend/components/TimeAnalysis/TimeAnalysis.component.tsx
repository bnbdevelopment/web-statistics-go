"use client";

import { useEffect, useState } from "react";
import { Card, Col, Row } from "antd";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  AreaChart,
  Area,
} from "recharts";



interface TimeAnalysisProps {
  site: string;
  from: Date | null;
  to: Date | null;
}

interface DayData {
  day: string;
  count: number;
}

interface HourData {
  hour: number;
  count: number;
}

const TimeAnalysis = ({ site, from, to }: TimeAnalysisProps) => {
  const [dayData, setDayData] = useState<DayData[]>([]);
  const [hourData, setHourData] = useState<HourData[]>([]);

  useEffect(() => {
    const fromDate = from ? `&from=${from.toISOString().split("T")[0]}` : "";
    const toDate = to ? `&to=${to.toISOString().split("T")[0]}` : "";

    // Fetch traffic by day of week
    fetch(
      `/api/v1/statistics/traffic-by-day-of-week?site=${site}${fromDate}${toDate}`,
    )
      .then((response) => response.json())
      .then((data) => setDayData(data || []))
      .catch((error) =>
        console.error("Error fetching traffic by day:", error),
      );

    // Fetch traffic by hour of day
    fetch(
      `/api/v1/statistics/traffic-by-hour-of-day?site=${site}${fromDate}${toDate}`,
    )
      .then((response) => response.json())
      .then((data) => setHourData(data || []))
      .catch((error) =>
        console.error("Error fetching traffic by hour:", error),
      );
  }, [site, from, to]);

  const formattedHourData = hourData.map(item => ({
    ...item,
    hourLabel: `${String(item.hour).padStart(2, '0')}:00`
  }));

  return (
    <Row gutter={[24, 24]} style={{ marginTop: "24px" }}>
      <Col xs={24} lg={12}>
        <Card title="Forgalom a hét napjai szerint (átlagos)">
          <div style={{ width: "100%", height: 400 }}>
            <ResponsiveContainer>
              <BarChart data={dayData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="day" />
                <YAxis />
                <Tooltip formatter={(value: number) => [value.toFixed(2), "Átlagos látogatások"]} />
                <Legend />
                <Bar dataKey="count" fill="#8884d8" name="Átlagos látogatások" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </Card>
      </Col>
      <Col xs={24} lg={12}>
        <Card title="Átlagos forgalom a nap órái szerint">
          <div style={{ width: "100%", height: 400 }}>
            <ResponsiveContainer>
              <AreaChart data={formattedHourData}>
                <defs>
                  <linearGradient id="colorTime" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#82ca9d" stopOpacity={0.8} />
                    <stop offset="95%" stopColor="#82ca9d" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="hourLabel" />
                <YAxis />
                <Tooltip formatter={(value: number) => [value.toFixed(2), "Átlagos látogatások"]} />
                <Legend />
                <Area
                  type="monotone"
                  dataKey="count"
                  stroke="#82ca9d"
                  fill="url(#colorTime)"
                  name="Átlagos látogatások"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </Card>
      </Col>
    </Row>
  );
};

export default TimeAnalysis;
