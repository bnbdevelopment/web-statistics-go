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

const CohortSummary = ({ data }: { data: CohortData[] }) => {
  if (!data || data.length < 2) {
    return null;
  }

  const keyWeek = 2; // Analyze performance at Week 2
  let totalRetention = 0;
  let validCohorts = 0;
  let bestCohort = { date: "", retention: -1 };
  
  data.forEach(cohort => {
    if (cohort.retention_data.length > keyWeek) {
      const retention = cohort.retention_data[keyWeek];
      totalRetention += retention;
      validCohorts++;
      if (retention > bestCohort.retention) {
        bestCohort = { date: cohort.cohort_date, retention: retention };
      }
    }
  });

  if (validCohorts === 0) return null;

  const averageRetention = totalRetention / validCohorts;
  const mostRecentCohort = data[0];
  const recentRetention = mostRecentCohort.retention_data.length > keyWeek ? mostRecentCohort.retention_data[keyWeek] : -1;
  
  const trend = recentRetention > averageRetention ? "jobb" : "alacsonyabb";
  const trendColor = recentRetention > averageRetention ? "green" : "orange";

  const insights = [];

  if (recentRetention >= 0) {
    insights.push(
      <li key="1">
        A legfrissebb ({new Date(mostRecentCohort.cohort_date).toLocaleDateString()}) csoport megtartása a {keyWeek}. héten{" "}
        <b style={{ color: trendColor }}>{recentRetention.toFixed(1)}%</b>, ami {trend} az átlagos {averageRetention.toFixed(1)}%-nál.
      </li>
    );
  }

  if (bestCohort.retention >= 0) {
    insights.push(
      <li key="2">
        A legjobban teljesítő csoport a <b style={{color: "green"}}>{new Date(bestCohort.date).toLocaleDateString()}-i</b> volt, {bestCohort.retention.toFixed(1)}%-os megtartással a {keyWeek}. héten.
      </li>
    );
  }

  return (
    <Alert
      message="Gyors Elemzés"
      description={<ul>{insights}</ul>}
      type="success"
      showIcon
      style={{ marginBottom: "24px" }}
    />
  );
};


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
    if (percentage < 0) return "#ffffff"; // Should not happen
    if (percentage > 50) return "rgba(34, 197, 94, 0.7)"; // Strong Green
    if (percentage > 25) return "rgba(163, 230, 53, 0.6)"; // Light Green
    if (percentage > 10) return "rgba(251, 191, 36, 0.5)"; // Light Orange/Yellow
    return "rgba(252, 165, 165, 0.4)"; // Light Red
  };

  const calculateAverageRetention = () => {
    if (!data || data.length === 0) return [];
    
    const weeklyAverages = Array(numberOfWeeks).fill(0);
    const cohortCountsPerWeek = Array(numberOfWeeks).fill(0);

    data.forEach(cohort => {
      cohort.retention_data.forEach((percentage, weekIndex) => {
        if (percentage >= 0) {
          weeklyAverages[weekIndex] += percentage;
          cohortCountsPerWeek[weekIndex]++;
        }
      });
    });

    return weeklyAverages.map((total, index) => ({
      name: `Hét ${index}`,
      "Átlagos megtartás": cohortCountsPerWeek[index] > 0 ? (total / cohortCountsPerWeek[index]).toFixed(2) : 0,
    }));
  };

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
                    color: "black",
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
    const chartData = calculateAverageRetention();

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
            domain={[0, 100]}
          />
          <Tooltip />
          <Legend />
          <Line
            key="average_retention"
            type="monotone"
            dataKey="Átlagos megtartás"
            stroke="#8884d8"
            strokeWidth={3}
          />
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
      <CohortSummary data={data} />
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
