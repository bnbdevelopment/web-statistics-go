"use client";
import { useEffect, useState } from "react";
import {
  Card,
  Statistic,
  Select,
  DatePicker,
  Row,
  Col,
  Typography,
} from "antd";
import {
  AreaChart,
  Area,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  PieChart,
  Pie,
} from "recharts";
import { useSearchParams } from "next/navigation";
import { useRouter } from "next/navigation";
import MapComponent from "../Map/Map.component";
import StatisticsTable from "../StatisticsTable/StatisticsTable.component";

const { RangePicker } = DatePicker;
const { Title } = Typography;

export default function Home() {
  const [sites, setSites] = useState<string[]>([]);
  const searchParams = useSearchParams();
  const [visitors, setVisitors] = useState(0);
  const [selectedSite, setSelectedSite] = useState(
    searchParams.get("site") || "",
  );
  const [visitToChart, setVisitToChart] = useState<
    { interval: number; uniqueSessions: number; totalRequests: number }[]
  >([]);
  const [spentTime, setSpentTime] = useState(0);
  const [activeUsers, setActiveUsers] = useState(0);
  const [fromDate, setFromDate] = useState<Date | null>(null);
  const [toDate, setToDate] = useState<Date | null>(null);
  const [sitesTraffic, setSitesTraffic] = useState<
    { page: string; count: number }[]
  >([]);
  const router = useRouter();

  useEffect(() => {
    fetch("/api/v1/get-sites", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    })
      .then((response) => response.json())
      .then((data) => setSites(data.sites || []))
      .catch((error) => console.error("Error fetching stats:", error));
  }, []);

  useEffect(() => {
    // visitors
    fetch(
      `/api/v1/traffic?page=${selectedSite}${
        fromDate ? `&from=${fromDate.toISOString().split("T")[0]}` : ""
      }${toDate ? `&to=${toDate.toISOString().split("T")[0]}` : ""}`,
      { method: "POST", headers: { "Content-Type": "application/json" } },
    )
      .then((response) => response.json())
      .then((data) => setVisitors(data.traffic || 0))
      .catch((error) => console.error("Error fetching stats:", error));

    // chart
    fetch(
      `/api/v1/graph?page=${selectedSite}&intervals=${
        fromDate && toDate
          ? Math.ceil(
              Math.sqrt(
                Math.ceil(
                  (toDate?.getTime() - fromDate?.getTime()) /
                    (1000 * 60 * 60 * 24),
                ) * 24,
              ),
            )
          : 24
      }${fromDate ? `&from=${fromDate.toISOString().split("T")[0]}` : ""}${
        toDate ? `&to=${toDate.toISOString().split("T")[0]}` : ""
      }`,
      { method: "POST", headers: { "Content-Type": "application/json" } },
    )
      .then((response) => response.json())
      .then((data) => setVisitToChart(data || []))
      .catch((error) => console.error("Error fetching stats:", error));

    fetch(`/api/v1/active?page=${selectedSite}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    })
      .then((response) => response.json())
      .then((data) => setActiveUsers(data.count || 0))
      .catch((error) => console.error("Error fetching stats:", error));

    fetch(
      `/api/v1/time?page=${selectedSite}${fromDate ? `&from=${fromDate.toISOString().split("T")[0]}` : ""}${
        toDate ? `&to=${toDate.toISOString().split("T")[0]}` : ""
      }`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
      },
    )
      .then((response) => response.json())
      .then((data) => setSpentTime(data.avgTimeSpent || 0))
      .catch((error) => console.error("Error fetching stats:", error));

    fetch(
      `/api/v1/sites?page=${selectedSite}${fromDate ? `&from=${fromDate.toISOString().split("T")[0]}` : ""}${
        toDate ? `&to=${toDate.toISOString().split("T")[0]}` : ""
      }`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
      },
    )
      .then((response) => response.json())
      .then((data) => setSitesTraffic(data || []))
      .catch((error) => console.error("Error fetching stats:", error));
  }, [selectedSite, fromDate, toDate]);

  return (
    <div style={{ padding: "1rem" }}>
      <Title level={2} style={{ textAlign: "center", marginBottom: "1rem" }}>
        Statisztikák
      </Title>

      {/* Filters */}
      <Row gutter={[16, 16]} justify="center">
        <Col xs={24} sm={12}>
          <Select
            style={{ width: "100%" }}
            placeholder="Válassz webhelyet"
            value={selectedSite || ""}
            options={sites
              .filter((site) => site !== "")
              .map((site: string) => ({ value: site, label: site }))}
            onChange={(value) => {
              setSelectedSite(value);
              const params = new URLSearchParams(searchParams.toString());
              params.set("site", value);
              router.replace(`?${params.toString()}`);
            }}
          />
        </Col>
        <Col xs={24} sm={12}>
          <RangePicker
            style={{ width: "100%" }}
            onChange={(e: any) => {
              setFromDate(e[0]?.["$d"] || null);
              setToDate(e[1]?.["$d"] || null);
            }}
          />
        </Col>
      </Row>

      {/* Stats Cards */}
      <Row gutter={[16, 16]} style={{ marginTop: "1rem" }}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic title="Látogatók száma" value={visitors} />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Oldalon töltött átlagos idő"
              value={spentTime}
              precision={1}
              suffix="min"
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic title="Aktív látogatók" value={activeUsers} />
          </Card>
        </Col>
      </Row>

      {/* Chart */}
      <Card title="Látogatók" style={{ marginTop: "1rem" }}>
        {visitToChart && visitToChart.length > 0 && (
          <div style={{ width: "100%", height: 200 }}>
            <ResponsiveContainer>
              <AreaChart data={visitToChart}>
                <defs>
                  <linearGradient
                    id="colorVisitors"
                    x1="0"
                    y1="0"
                    x2="0"
                    y2="1"
                  >
                    <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8} />
                    <stop offset="95%" stopColor="#8884d8" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <XAxis dataKey="interval" />
                <YAxis />
                <Tooltip />
                <Area
                  type="monotone"
                  dataKey="uniqueSessions"
                  stroke="#8884d8"
                  fill="url(#colorVisitors)"
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        )}
      </Card>
      <Card title="Oldalak látogatottsága" style={{ marginTop: "1rem" }}>
        <div style={{ width: "100%", height: 400 }}>
          <ResponsiveContainer>
            <PieChart>
              <Pie
                label={({ name, percent }: any) =>
                  `${name} ${(percent * 100).toFixed(0)}%`
                }
                nameKey={"page"}
                data={sitesTraffic}
                cx="50%"
                cy="50%"
                innerRadius={60}
                outerRadius={80}
                fill="#8884d8"
                dataKey="count"
              />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </Card>
      <Card title="Látogatók eloszlása" style={{ marginTop: "1rem" }}>
        <MapComponent from={fromDate} to={toDate} site={selectedSite} />
      </Card>
      <Card title="Látogatók eloszlása táblázat" style={{ marginTop: "1rem" }}>
        <StatisticsTable from={fromDate} to={toDate} site={selectedSite} />
      </Card>
    </div>
  );
}
