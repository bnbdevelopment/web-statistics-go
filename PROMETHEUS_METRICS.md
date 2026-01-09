# Prometheus Metrics Documentation

This document describes all Prometheus metrics exported by the Web Statistics application at the `/metrics` endpoint.

**Update Frequency**: All metrics are updated every 2 seconds via a background goroutine.

## Standard Traffic Metrics

### `statistics_traffic`
**Type**: Gauge
**Description**: Unique sessions (traffic) in last 24 hours by site
**Labels**:
- `site` - The domain being tracked

**Example**:
```promql
statistics_traffic{site="example.com"}
```

**Use Case**: Track total unique visitors to each site over the past 24 hours.

---

### `statistics_active_users`
**Type**: Gauge
**Description**: Number of currently active users by site (last 5 minutes)
**Labels**:
- `site` - The domain being tracked

**Example**:
```promql
statistics_active_users{site="example.com"}
```

**Use Case**: Real-time monitoring of active users currently browsing the site.

---

### `statistics_minute_spent`
**Type**: Gauge
**Description**: Average minutes spent on site in last 24h by site
**Labels**:
- `site` - The domain being tracked

**Example**:
```promql
statistics_minute_spent{site="example.com"}
```

**Use Case**: Track user engagement by measuring average session duration.

---

## Geolocation Metrics

### `statistics_traffic_by_city`
**Type**: Gauge
**Description**: Unique sessions in last 24 hours by city and country
**Labels**:
- `site` - The domain being tracked
- `city` - City name (or "Unknown" if not available)
- `country_code` - ISO 3166-1 alpha-2 country code (e.g., "US", "GB", or "XX" for unknown)
- `country_name` - Full country name (or "Unknown" if not available)

**Example**:
```promql
statistics_traffic_by_city{site="example.com", country_code="US"}
statistics_traffic_by_city{site="example.com", city="New York", country_code="US", country_name="United States"}
```

**Use Cases**:
- Breakdown traffic by city
- Identify top cities for each country
- Create city-level traffic dashboards

**Grafana Query Examples**:
```promql
# Top 10 cities by traffic
topk(10, statistics_traffic_by_city{site="example.com"})

# Traffic from specific country
sum by (city) (statistics_traffic_by_city{site="example.com", country_code="US"})
```

---

### `statistics_active_users_by_country`
**Type**: Gauge
**Description**: Currently active users by country (last 5 minutes)
**Labels**:
- `site` - The domain being tracked
- `country_code` - ISO 3166-1 alpha-2 country code (e.g., "US", "GB", or "XX" for unknown)
- `country_name` - Full country name (or "Unknown" if not available)

**Example**:
```promql
statistics_active_users_by_country{site="example.com", country_code="US", country_name="United States"}
```

**Use Cases**:
- Real-time geographic distribution of active users
- Monitor peak hours by region
- Identify growing markets

**Grafana Query Examples**:
```promql
# Top 5 countries with active users
topk(5, statistics_active_users_by_country{site="example.com"})

# Total active users across all countries
sum(statistics_active_users_by_country{site="example.com"})
```

---

### `statistics_traffic_coordinates`
**Type**: Gauge
**Description**: Traffic by geographic coordinates for geomap visualization
**Labels**:
- `site` - The domain being tracked
- `geohash` - Geohash encoded location (precision 7, ~152m accuracy, e.g., "9q8yyk8")
- `latitude` - Latitude coordinate (formatted to 4 decimal places, e.g., "37.7749")
- `longitude` - Longitude coordinate (formatted to 4 decimal places, e.g., "-122.4194")
- `city` - City name (or "Unknown" if not available)
- `country_code` - ISO 3166-1 alpha-2 country code (e.g., "US", or "XX" for unknown)

**Example**:
```promql
statistics_traffic_coordinates{site="example.com", geohash="9q8yyk8", latitude="37.7749", longitude="-122.4194", city="San Francisco", country_code="US"}
```

**Use Cases**:
- **Grafana Geomap Panel**: Visualize traffic on a world map using geohash
- Plot user locations geographically
- Identify geographic clusters of users
- Efficient spatial queries using geohash prefixes

**Grafana Geomap Setup**:
1. Add Geomap panel
2. Use query: `statistics_traffic_coordinates{site="example.com"}`
3. Configure panel to use `geohash` label (recommended) or `latitude`/`longitude` labels
4. Size markers by metric value (traffic count)

**Example Queries**:
```promql
# All traffic with coordinates
statistics_traffic_coordinates{site="example.com"}

# Traffic from specific region using geohash prefix (San Francisco Bay Area)
statistics_traffic_coordinates{site="example.com", geohash=~"9q8.*"}

# Traffic from specific region using lat/long (approximate)
statistics_traffic_coordinates{site="example.com", latitude=~"37.*", longitude=~"-122.*"}

# Filter by geohash for efficient spatial queries
sum by (geohash) (statistics_traffic_coordinates{site="example.com"})
```

**Geohash Benefits**:
- **Efficient spatial queries**: Filter by geohash prefix to query regions (e.g., `geohash=~"9q.*"` for California)
- **Reduced cardinality**: Geohash provides natural clustering of nearby locations
- **Better Grafana performance**: Single label instead of parsing lat/long pairs
- **Hierarchical zoom**: Different precision levels for different zoom levels

---

## Common Query Patterns

### Total Traffic
```promql
sum(statistics_traffic)
```

### Traffic per Site
```promql
statistics_traffic{site=~".*"}
```

### Active Users Comparison
```promql
sum by (site) (statistics_active_users)
```

### Geographic Distribution (Top 10 Countries)
```promql
topk(10, sum by (country_name) (statistics_traffic_by_city))
```

### Average Engagement Time Across All Sites
```promql
avg(statistics_minute_spent)
```

### Real-time Active Users by Country
```promql
sort_desc(statistics_active_users_by_country)
```

---

## Notes

- **Unknown Values**: When geolocation data is unavailable (failed lookup, invalid IP, etc.), the following defaults are used:
  - `city`: "Unknown"
  - `country_code`: "XX"
  - `country_name`: "Unknown"

- **Coordinate Precision**: Latitude and longitude are formatted to 4 decimal places (~11 meter precision) to reduce metric cardinality while maintaining useful geographic granularity.

- **Geohash Precision**: Geohashes are encoded with precision 7, providing ~152 meter accuracy. This balances location accuracy with metric cardinality. Geohash prefix matching enables efficient regional queries (e.g., all of California with `geohash=~"9.*"`).

- **Time Windows**:
  - Last 24 hours: `statistics_traffic`, `statistics_traffic_by_city`, `statistics_traffic_coordinates`
  - Last 5 minutes: `statistics_active_users`, `statistics_active_users_by_country`

- **Metric Cardinality**: Geolocation metrics may create thousands of time series depending on the number of unique cities/coordinates. Monitor Prometheus memory usage if tracking many high-traffic sites.
