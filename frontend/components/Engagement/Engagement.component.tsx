"use client";
import { useEffect, useState } from "react";
import { Card, Col, Progress, Row, Table, Tag, Typography, Skeleton, Empty } from "antd";
import { formatLocalDate } from "@/utils/date";

interface EngagementSegment {
  name: string;
  description: string;
  count: number;
  percentage: number;
}

interface EngagementLead {
  sessionId: string;
  score: number;
  duration: number;
  pages: number;
}

interface EngagementResponse {
  segments: EngagementSegment[];
  topSessions: EngagementLead[];
  averageScore: number;
}

interface Props {
  site: string;
  from: Date | null;
  to: Date | null;
}

const { Text } = Typography;

const columns = [
  { title: "Session", dataIndex: "sessionId", key: "sessionId" },
  { title: "Pontszám", dataIndex: "score", key: "score", render: (val: number) => `${val.toFixed(1)}` },
  { title: "Időtartam (mp)", dataIndex: "duration", key: "duration", render: (val: number) => Math.round(val) },
  { title: "Oldalak", dataIndex: "pages", key: "pages" },
];

const Engagement = ({ site, from, to }: Props) => {
  const [data, setData] = useState<EngagementResponse | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const params = new URLSearchParams();
        if (site) params.set("site", site);
        if (from) params.set("from", formatLocalDate(from));
        if (to) params.set("to", formatLocalDate(to));

        const res = await fetch(`/api/v1/statistics/engagement?${params.toString()}`);
        if (!res.ok) throw new Error("Engagement fetch failed");
        const payload = (await res.json()) as EngagementResponse;
        setData(payload);
      } catch (err) {
        console.error("Engagement fetch error", err);
        setData(null);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [site, from, to]);

  if (loading) {
    return <Skeleton active paragraph={{ rows: 6 }} />;
  }

  if (!data) {
    return <Empty description="Nincs elköteleződési adat" />;
  }

  return (
    <Row gutter={[24, 24]}>
      <Col xs={24} lg={10}>
        <Card title={`Átlagos elköteleződés: ${data.averageScore.toFixed(1)}`}>
          {data.segments.map((segment) => (
            <div key={segment.name} style={{ marginBottom: 16 }}>
              <div style={{ display: "flex", justifyContent: "space-between" }}>
                <Text strong>{segment.name}</Text>
                <Tag>
                  {segment.count} / {segment.percentage.toFixed(1)}%
                </Tag>
              </div>
              <Text type="secondary" style={{ display: "block", marginBottom: 4 }}>
                {segment.description}
              </Text>
              <Progress percent={segment.percentage} showInfo={false} />
            </div>
          ))}
        </Card>
      </Col>
      <Col xs={24} lg={14}>
        <Card title="Top munkamenetek">
          <Table columns={columns} dataSource={data.topSessions} pagination={false} rowKey="sessionId" size="small" />
        </Card>
      </Col>
    </Row>
  );
};

export default Engagement;
