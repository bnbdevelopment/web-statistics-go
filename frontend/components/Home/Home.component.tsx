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
import Funnel from "../Funnel/Funnel.component";
import Engagement from "../Engagement/Engagement.component";
import Alerts from "../Alerts/Alerts.component";
import LandingPages from "../LandingPages/LandingPages.component";
import { formatLocalDate } from "@/utils/date";

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
    { interval: number; label?: string; timestamp?: string; uniqueSessions: number; totalRequests: number }[]
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
    fetch("/api/v1/get-sites")
      .then((response) => response.json())
      .then((data) => setSites(data.sites || []))
      .catch((error) => console.error("Error fetching sites:", error));
  }, []);

  useEffect(() => {
    const fetchActiveUsers = () => {
      fetch(`/api/v1/active?site=${encodeURIComponent(selectedSite)}`)
        .then((response) => response.json())
        .then((data) => setActiveUsers(data.count || 0))
        .catch((error) => console.error("Error fetching active users:", error));
    };

    fetchActiveUsers();
    const interval = setInterval(fetchActiveUsers, 30000);

    return () => clearInterval(interval);
  }, [selectedSite]);

  useEffect(() => {
    const fromStr = formatLocalDate(fromDate);
    const toStr = formatLocalDate(toDate);
    const from = fromStr ? `&from=${fromStr}` : "";
    const to = toStr ? `&to=${toStr}` : "";
    const siteParam = `site=${encodeURIComponent(selectedSite)}`;

    let fromYesterday = "";
    let toYesterday = "";

    if (fromDate && toDate) {
      const diff = toDate.getTime() - fromDate.getTime();
      const prevFromDate = new Date(fromDate.getTime() - diff);
      const prevToDate = fromDate;
      fromYesterday = `&from=${formatLocalDate(prevFromDate)}`;
      toYesterday = `&to=${formatLocalDate(prevToDate)}`;
    } else {
      const now = new Date();
      const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000);
      const twoDaysAgo = new Date(now.getTime() - 48 * 60 * 60 * 1000);
      fromYesterday = `&from=${formatLocalDate(twoDaysAgo)}`;
      toYesterday = `&to=${formatLocalDate(yesterday)}`;
    }

    // visitors current
    fetch(`/api/v1/traffic?${siteParam}${from}${to}`)
      .then((response) => response.json())
      .then((data) => setVisitors(data.traffic || 0))
      .catch((error) => console.error("Error fetching visitors:", error));

    // visitors previous period
    fetch(`/api/v1/traffic?${siteParam}${fromYesterday}${toYesterday}`)
      .then((response) => response.json())
      .then((data) => setVisitorsYesterday(data.traffic || 0))
      .catch((error) => console.error("Error fetching visitors yesterday:", error));

    // chart interval calculation:
    let numIntervals = 24;
    if (fromDate && toDate) {
      const days = Math.max(1, Math.ceil((toDate.getTime() - fromDate.getTime()) / (1000 * 60 * 60 * 24)));
      if (days <= 2) numIntervals = 24;
      else if (days <= 7) numIntervals = 28;
      else if (days <= 30) numIntervals = days;
      else numIntervals = 30;
    }

    fetch(`/api/v1/graph?${siteParam}&intervals=${numIntervals}${from}${to}`)
      .then((response) => response.json())
      .then((data) => setVisitToChart(data || []))
      .catch((error) => console.error("Error fetching graph data:", error));

    fetch(`/api/v1/time?${siteParam}${from}${to}`)
      .then((response) => response.json())
      .then((data) => setSpentTime(data.avgTimeSpent || 0))
      .catch((error) => console.error("Error fetching time:", error));

    fetch(`/api/v1/time?${siteParam}${fromYesterday}${toYesterday}`)
      .then((response) => response.json())
      .then((data) => setSpentTimeYesterday(data.avgTimeSpent || 0))
      .catch((error) => console.error("Error fetching time yesterday:", error));

    fetch(`/api/v1/sites?${siteParam}${from}${to}`)
      .then((response) => response.json())
      .then((data) => setSitesTraffic(data || []))
      .catch((error) => console.error("Error fetching site pages:", error));

    fetch(`/api/v1/bounce-rate?${siteParam}${from}${to}`)
      .then((response) => response.json())
      .then((data) => setBounceRate(data.bounceRate || 0))
      .catch((error) => console.error("Error fetching bounce rate:", error));

    fetch(`/api/v1/bounce-rate?${siteParam}${fromYesterday}${toYesterday}`)
      .then((response) => response.json())
      .then((data) => setBounceRateYesterday(data.bounceRate || 0))
      .catch((error) => console.error("Error fetching bounce rate yesterday:", error));
  }, [selectedSite, fromDate, toDate]);

  const calcDelta = (current: number, prev: number) => {
    if (prev <= 0) return null;
    return ((current - prev) / prev) * 100;
  };

  const visitorsDelta = calcDelta(visitors, visitorsYesterday);
  const timeDelta = calcDelta(spentTime, spentTimeYesterday);
  const bounceDelta = calcDelta(bounceRate, bounceRateYesterday);

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
            Pillanatkép {selectedSite ? `(${selectedSite})` : "(Összes webhely)"}
          </Title>
          <Row gutter={[24, 24]}>
            <Col xs={24} sm={12} lg={6}>
              <Card style={{ height: "100%" }}>
                <Statistic
                  title="Látogatók száma (fő)"
                  value={visitors}
                  valueStyle={{
                    color: visitors >= visitorsYesterday ? "#3f8600" : "#cf1322",
                  }}
                  prefix={
                    visitors >= visitorsYesterday ? (
                      <ArrowUpOutlined />
                    ) : (
                      <ArrowDownOutlined />
                    )
                  }
                />
                {visitorsDelta !== null && (
                  <div style={{ fontSize: 12, marginTop: 4, color: visitorsDelta >= 0 ? "#3f8600" : "#cf1322" }}>
                    {visitorsDelta >= 0 ? "+" : ""}{visitorsDelta.toFixed(1)}% előző időszakhoz képest
                  </div>
                )}
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
                      spentTime >= spentTimeYesterday ? "#3f8600" : "#cf1322",
                  }}
                  prefix={
                    spentTime >= spentTimeYesterday ? (
                      <ArrowUpOutlined />
                    ) : (
                      <ArrowDownOutlined />
                    )
                  }
                />
                {timeDelta !== null && (
                  <div style={{ fontSize: 12, marginTop: 4, color: timeDelta >= 0 ? "#3f8600" : "#cf1322" }}>
                    {timeDelta >= 0 ? "+" : ""}{timeDelta.toFixed(1)}% előző időszakhoz képest
                  </div>
                )}
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card style={{ height: "100%" }}>
                <Statistic
                  title="Visszafordulási arány (%)"
                  value={bounceRate}
                  precision={1}
                  valueStyle={{
                    color:
                      bounceRate <= bounceRateYesterday ? "#3f8600" : "#cf1322",
                  }}
                  prefix={
                    bounceRate <= bounceRateYesterday ? (
                      <ArrowDownOutlined />
                    ) : (
                      <ArrowUpOutlined />
                    )
                  }
                />
                {bounceDelta !== null && (
                  <div style={{ fontSize: 12, marginTop: 4, color: bounceDelta <= 0 ? "#3f8600" : "#cf1322" }}>
                    {bounceDelta >= 0 ? "+" : ""}{bounceDelta.toFixed(1)}% előző időszakhoz képest
                  </div>
                )}
              </Card>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <Card style={{ height: "100%" }}>
                <Statistic
                  title="Aktív látogatók (fő)"
                  value={activeUsers}
                  valueStyle={{ color: "#1890ff" }}
                />
                <div style={{ fontSize: 12, marginTop: 4, color: "#8c8c8c" }}>
                  Utolsó 5 perc forgalma
                </div>
              </Card>
            </Col>
          </Row>

          {/* --- Row 2: Main Visitors Chart --- */}
          <Row gutter={[24, 24]} style={{ marginTop: "24px" }}>
            <Col xs={24}>
              <Card title="Látogatói trend és kérések" style={{ height: "100%" }}>
                {visitToChart && visitToChart.length > 0 && (
                  <div style={{ width: "100%", height: 320 }}>
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
                              stopColor="#1890ff"
                              stopOpacity={0.8}
                            />
                            <stop
                              offset="95%"
                              stopColor="#1890ff"
                              stopOpacity={0}
                            />
                          </linearGradient>
                          <linearGradient
                            id="colorRequests"
                            x1="0"
                            y1="0"
                            x2="0"
                            y2="1"
                          >
                            <stop
                              offset="5%"
                              stopColor="#52c41a"
                              stopOpacity={0.6}
                            />
                            <stop
                              offset="95%"
                              stopColor="#52c41a"
                              stopOpacity={0}
                            />
                          </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="label" />
                        <YAxis />
                        <RechartsTooltip />
                        <Area
                          type="monotone"
                          dataKey="uniqueSessions"
                          name="Egyedi munkamenet"
                          stroke="#1890ff"
                          fill="url(#colorVisitors)"
                          strokeWidth={2}
                        />
                        <Area
                          type="monotone"
                          dataKey="totalRequests"
                          name="Összes kérés"
                          stroke="#52c41a"
                          fill="url(#colorRequests)"
                          strokeWidth={1.5}
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
              <Card title="Látogatók földrajzi eloszlása" style={{ height: "100%" }}>
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
                      <Bar dataKey="count" fill="#1890ff" name="Megtekintés" />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </Card>
            </Col>
          </Row>

          {/* --- Deeper Dive Section --- */}
          <Divider style={{ marginTop: "48px", marginBottom: "24px" }}>
            <Title level={4}>Mélyebb elemzések és intelligens betekintők</Title>
          </Divider>

          <Collapse ghost defaultActiveKey={["landing-exit", "funnel"]} accordion={false}>
            <Collapse.Panel
              key="landing-exit"
              header="Belépő és Kilépő oldalak elemzése (Landing & Exit Pages)"
            >
              <LandingPages site={selectedSite} from={fromDate} to={toDate} />
            </Collapse.Panel>

            <Collapse.Panel
              key="funnel"
              header="Konverziós tölcsér és automatikus riasztások"
            >
              <Row gutter={[24, 24]}>
                <Col xs={24} lg={14}>
                  <Card title="Konverziós tölcsér (Dinamikusan konfigurálható)">
                    <Funnel site={selectedSite} from={fromDate} to={toDate} />
                  </Card>
                </Col>
                <Col xs={24} lg={10}>
                  <Card title="Automatikus riasztások és UX anomáliák">
                    <Alerts site={selectedSite} from={fromDate} to={toDate} />
                  </Card>
                </Col>
              </Row>
            </Collapse.Panel>

            <Collapse.Panel
              key="engagement"
              header="Felhasználói elköteleződés és viselkedési szegmensek"
            >
              <Engagement site={selectedSite} from={fromDate} to={toDate} />
            </Collapse.Panel>

            <Collapse.Panel
              key="archetypes"
              header="Felhasználói Archetípusok (Automatikus viselkedés-elemzés)"
            >
              <Archetypes site={selectedSite} from={fromDate} to={toDate} />
            </Collapse.Panel>

            <Collapse.Panel
              key="time-analysis"
              header="Napszaki és heti bontású látogatottsági minták"
            >
              <TimeAnalysis site={selectedSite} from={fromDate} to={toDate} />
            </Collapse.Panel>

            <Collapse.Panel
              key="detailed-tables"
              header="Részletes forgalmi és földrajzi táblázatok"
            >
              <Row gutter={[24, 24]}>
                <Col xs={24} lg={12}>
                  <Card
                    title="Összes aloldal látogatottsága"
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
                          <Bar dataKey="count" fill="#8884d8" name="Megtekintés" />
                        </BarChart>
                      </ResponsiveContainer>
                    </div>
                  </Card>
                </Col>
                <Col xs={24} lg={12}>
                  <Card
                    title="Látogatók eloszlása város szerint"
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
                  Látogatók Visszatérése Hetente (Kohorsz Elemzés){" "}
                  <Tooltip title="Ez a nézet megmutatja, hogy az egy adott héten indult látogatók közül hányan tértek vissza a következő hetekben.">
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
              header="Átlagos felhasználói útvonal elemzés (Sankey / ReactFlow diagram)"
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
