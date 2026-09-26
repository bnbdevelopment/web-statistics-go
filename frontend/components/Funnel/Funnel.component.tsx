"use client";
import { useEffect, useMemo, useState } from "react";
import { Card, Col, Progress, Row, Typography, Skeleton, Empty, Select, Space, Tag } from "antd";
import { FilterOutlined } from "@ant-design/icons";
import { formatLocalDate } from "@/utils/date";

interface FunnelStepStat {
  step: string;
  users: number;
  conversion: number;
  dropoff: number;
}

interface FunnelResponse {
  steps: FunnelStepStat[];
  firstStepSize: number;
  overallConversion: number;
}

interface Props {
  site: string;
  from: Date | null;
  to: Date | null;
}

const { Text } = Typography;

const Funnel = ({ site, from, to }: Props) => {
  const [funnel, setFunnel] = useState<FunnelResponse | null>(null);
  const [availablePages, setAvailablePages] = useState<string[]>([]);
  const [customSteps, setCustomSteps] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  // Fetch unique pages for the custom funnel step builder
  useEffect(() => {
    const fetchPages = async () => {
      try {
        const params = new URLSearchParams();
        if (site) params.set("site", site);
        if (from) params.set("from", formatLocalDate(from));
        if (to) params.set("to", formatLocalDate(to));

        const res = await fetch(`/api/v1/statistics/unique-pages?${params.toString()}`);
        if (res.ok) {
          const data = await res.json();
          setAvailablePages(data.pages || []);
        }
      } catch (err) {
        console.error("Failed to fetch unique pages for funnel:", err);
      }
    };
    fetchPages();
  }, [site, from, to]);

  // Fetch funnel statistics
  useEffect(() => {
    const fetchFunnel = async () => {
      setLoading(true);
      try {
        const params = new URLSearchParams();
        if (site) params.set("site", site);
        if (from) params.set("from", formatLocalDate(from));
        if (to) params.set("to", formatLocalDate(to));
        if (customSteps.length > 0) {
          params.set("steps", customSteps.join(","));
        }

        const res = await fetch(`/api/v1/statistics/funnel?${params.toString()}`);
        if (!res.ok) throw new Error("Funnel fetch failed");
        const data = (await res.json()) as FunnelResponse;
        setFunnel(data);
      } catch (err) {
        console.error("Funnel fetch error", err);
        setFunnel(null);
      } finally {
        setLoading(false);
      }
    };

    fetchFunnel();
  }, [site, from, to, customSteps]);

  const renderedSteps = useMemo(() => {
    if (!funnel || !funnel.steps.length) return null;
    return funnel.steps.map((step, index) => (
      <Card key={`${step.step}-${index}`} size="small" style={{ marginBottom: 12 }}>
        <Row align="middle" justify="space-between">
          <Col span={9}>
            <Text strong>{`${index + 1}. ${step.step}`}</Text>
            <div style={{ color: "#888", fontSize: 12 }}>
              {step.users} felhasználó {index > 0 ? `(-${step.dropoff.toFixed(1)}% lemorzsolódás)` : ""}
            </div>
          </Col>
          <Col span={15}>
            <Progress
              percent={step.conversion}
              strokeColor={index === funnel.steps.length - 1 ? "#52c41a" : "#1890ff"}
            />
          </Col>
        </Row>
      </Card>
    ));
  }, [funnel]);

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Space direction="vertical" style={{ width: "100%" }}>
          <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
            <FilterOutlined style={{ color: "#1890ff" }} />
            <Text type="secondary" style={{ fontSize: 12 }}>
              Egyedi tölcsér összeállítása (válaszd ki a lépéseket sorrendben):
            </Text>
          </div>
          <Select
            mode="tags"
            style={{ width: "100%" }}
            placeholder="Alapértelmezett tölcsér (vagy válassz ki aloldalakat...)"
            value={customSteps}
            onChange={(values) => setCustomSteps(values)}
            options={availablePages.map((p) => ({ value: p, label: p }))}
            allowClear
          />
        </Space>
      </div>

      {loading ? (
        <Skeleton active paragraph={{ rows: 4 }} />
      ) : !funnel || !funnel.steps.length ? (
        <Empty description="Nincs elég forgalom vagy egyező lépés a tölcsér megjelenítéséhez. Válassz konkrét oldalakat a fenti mezőben!" />
      ) : (
        <div>
          <Card
            size="small"
            style={{ marginBottom: 16, background: "linear-gradient(120deg,#fdfbfb,#ebedee)" }}
          >
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <div>
                <Text strong style={{ fontSize: 15 }}>
                  Összesített konverzió: {funnel.overallConversion}%
                </Text>
                <div style={{ color: "#595959", fontSize: 12 }}>
                  {funnel.firstStepSize} látogató lépett be az 1. lépésnél.
                </div>
              </div>
              <Tag color={funnel.overallConversion > 10 ? "green" : "blue"}>
                {funnel.steps.length} lépéses tölcsér
              </Tag>
            </div>
          </Card>
          {renderedSteps}
        </div>
      )}
    </div>
  );
};

export default Funnel;
