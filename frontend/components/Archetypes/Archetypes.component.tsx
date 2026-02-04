"use client";
import { useEffect, useState } from "react";
import { Card, Row, Col, List, Tag, Spin, Alert, Empty } from "antd";
import { UserOutlined, ClockCircleOutlined, RetweetOutlined, AimOutlined } from "@ant-design/icons";

// --- Types ---
interface ArchetypeCharacteristic {
  name: string;
  value: string;
}

interface Archetype {
  name: string;
  percentage: number;
  characteristics: ArchetypeCharacteristic[];
  example_session_id: string;
}

interface ArchetypesProps {
  site: string;
  from: Date | null;
  to: Date | null;
}

const ArchetypeIcon = ({ name }: { name: string }) => {
    switch (name) {
        case "Célirányos Látogató":
            return <AimOutlined style={{ fontSize: 24, color: "#1890ff" }} />;
        case "Elmélyült Böngésző":
            return <ClockCircleOutlined style={{ fontSize: 24, color: "#52c41a" }} />;
        case "Frusztrált vagy Céltalan":
            return <RetweetOutlined style={{ fontSize: 24, color: "#f5222d" }} />;
        default:
            return <UserOutlined style={{ fontSize: 24, color: "#8c8c8c" }} />;
    }
}


const Archetypes = ({ site, from, to }: ArchetypesProps) => {
  const [archetypes, setArchetypes] = useState<Archetype[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    const fromDate = from ? `&from=${from.toISOString().split("T")[0]}` : "";
    const toDate = to ? `&to=${to.toISOString().split("T")[0]}` : "";
    const siteFilter = site ? `&site=${site}` : "";

    fetch(`/api/v1/statistics/archetypes?${siteFilter}${fromDate}${toDate}`)
      .then((response) => {
        if (!response.ok) {
          throw new Error("Network response was not ok");
        }
        return response.json();
      })
      .then((data: Archetype[]) => {
        setArchetypes(data || []);
        setError(null);
      })
      .catch((err) => {
        console.error("Error fetching archetypes:", err);
        setError("Failed to load archetype data.");
        setArchetypes([]);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [site, from, to]);

  if (loading) {
    return <div style={{display: 'flex', justifyContent: 'center', padding: '48px'}}><Spin size="large" /></div>;
  }

  if (error) {
    return <Alert message={error} type="error" />;
  }
  
  if (archetypes.length === 0) {
    return <Empty description="Nincs elég adat az archetípusok azonosításához." />;
  }

  return (
    <Row gutter={[24, 24]}>
      {archetypes.map((archetype) => (
        <Col xs={24} md={12} lg={6} key={archetype.name}>
          <Card
            title={
                <div style={{display: 'flex', alignItems: 'center', gap: '8px'}}>
                    <ArchetypeIcon name={archetype.name} />
                    <span style={{fontWeight: 'bold'}}>{archetype.name}</span>
                </div>
            }
            style={{ height: "100%" }}
            extra={<Tag color="blue">{archetype.percentage}%</Tag>}
          >
            <List
              dataSource={archetype.characteristics}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    title={item.name}
                    description={item.value}
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      ))}
    </Row>
  );
};

export default Archetypes;
