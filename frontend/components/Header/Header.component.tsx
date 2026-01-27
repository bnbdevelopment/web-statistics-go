"use client";
import { Row, Col, Typography, Select, DatePicker } from "antd";
import { useSearchParams, useRouter } from "next/navigation";
import Image from "next/image";

const { RangePicker } = DatePicker;
const { Title } = Typography;

interface HeaderProps {
  sites: string[];
  selectedSite: string;
  onSiteChange: (site: string) => void;
  onDateChange: (dates: [Date | null, Date | null]) => void;
}

export default function Header({
  sites,
  selectedSite,
  onSiteChange,
  onDateChange,
}: HeaderProps) {
  const searchParams = useSearchParams();
  const router = useRouter();

  return (
    <header
      style={{
        background: "#fafafa",
        padding: "16px 24px",
        borderBottom: "1px solid #f0f0f0",
      }}
    >
      <Row justify="space-between" align="middle" gutter={[16, 16]}>
        <Col>
          <Row align="middle" gutter={16}>
            <Col>
              <Image src="/logo.png" alt="logo" width={32} height={32} />
            </Col>
            <Col>
              <Title level={4} style={{ margin: 0 }}>
                Statisztikák
              </Title>
            </Col>
          </Row>
        </Col>
        <Col xs={24} sm={24} md={12}>
          <Row gutter={[16, 16]} justify="end">
            <Col xs={24} sm={12}>
              <Select
                style={{ width: "100%" }}
                placeholder="Válassz webhelyet"
                value={selectedSite || ""}
                options={sites
                  .filter((site) => site !== "")
                  .map((site: string) => ({ value: site, label: site }))}
                onChange={(value) => {
                  onSiteChange(value);
                  const params = new URLSearchParams(searchParams.toString());
                  params.set("site", value);
                  router.replace(`?${params.toString()}`);
                }}
              />
            </Col>
            <Col xs={24} sm={12}>
              <RangePicker
                style={{ width: "100%" }}
                onChange={(dates: any) => {
                  onDateChange(
                    dates
                      ? [dates[0]?.["$d"] || null, dates[1]?.["$d"] || null]
                      : [null, null],
                  );
                }}
              />
            </Col>
          </Row>
        </Col>
      </Row>
    </header>
  );
}
