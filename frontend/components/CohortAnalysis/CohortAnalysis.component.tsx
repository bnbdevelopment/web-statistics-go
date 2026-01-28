"use client";
import { useEffect, useState } from "react";
import {
  Table,
  Spin,
  Alert,
  Tooltip,
  Segmented,
  Empty,
} from "antd";
import type { TableProps } from "antd";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  ResponsiveContainer,
  Legend,
} from "recharts";

interface CohortData {
  cohort_date: string;
  total_users: number;
  retention_data: number[];
}

const CohortAnalysis = ({
  site,
  from,
  to,
}: {
  site?: string;
  from?: Date | null;
  to?: Date | null;
}) => {
  const [data, setData] = useState<CohortData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [view, setView] = useState<string | number>("Táblázat");
  const numberOfWeeks = 12;

  useEffect(() => {
    setLoading(true);
    const fromDate = from ? `&from=${from.toISOString().split("T")[0]}` : "";
    const toDate = to ? `&to=${to.toISOString().split("T")[0]}` : "";
    const siteFilter = site ? `&site=${site}` : "";

    fetch(
      `/api/v1/cohort?weeks=${numberOfWeeks}${siteFilter}${fromDate}${toDate}`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
      },
    )
      .then((response) => {
        if (!response.ok) {
          throw new Error("Network response was not ok");
        }
        return response.json();
      })
      .then((data) => {
        setData(data || []);
        setError(null);
      })
      .catch((err) => {
        console.error("Error fetching cohort data:", err);
        setError("Failed to load cohort data.");
        setData([]);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [site, from, to]);

  const getColorForPercentage = (percentage: number) => {
    if (percentage < 0) return "#ffffff";
    const opacity = percentage / 100;
    return `rgba(59, 130, 246, ${opacity})`; // Blue scale
  };

  const transformDataForChart = () => {
    const chartData: any[] = [];
    if (!data) return chartData;

    for (let i = 0; i < numberOfWeeks; i++) {
      const weekData: { [key: string]: any } = { name: `Hét ${i}` };
      data.forEach((cohort) => {
        weekData[cohort.cohort_date] = cohort.retention_data[i]?.toFixed(2) || 0;
      });
      chartData.push(weekData);
    }
    return chartData;
  };
  
  const COLORS = ["#8884d8", "#82ca9d", "#ffc658", "#ff7300", "#0088FE", "#00C49F", "#FFBB28", "#FF8042"];


  const renderTableView = () => {
    const columns: TableProps<CohortData>["columns"] = [
      {
        title: "Kohorsz",
        dataIndex: "cohort_date",
        key: "cohort_date",
        render: (text: string) => new Date(text).toLocaleDateString(),
        fixed: "left",
      },
      {
        title: "Felhasználók",
        dataIndex: "total_users",
        key: "total_users",
        fixed: "left",
      },
      {
        title: "Megtartás hetek szerint",
        children: Array.from({ length: numberOfWeeks }, (_, i) => ({
          title: `Hét ${i}`,
          dataIndex: "retention_data",
          key: `week_${i}`,
          render: (retention_data: number[], record: CohortData) => {
            const percentage = retention_data[i];
            const returningUsers = Math.round(
              (record.total_users * percentage) / 100,
            );
            const tooltipText = `A ${new Date(
              record.cohort_date,
            ).toLocaleDateString()} héten érkezett ${
              record.total_users
            } felhasználóból ${returningUsers} tért vissza a(z) ${i}. héten.`;

            return (
              <Tooltip title={tooltipText}>
                <div
                  style={{
                    backgroundColor: getColorForPercentage(percentage),
                    padding: "10px",
                    color: percentage > 50 ? "white" : "black",
                    textAlign: "center",
                    borderRadius: "4px",
                    cursor: "pointer",
                  }}
                >
                  {percentage.toFixed(2)}%
                </div>
              </Tooltip>
            );
          },
        })),
      },
    ];

    return (
      <Table
        dataSource={data}
        columns={columns}
        pagination={false}
        rowKey="cohort_date"
        bordered
        scroll={{ x: "max-content" }}
      />
    );
  };

  const renderChartView = () => {
    const chartData = transformDataForChart();

    return (
      <ResponsiveContainer width="100%" height={400}>
        <LineChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="name" />
          <YAxis
            label={{
              value: "Megtartás (%)",
              angle: -90,
              position: "insideLeft",
            }}
          />
          <Tooltip />
          <Legend />
          {data.map((cohort, index) => (
            <Line
              key={cohort.cohort_date}
              type="monotone"
              dataKey={cohort.cohort_date}
              stroke={COLORS[index % COLORS.length]}
              name={new Date(cohort.cohort_date).toLocaleDateString()}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    );
  };

  if (loading) {
    return <Spin size="large" />;
  }

  if (error) {
    return <Alert message={error} type="error" />;
  }

  return (
    <div>
      <Alert
        message="Hogyan kell értelmezni ezt az elemzést?"
        description={
          <div>
            <p>
              Ez az elemzés a felhasználói megtartást mutatja. A felhasználókat
              csoportokba (kohorszokba) soroljuk az első látogatásuk hete
              alapján.
            </p>
            <p>
              <b>Táblázat nézet:</b> Minden sor egy-egy kohorszot jelöl,
              melyek az első látogatásuk hete alapján csoportosított
              felhasználók. Az oszlopok a kohorsz hetétől eltelt heteket mutatják
              (pl. &apos;Hét 0&apos; az első látogatás hete, &apos;Hét 1&apos; az azt követő hét,
              stb.). A cellákban az látható, hogy az adott kohorszba tartozó
              felhasználók hány százaléka tért vissza az adott, első látogatásukat
              követő héten. A sötétebb szín magasabb megtartást jelez.
            </p>
            <p>
              <b>Grafikon nézet:</b> Minden vonal egy-egy kohorszot ábrázol. A
              grafikonon könnyen összehasonlítható, hogy a különböző
              időszakokban szerzett felhasználók megtartása hogyan alakult az
              idő múlásával. Egy magasabban futó vonal jobb felhasználói
              hűséget jelent.
            </p>
          </div>
        }
        type="info"
        showIcon
        style={{ marginBottom: "24px" }}
      />
      <Segmented
        options={["Táblázat", "Grafikon"]}
        value={view}
        onChange={setView}
        style={{ marginBottom: "24px" }}
      />
      {data && data.length > 0 ? (
        view === "Táblázat" ? renderTableView() : renderChartView()
      ) : (
        <Empty description="Nincs elegendő adat a kohorsz elemzéshez." />
      )}
    </div>
  );
};

export default CohortAnalysis;
