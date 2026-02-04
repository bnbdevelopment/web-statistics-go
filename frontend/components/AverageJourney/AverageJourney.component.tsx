"use client";
import React, { useEffect, useState, useMemo } from "react";
import ReactFlow, {
  Background,
  Controls,
  useNodesState,
  useEdgesState,
  MarkerType,
  Position,
  Edge,
  Node,
  NodeProps,
  Handle,
} from "reactflow";
import "reactflow/dist/style.css";
import dagre from "dagre";
import { Spin, Alert, Empty, Select } from "antd";
import { TeamOutlined } from "@ant-design/icons";

// --- TÍPUSOK ---
interface ApiNode {
  name: string;
}

interface ApiLink {
  source: number;
  target: number;
  value: number;
}

interface ApiData {
  nodes: ApiNode[];
  links: ApiLink[];
}

// --- EGYEDI "KÁRTYA" NODE ---
const AnalyticsNode = ({ data }: NodeProps) => {
  return (
    <div
      style={{
        padding: "8px 12px",
        background: "#fff",
        border: "1px solid #e0e0e0",
        borderLeft: `4px solid ${data.isStart ? "#52c41a" : "#1890ff"}`, // Zöld ha start, kék ha egyéb
        borderRadius: "6px",
        minWidth: "160px",
        boxShadow: "0 2px 5px rgba(0,0,0,0.08)",
        fontFamily: "sans-serif",
        fontSize: "12px",
      }}
    >
      <Handle
        type="target"
        position={Position.Left}
        style={{ visibility: "hidden" }}
      />

      <div style={{ fontWeight: 700, color: "#222", marginBottom: 4 }}>
        {data.label}
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: "5px",
          color: "#666",
        }}
      >
        <TeamOutlined />
        <span>~{data.visitors} látogatás</span>
      </div>

      <Handle
        type="source"
        position={Position.Right}
        style={{ visibility: "hidden" }}
      />
    </div>
  );
};

// --- LAYOUT SZÁMÍTÁS ---
const getLayoutedElements = (nodes: Node[], edges: Edge[]) => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  const nodeWidth = 180;
  const nodeHeight = 60;

  dagreGraph.setGraph({
    rankdir: "LR",
    ranksep: 120, // Távolság az oszlopok közt
    nodesep: 40, // Távolság a sorok közt
  });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  nodes.forEach((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    node.targetPosition = Position.Left;
    node.sourcePosition = Position.Right;
    node.position = {
      x: nodeWithPosition.x - nodeWidth / 2,
      y: nodeWithPosition.y - nodeHeight / 2,
    };
  });

  return { nodes, edges };
};

const AverageJourney = ({
  site,
  from,
  to,
}: {
  site?: string;
  from?: Date | null;
  to?: Date | null;
}) => {
  const nodeTypes = useMemo(() => ({ analyticsNode: AnalyticsNode }), []);

  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesState] = useEdgesState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filter states
  const [startPageFilter, setStartPageFilter] = useState<string>("");
  const [endPageFilter, setEndPageFilter] = useState<string>("");
  const [uniquePages, setUniquePages] = useState<string[]>([]);

  // Dátumok stringesítése
  const fromStr = from ? from.toISOString().split("T")[0] : "";
  const toStr = to ? to.toISOString().split("T")[0] : "";

  // Fetch unique pages for filter dropdowns
  useEffect(() => {
    let isMounted = true;
    const siteFilter = site ? `&site=${site}` : "";
    const fromParam = fromStr ? `&from=${fromStr}` : "";
    const toParam = toStr ? `&to=${toStr}` : "";

    fetch(`/api/v1/statistics/unique-pages?${siteFilter}${fromParam}${toParam}`)
      .then((res) => {
        if (!res.ok) throw new Error("API Hiba: Unique Pages");
        return res.json();
      })
      .then((data: { pages: string[] }) => {
        if (isMounted) {
          setUniquePages(data.pages || []);
        }
      })
      .catch((err) => {
        console.error("Error fetching unique pages:", err);
        if (isMounted) {
          setError("Hiba történt az oldalak listájának betöltésekor.");
        }
      });
      
    return () => {
      isMounted = false;
    };
  }, [site, fromStr, toStr]);

  useEffect(() => {
    let isMounted = true;
    setLoading(true);

    const siteFilter = site ? `&site=${site}` : "";
    const fromParam = fromStr ? `&from=${fromStr}` : "";
    const toParam = toStr ? `&to=${toStr}` : "";
    const startPageParam = startPageFilter ? `&start_page=${encodeURIComponent(startPageFilter)}` : "";
    const endPageParam = endPageFilter ? `&end_page=${encodeURIComponent(endPageFilter)}` : "";

    fetch(`/api/v1/average-journey?${siteFilter}${fromParam}${toParam}${startPageParam}${endPageParam}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    })
      .then((res) => {
        if (!res.ok) throw new Error("API Hiba");
        return res.json();
      })
      .then((data: ApiData) => {
        if (!isMounted) return;

        if (!data?.nodes?.length || !data?.links?.length) {
          setNodes([]);
          setEdges([]);
          setLoading(false);
          return;
        }

        // --- OKOS SZŰRÉS LOGIKA ---

        // 1. Átlagos forgalom kiszámítása a küszöbértékhez
        const allValues = data.links.map((l) => l.value);
        const maxVal = Math.max(...allValues);
        const avgVal = allValues.reduce((a, b) => a + b, 0) / allValues.length;

        // Küszöb: Vagy az átlag fele, vagy minimum 3 látogató.
        // Ez kiszűri a véletlenszerű 1-2 fős kattintásokat.
        const THRESHOLD = Math.max(3, avgVal * 0.4);

        // 2. Csak a releváns linkek megtartása
        const relevantLinks = data.links.filter(
          (link) => link.value >= THRESHOLD,
        );

        // 3. Mely node-ok érintettek a releváns linkekben? (Hogy ne legyenek árva dobozok)
        const activeNodeIndices = new Set<number>();
        relevantLinks.forEach((l) => {
          activeNodeIndices.add(l.source);
          activeNodeIndices.add(l.target);
        });

        // Látogatók számolása node-onként (csak a szűrt adatokból)
        const nodeVisits: Record<string, number> = {};
        relevantLinks.forEach((l) => {
          nodeVisits[l.target] = (nodeVisits[l.target] || 0) + l.value;
          // A source-ot is növeljük, ha ő a kiindulópont és még nincs értéke
          if (nodeVisits[l.source] === undefined) nodeVisits[l.source] = 0;
        });

        // 4. Node-ok létrehozása (csak az aktívakat)
        const flowNodes: Node[] = [];
        data.nodes.forEach((node, index) => {
          if (activeNodeIndices.has(index)) {
            // Ha a 0. index vagy a "/" a neve, jelöljük Startnak
            const isStart = index === 0 || node.name === "/";
            // Ha start, akkor a kimenő forgalmat becsüljük bejövőnek
            const displayVisits = isStart
              ? relevantLinks
                  .filter((l) => l.source === index)
                  .reduce((sum, l) => sum + l.value, 0)
              : nodeVisits[index] || 0;

            flowNodes.push({
              id: index.toString(),
              type: "analyticsNode",
              data: {
                label: node.name,
                visitors: displayVisits,
                isStart: isStart,
              },
              position: { x: 0, y: 0 },
            });
          }
        });

        // 5. Élek létrehozása
        const flowEdges: Edge[] = relevantLinks.map((link) => {
          const ratio = link.value / maxVal; // Arány a maximumhoz képest
          const width = Math.max(2, ratio * 8); // 2px és 10px közötti vastagság

          return {
            id: `e${link.source}-${link.target}`,
            source: link.source.toString(),
            target: link.target.toString(),
            type: "default",
            animated: true, // Minden fő útvonal mozogjon
            label: link.value.toString(),
            labelStyle: { fill: "#1890ff", fontWeight: 700 },
            labelBgStyle: { fill: "#fff", fillOpacity: 0.8 },
            style: {
              stroke: "#1890ff",
              strokeWidth: width,
              opacity: 0.8,
            },
            markerEnd: { type: MarkerType.ArrowClosed, color: "#1890ff" },
          };
        });

        const layouted = getLayoutedElements(flowNodes, flowEdges);
        setNodes(layouted.nodes);
        setEdges(layouted.edges);
        setLoading(false);
      })
      .catch((err) => {
        console.error(err);
        if (isMounted) {
          setError("Hiba történt.");
          setLoading(false);
        }
      });

    return () => {
      isMounted = false;
    };
  }, [site, fromStr, toStr, setNodes, setEdges]);

  if (loading)
    return (
      <div
        style={{
          height: 500,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        <Spin size="large" />
      </div>
    );
  if (error) return <Alert message={error} type="error" />;
  if (nodes.length === 0)
    return (
      <Empty description="Nincs elég adat a fő útvonalak kirajzolásához." />
    );

  return (
    <div
      style={{
        width: "100%",
        height: "600px",
        background: "#f5f7fa",
        borderRadius: "12px",
        border: "1px solid #d9d9d9",
      }}
    >
      <div
        style={{
          padding: "16px",
          display: "flex",
          gap: "16px",
          backgroundColor: "#fff",
          borderBottom: "1px solid #d9d9d9",
        }}
      >
        <Select
          showSearch
          placeholder="Kezdő oldal szűrése"
          optionFilterProp="children"
          onChange={(value) => setStartPageFilter(value)}
          value={startPageFilter}
          style={{ width: 200 }}
          options={[
            { value: "", label: "Összes kezdő oldal" },
            ...uniquePages.map((page) => ({ value: page, label: page })),
          ]}
        />
        <Select
          showSearch
          placeholder="Cél oldal szűrése"
          optionFilterProp="children"
          onChange={(value) => setEndPageFilter(value)}
          value={endPageFilter}
          style={{ width: 200 }}
          options={[
            { value: "", label: "Összes cél oldal" },
            ...uniquePages.map((page) => ({ value: page, label: page })),
          ]}
        />
      </div>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesState}
        fitView
        attributionPosition="bottom-right"
      >
        <Controls showInteractive={false} />
        <Background gap={20} size={1} color="#ccc" />
      </ReactFlow>
    </div>
  );
};

export default AverageJourney;
