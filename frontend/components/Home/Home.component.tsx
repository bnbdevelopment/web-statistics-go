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
  Tooltip,
  Collapse,
} from "antd";
import {
  AreaChart,
  Area,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
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
import CohortAnalysis from "../CohortAnalysis/CohortAnalysis.component";
import AverageJourney from "../AverageJourney/AverageJourney.component";
import TimeAnalysis from "../TimeAnalysis/TimeAnalysis.component";
import Archetypes from "../Archetypes/Archetypes.component";
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
    const from = fromDate
      ? `&from=${fromDate.toISOString().split("T")[0]}`
      : "";
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

    fetch(
      `/api/v1/traffic?page=${selectedSite}${fromYesterday}${toYesterday}`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
      },
    )
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
          {/* --- Row 1: KPI Cards --- */}
          <Title level={4} style={{ marginBottom: "16px" }}>
            Pillanatkép
          </Title>
          <Row gutter={[24, 24]}>
            <Col xs={24} sm={12} lg={6}>
              <Card style={{ height: "100%" }}>
                <Statistic
                  title="Látogatók száma (fő)"
                  value={visitors}
                  valueStyle={{
                    color: visitors > visitorsYesterday ? "#3f8600" : "#cf1322",
                  }}
                  prefix={
                    visitors > visitorsYesterday ? (
                      <ArrowUpOutlined />
                    ) : (
                      <ArrowDownOutlined />
                    )
                  }
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card style={{ height: "100%" }}>
                <Statistic
                  title="Oldalon töltött átlagos idő (perc)"
                  value={spentTime}
                  precision={1}
                  valueStyle={{
                    color:
                      spentTime > spentTimeYesterday ? "#3f8600" : "#cf1322",
                  }}
                  prefix={
                    spentTime > spentTimeYesterday ? (
                      <ArrowUpOutlined />
                    ) : (
                      <ArrowDownOutlined />
                    )
                  }
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card style={{ height: "100%" }}>
                <Statistic
                  title={"Visszafordulási arány (%)"}
                  value={bounceRate}
                  precision={2}
                  valueStyle={{
                    color:
                      bounceRate < bounceRateYesterday ? "#3f8600" : "#cf1322",
                  }}
                  prefix={
                    bounceRate < bounceRateYesterday ? (
                      <ArrowUpOutlined />
                    ) : (
                      <ArrowDownOutlined />
                    )
                  }
                />
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card style={{ height: "100%" }}>
                <Statistic title="Aktív látogatók (fő)" value={activeUsers} />
              </Card>
            </Col>
          </Row>

          {/* --- Row 2: Main Visitors Chart --- */}
          <Row gutter={[24, 24]} style={{ marginTop: "24px" }}>
            <Col xs={24}>
              <Card title="Látogatói trend" style={{ height: "100%" }}>
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
                        <RechartsTooltip />
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
          </Row>

          {/* --- Row 3: Geo and Content --- */}
          <Row gutter={[24, 24]} style={{ marginTop: "24px" }}>
            <Col xs={24} lg={12}>
              <Card title="Látogatók eloszlása" style={{ height: "100%" }}>
                <MapComponent from={fromDate} to={toDate} site={selectedSite} />
              </Card>
            </Col>
            <Col xs={24} lg={12}>
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
                      <YAxis dataKey="page" type="category" width={120} />
                      <RechartsTooltip />
                      <Bar dataKey="count" fill="#8884d8" />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </Card>
            </Col>
          </Row>

          {/* --- Deeper Dive Section --- */}
          <Divider style={{ marginTop: "48px", marginBottom: "24px" }}>
            <Title level={4}>Mélyebb elemzések</Title>
          </Divider>

          <Collapse ghost defaultActiveKey={['archetypes']} accordion>
            <Collapse.Panel
              key="archetypes"
              header="Felhasználói Archetípusok (Automatikus elemzés)"
            >
              <Archetypes site={selectedSite} from={fromDate} to={toDate} />
            </Collapse.Panel>
            <Collapse.Panel
              key="time-analysis"
              header="Napszaki és heti bontású elemzés"
            >
              <TimeAnalysis site={selectedSite} from={fromDate} to={toDate} />
            </Collapse.Panel>
            <Collapse.Panel
              key="detailed-tables"
              header="Részletes táblázatok"
            >
              <Row gutter={[24, 24]}>
                <Col xs={24} lg={12}>
                  <Card
                    title="Oldalak látogatottsága"
                    style={{ height: "100%" }}
                  >
                    <div style={{ width: "100%", height: 400 }}>
                      <ResponsiveContainer>
                        <BarChart
                          layout="vertical"
                          data={sitesTraffic}
                          margin={{ top: 20, right: 30, left: 20, bottom: 5 }}
                        >
                          <CartesianGrid strokeDasharray="3 3" />
                          <XAxis type="number" />
                          <YAxis dataKey="page" type="category" width={120} />
                          <RechartsTooltip />
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
            </Collapse.Panel>
            <Collapse.Panel
              key="cohort-analysis"
              header={
                <span>
                  Diákok Visszatérése Hetente{" "}
                  <Tooltip title="Ez a nézet megmutatja, hogy az egy adott héten indult diákok közül hányan tértek vissza gyakorolni a következő hetekben.">
                    <InfoCircleOutlined />
                  </Tooltip>
                </span>
              }
            >
              <CohortAnalysis
                site={selectedSite}
                from={fromDate}
                to={toDate}
              />
            </Collapse.Panel>
            <Collapse.Panel
              key="sankey-diagram"
              header="Átlagos felhasználói útvonal elemzés (Sankey diagram)"
            >
              <AverageJourney
                site={selectedSite}
                from={fromDate}
                to={toDate}
              />
            </Collapse.Panel>
          </Collapse>
        </Content>
      </Layout>
      <Footer />
    </Layout>
  );
}
