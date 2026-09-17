"use client";
import React, { useEffect, useState } from "react";
import { Card, Col, Row, Table, Tag, Typography, Progress, Spin, Alert, Button, Empty } from "antd";
import type { TableProps } from "antd";
import { DownloadOutlined, LoginOutlined, LogoutOutlined } from "@ant-design/icons";
import { formatLocalDate } from "@/utils/date";

const { Text } = Typography;

interface LandingPageStat {
  page: string;
  sessions: number;
  bounces: number;
  bounceRate: number;
}

interface ExitPageStat {
  page: string;
  exits: number;
  exitRate: number;
}

interface Props {
  site: string;
  from: Date | null;
  to: Date | null;
}

const landingColumns: TableProps<LandingPageStat>["columns"] = [
  {
    title: "Belépési oldal (Landing)",
    dataIndex: "page",
    key: "page",
    render: (text: string) => <Text strong>{text || "/"}</Text>,
  },
  {
    title: "Munkamenetek",
    dataIndex: "sessions",
    key: "sessions",
    sorter: (a, b) => a.sessions - b.sessions,
  },
  {
    title: "Visszafordult",
    dataIndex: "bounces",
    key: "bounces",
    sorter: (a, b) => a.bounces - b.bounces,
  },
  {
    title: "Bounce Rate",
    dataIndex: "bounceRate",
    key: "bounceRate",
    sorter: (a, b) => a.bounceRate - b.bounceRate,
    render: (rate: number) => {
      let color = "#3f8600";
      if (rate > 60) color = "#cf1322";
      else if (rate > 35) color = "#faad14";
      return (
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <Progress
            percent={rate}
            size="small"
            strokeColor={color}
            style={{ width: 80, margin: 0 }}
            showInfo={false}
          />
          <span>{rate.toFixed(1)}%</span>
        </div>
      );
    },
  },
];

const exitColumns: TableProps<ExitPageStat>["columns"] = [
  {
    title: "Kilépési oldal (Exit)",
    dataIndex: "page",
    key: "page",
    render: (text: string) => <Text strong>{text || "/"}</Text>,
  },
  {
    title: "Kilépések száma",
    dataIndex: "exits",
    key: "exits",
    sorter: (a, b) => a.exits - b.exits,
  },
  {
    title: "Kilépési arány",
    dataIndex: "exitRate",
    key: "exitRate",
    sorter: (a, b) => a.exitRate - b.exitRate,
    render: (rate: number) => <Tag color="orange">{rate.toFixed(1)}%</Tag>,
  },
];

export default function LandingPages({ site, from, to }: Props) {
  const [landingPages, setLandingPages] = useState<LandingPageStat[]>([]);
  const [exitPages, setExitPages] = useState<ExitPageStat[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      setError(null);
      try {
        const params = new URLSearchParams();
        if (site) params.set("site", site);
        if (from) params.set("from", formatLocalDate(from));
        if (to) params.set("to", formatLocalDate(to));

        const [landingRes, exitRes] = await Promise.all([
          fetch(`/api/v1/statistics/landing-pages?${params.toString()}`),
          fetch(`/api/v1/statistics/exit-pages?${params.toString()}`),
        ]);

        if (!landingRes.ok || !exitRes.ok) {
          throw new Error("Hiba a landing/exit oldalak lekérésekor");
        }

        const landingData = await landingRes.json();
        const exitData = await exitRes.json();

        setLandingPages(landingData || []);
        setExitPages(exitData || []);
      } catch (err) {
        console.error("Landing/Exit pages fetch error:", err);
        setError("Nem sikerült betölteni az adatokat.");
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [site, from, to]);

  const exportCSV = (data: Record<string, unknown>[], filename: string) => {
    if (!data.length) return;
    const headers = Object.keys(data[0]).join(",");
    const rows = data.map((item) => Object.values(item).join(",")).join("\n");
    const blob = new Blob([`${headers}\n${rows}`], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", `${filename}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (loading) {
    return (
      <div style={{ display: "flex", justifyContent: "center", padding: "48px" }}>
        <Spin size="large" />
      </div>
    );
  }

  if (error) {
    return <Alert message={error} type="error" showIcon />;
  }

  return (
    <Row gutter={[24, 24]}>
      <Col xs={24} lg={12}>
        <Card
          title={
            <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <LoginOutlined style={{ color: "#1890ff" }} />
              <span>Leggyakoribb Belépő Oldalak & Visszafordulás</span>
            </div>
          }
          extra={
            <Button
              size="small"
              icon={<DownloadOutlined />}
              onClick={() => exportCSV(landingPages, "landing_pages")}
            >
              CSV
            </Button>
          }
        >
          {landingPages.length === 0 ? (
            <Empty description="Nincs elegendő adat" />
          ) : (
            <Table
              columns={landingColumns}
              dataSource={landingPages}
              rowKey="page"
              pagination={{ pageSize: 5 }}
              size="small"
            />
          )}
        </Card>
      </Col>

      <Col xs={24} lg={12}>
        <Card
          title={
            <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <LogoutOutlined style={{ color: "#faad14" }} />
              <span>Leggyakoribb Kilépő Oldalak (Lemondási Pontok)</span>
            </div>
          }
          extra={
            <Button
              size="small"
              icon={<DownloadOutlined />}
              onClick={() => exportCSV(exitPages, "exit_pages")}
            >
              CSV
            </Button>
          }
        >
          {exitPages.length === 0 ? (
            <Empty description="Nincs elegendő adat" />
          ) : (
            <Table
              columns={exitColumns}
              dataSource={exitPages}
              rowKey="page"
              pagination={{ pageSize: 5 }}
              size="small"
            />
          )}
        </Card>
      </Col>
    </Row>
  );
}
