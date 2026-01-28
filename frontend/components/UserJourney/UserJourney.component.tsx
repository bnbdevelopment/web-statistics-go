"use client";
import { useState } from "react";
import { Input, Button, Timeline, Spin, Alert, Empty } from "antd";

interface JourneyEntry {
  page: string;
  timestamp: string;
}

const UserJourney = () => {
  const [sessionId, setSessionId] = useState("");
  const [journey, setJourney] = useState<JourneyEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchJourney = () => {
    if (!sessionId) {
      setError("Kérlek, adj meg egy munkamenet azonosítót.");
      return;
    }
    setLoading(true);
    setError(null);
    setJourney([]);

    fetch("/api/v1/user-journey", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ sessionId }),
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error("A munkamenet nem található vagy hiba történt.");
        }
        return response.json();
      })
      .then((data) => {
        setJourney(data || []);
      })
      .catch((err) => {
        setError(err.message);
      })
      .finally(() => {
        setLoading(false);
      });
  };

  return (
    <div>
      <Input.Search
        placeholder="Munkamenet azonosító (sessionId) megadása"
        enterButton="Útvonal mutatása"
        size="large"
        value={sessionId}
        onChange={(e) => setSessionId(e.target.value)}
        onSearch={fetchJourney}
        loading={loading}
      />
      {error && (
        <Alert
          message={error}
          type="error"
          showIcon
          style={{ marginTop: "16px" }}
        />
      )}
      <div style={{ marginTop: "24px", maxHeight: "400px", overflowY: "auto" }}>
        {loading ? (
          <div style={{ textAlign: "center" }}>
            <Spin />
          </div>
        ) : journey.length > 0 ? (
          <Timeline>
            {journey.map((entry, index) => (
              <Timeline.Item key={index}>
                <p>
                  <b>Oldal:</b> {entry.page}
                </p>
                <p>
                  <b>Időpont:</b>{" "}
                  {new Date(entry.timestamp).toLocaleString()}
                </p>
              </Timeline.Item>
            ))}
          </Timeline>
        ) : (
          !error && <Empty description="Nincs megjeleníthető útvonal. Végezz egy keresést." />
        )}
      </div>
    </div>
  );
};

export default UserJourney;
