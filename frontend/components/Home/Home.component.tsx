"use client";
import { useEffect, useState } from "react";
import {
  Card,
  Statistic,
  Row,
  Col,
  Layout,
  Divider,
  Typography,
} from "antd";
import {
  AreaChart,
  Area,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  BarChart,
  Bar,
  CartesianGrid,
} from "recharts";
import { useSearchParams } from "next/navigation";
import MapComponent from "../Map/Map.component";
import StatisticsTable from "../StatisticsTable/StatisticsTable.component";
import Header from "../Header/Header.component";
import Footer from "../Footer/Footer.component";
const { Content } = Layout;
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
    const fetchActiveUsers = () => {
      fetch(`/api/v1/active?page=${selectedSite}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
      })
        .then((response) => response.json())
        .then((data) => setActiveUsers(data.count || 0))
        .catch((error) => console.error("Error fetching active users:", error));
    };

    fetchActiveUsers();
    const interval = setInterval(fetchActiveUsers, 30000);

    return () => clearInterval(interval);
  }, [selectedSite]);

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
    <Layout style={{ minHeight: "100vh" }}>
      <Header
        sites={sites}
        selectedSite={selectedSite}
        onSiteChange={setSelectedSite}
        onDateChange={([from, to]) => {
          setFromDate(from);
          setToDate(to);
        }}
      />
      <Layout>
        <Content style={{ padding: "24px", background: "#f0f2f5" }}>
          {/* Quick & Relevant */}
          <Title level={4} style={{ marginBottom: "24px" }}>
            Gyors áttekintés
          </Title>
          <Row gutter={[24, 24]}>
            <Col xs={24} lg={16}>
              <Card title="Látogatók" style={{ height: "100%" }}>
                {visitToChart && visitToChart.length > 0 && (
                  <div style={{ width: "100%", height: 300 }}>
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
                            <stop
                              offset="5%"
                              stopColor="#8884d8"
                              stopOpacity={0.8}
                            />
                            <stop
                              offset="95%"
                              stopColor="#8884d8"
                              stopOpacity={0}
                            />
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
            </Col>
            <Col xs={24} lg={8}>
              <Row gutter={[24, 24]}>
                <Col xs={24}>
                  <Card style={{ height: "100%" }}>
                    <Statistic title="Látogatók száma" value={visitors} />
                  </Card>
                </Col>
                <Col xs={24}>
                  <Card style={{ height: "100%" }}>
                    <Statistic
                      title="Oldalon töltött átlagos idő"
                      value={spentTime}
                      precision={1}
                      suffix="min"
                    />
                  </Card>
                </Col>
                <Col xs={24}>
                  <Card style={{ height: "100%" }}>
                    <Statistic title="Aktív látogatók" value={activeUsers} />
                  </Card>
                </Col>
              </Row>
            </Col>
          </Row>

          <Divider style={{ marginTop: "48px", marginBottom: "48px" }}>
            <Title level={4}>Részletes elemzés</Title>
          </Divider>

          {/* Deeper Dive */}
          <Row gutter={[24, 24]}>
            <Col xs={24} lg={16}>
              <Card title="Látogatók eloszlása" style={{ height: "100%" }}>
                <MapComponent from={fromDate} to={toDate} site={selectedSite} />
              </Card>
            </Col>
            <Col xs={24} lg={8}>
              <Card title="Top 5 leglátogatottabb oldal" style={{ height: "100%" }}>
                <div style={{ width: "100%", height: 400 }}>
                  <ResponsiveContainer>
                    <BarChart
                      layout="vertical"
                      data={[...sitesTraffic]
                        .sort((a, b) => b.count - a.count)
                        .slice(0, 5)}
                      margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis type="number" />
                      <YAxis dataKey="page" type="category" />
                      <Tooltip />
                      <Bar dataKey="count" fill="#8884d8" />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </Card>
            </Col>
          </Row>

          <Row gutter={[24, 24]} style={{ marginTop: "24px" }}>
            <Col xs={24} lg={12}>
              <Card title="Oldalak látogatottsága" style={{ height: "100%" }}>
                <div style={{ width: "100%", height: 400 }}>
                  <ResponsiveContainer>
                    <BarChart
                      layout="vertical"
                      data={sitesTraffic}
                      margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis type="number" />
                      <YAxis dataKey="page" type="category" />
                      <Tooltip />
                      <Bar dataKey="count" fill="#8884d8" />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card title="Látogatók eloszlása táblázat" style={{ height: "100%" }}>
                <StatisticsTable
                  from={fromDate}
                  to={toDate}
                  site={selectedSite}
                />
              </Card>
            </Col>
          </Row>
        </Content>
      </Layout>
      <Footer />
    </Layout>
  );
}
