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
import {
  ArrowDownOutlined,
  ArrowUpOutlined,
  InfoCircleOutlined,
} from "@ant-design/icons";
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
  const [visitorsYesterday, setVisitorsYesterday] = useState(0);
  const [selectedSite, setSelectedSite] = useState(
    searchParams.get("site") || "",
  );
  const [visitToChart, setVisitToChart] = useState<
    { interval: number; uniqueSessions: number; totalRequests: number }[]
  >([]);
  const [spentTime, setSpentTime] = useState(0);
  const [spentTimeYesterday, setSpentTimeYesterday] = useState(0);
  const [activeUsers, setActiveUsers] = useState(0);
  const [fromDate, setFromDate] = useState<Date | null>(null);
  const [toDate, setToDate] = useState<Date | null>(null);
  const [sitesTraffic, setSitesTraffic] = useState<
    { page: string; count: number }[]
  >([]);
  const [bounceRate, setBounceRate] = useState(0);
  const [bounceRateYesterday, setBounceRateYesterday] = useState(0);

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
    const from = fromDate ? `&from=${fromDate.toISOString().split("T")[0]}` : "";
    const to = toDate ? `&to=${toDate.toISOString().split("T")[0]}` : "";

    let fromYesterday = "";
    let toYesterday = "";

    if (fromDate && toDate) {
      const diff = toDate.getTime() - fromDate.getTime();
      const prevFromDate = new Date(fromDate.getTime() - diff);
      const prevToDate = fromDate;
      fromYesterday = `&from=${prevFromDate.toISOString().split("T")[0]}`;
      toYesterday = `&to=${prevToDate.toISOString().split("T")[0]}`;
    } else {
      const now = new Date();
      const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000);
      const twoDaysAgo = new Date(now.getTime() - 48 * 60 * 60 * 1000);
      fromYesterday = `&from=${twoDaysAgo.toISOString().split("T")[0]}`;
      toYesterday = `&to=${yesterday.toISOString().split("T")[0]}`;
    }

    // visitors
    fetch(`/api/v1/traffic?page=${selectedSite}${from}${to}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    })
      .then((response) => response.json())
      .then((data) => setVisitors(data.traffic || 0))
      .catch((error) => console.error("Error fetching stats:", error));

    fetch(`/api/v1/traffic?page=${selectedSite}${fromYesterday}${toYesterday}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    })
      .then((response) => response.json())
      .then((data) => setVisitorsYesterday(data.traffic || 0))
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
      }${from}${to}`,
      { method: "POST", headers: { "Content-Type": "application/json" } },
    )
      .then((response) => response.json())
      .then((data) => setVisitToChart(data || []))
      .catch((error) => console.error("Error fetching stats:", error));

    fetch(`/api/v1/time?page=${selectedSite}${from}${to}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    })
      .then((response) => response.json())
      .then((data) => setSpentTime(data.avgTimeSpent || 0))
      .catch((error) => console.error("Error fetching stats:", error));

    fetch(`/api/v1/time?page=${selectedSite}${fromYesterday}${toYesterday}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    })
      .then((response) => response.json())
      .then((data) => setSpentTimeYesterday(data.avgTimeSpent || 0))
      .catch((error) => console.error("Error fetching stats:", error));

    fetch(`/api/v1/sites?page=${selectedSite}${from}${to}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    })
      .then((response) => response.json())
      .then((data) => setSitesTraffic(data || []))
      .catch((error) => console.error("Error fetching stats:", error));

    fetch(`/api/v1/bounce-rate?site=${selectedSite}${from}${to}`, {
      method: "GET",
      headers: { "Content-Type": "application/json" },
    })
      .then((response) => response.json())
      .then((data) => setBounceRate(data.bounceRate || 0))
      .catch((error) => console.error("Error fetching bounce rate:", error));

    fetch(
      `/api/v1/bounce-rate?site=${selectedSite}${fromYesterday}${toYesterday}`,
      {
        method: "GET",
        headers: { "Content-Type": "application/json" },
      },
    )
      .then((response) => response.json())
      .then((data) => setBounceRateYesterday(data.bounceRate || 0))
      .catch((error) =>
        console.error("Error fetching bounce rate yesterday:", error),
      );
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
                    <Statistic
                      title="Látogatók száma (fő)"
                      value={visitors}
                      valueStyle={{
                        color:
                          visitors > visitorsYesterday
                            ? "#3f8600"
                            : "#cf1322",
                      }}
                      prefix={
                        visitors > visitorsYesterday ? (
                          <ArrowUpOutlined />
                        ) : (
                          <ArrowDownOutlined />
                        )
                      }
                      suffix={
                        <span style={{ fontSize: "14px", marginLeft: "8px" }}>
                          {visitorsYesterday === 0 && visitors === 0
                            ? "0.00%"
                            : visitorsYesterday === 0
                              ? "100.00%"
                              : `${(
                                  ((visitors - visitorsYesterday) /
                                    visitorsYesterday) *
                                  100
                                ).toFixed(2)}%`}
                        </span>
                      }
                    />
                    <Typography.Text type="secondary">
                      a 24 órával ezelőttihez képest
                    </Typography.Text>
                  </Card>
                </Col>
                <Col xs={24}>
                  <Card style={{ height: "100%" }}>
                    <Statistic
                      title="Oldalon töltött átlagos idő (perc)"
                      value={spentTime}
                      precision={1}
                      suffix={
                        <span style={{ fontSize: "14px", marginLeft: "8px" }}>
                          {spentTimeYesterday === 0 && spentTime === 0
                            ? "0.00%"
                            : spentTimeYesterday === 0
                              ? "100.00%"
                              : `${(
                                  ((spentTime - spentTimeYesterday) /
                                    spentTimeYesterday) *
                                  100
                                ).toFixed(2)}%`}
                        </span>
                      }
                      valueStyle={{
                        color:
                          spentTime > spentTimeYesterday
                            ? "#3f8600"
                            : "#cf1322",
                      }}
                      prefix={
                        spentTime > spentTimeYesterday ? (
                          <ArrowUpOutlined />
                        ) : (
                          <ArrowDownOutlined />
                        )
                      }
                    />
                    <Typography.Text type="secondary">
                      a 24 órával ezelőttihez képest
                    </Typography.Text>
                  </Card>
                </Col>
                <Col xs={24}>
                  <Card style={{ height: "100%" }}>
                    <Statistic title="Aktív látogatók (fő)" value={activeUsers} />
                  </Card>
                </Col>
                <Col xs={24}>
                  <Card style={{ height: "100%" }}>
                    <Statistic
                      title={"Visszafordulási arány (%)"}
                      value={bounceRate}
                      precision={2}
                      suffix={
                        <span style={{ fontSize: "14px", marginLeft: "8px" }}>
                          {bounceRateYesterday === 0 && bounceRate === 0
                            ? "0.00%"
                            : bounceRateYesterday === 0
                              ? "100.00%"
                              : `${(
                                  ((bounceRate - bounceRateYesterday) /
                                    bounceRateYesterday) *
                                  100
                                ).toFixed(2)}%`}
                        </span>
                      }
                      valueStyle={{
                        color:
                          bounceRate < bounceRateYesterday
                            ? "#3f8600"
                            : "#cf1322",
                      }}
                      prefix={
                        bounceRate < bounceRateYesterday ? (
                          <ArrowDownOutlined />
                        ) : (
                          <ArrowUpOutlined />
                        )
                      }
                    />
                    <Typography.Text type="secondary">
                      a 24 órával ezelőttihez képest
                    </Typography.Text>
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
              <Card
                title="Top 5 leglátogatottabb oldal"
                style={{ height: "100%" }}
              >
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
              <Card
                title="Látogatók eloszlása táblázat"
                style={{ height: "100%" }}
              >
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
