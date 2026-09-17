# Web Statistics Helm chart

This chart installs the Go API and Next.js dashboard. It does not install PostgreSQL, TimescaleDB, an ingress controller, Prometheus Operator, Traefik, or an OpenTelemetry backend.

## Prerequisites

- Kubernetes 1.25 or newer
- An external PostgreSQL or TimescaleDB database
- Access to the configured Artifact Registry images
- Helm 3.8 or newer for OCI registries

The `ServiceMonitor` and `IngressRoute` options require their corresponding CRDs to exist in the cluster. The chart never installs third party CRDs.

## Install

Create the database secret first. The key names can be changed through `database.secretKeys`.

```bash
kubectl create namespace web-statistics
kubectl -n web-statistics create secret generic web-statistics-database \
  --from-literal=host=postgres.example.internal \
  --from-literal=port=5432 \
  --from-literal=name=statistics \
  --from-literal=username=web_statistics \
  --from-literal=password='replace-me' \
  --from-literal=sslmode=require
```

Create a small values file:

```yaml
database:
  existingSecret: web-statistics-database

imagePullSecrets:
  - name: artifact-registry

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: analytics.example.com
      paths:
        - path: /api/v1
          pathType: Prefix
          service: backend
        - path: /
          pathType: Prefix
          service: frontend
  tls:
    - secretName: analytics-tls
      hosts: [analytics.example.com]
```

Then install from Artifact Registry:

```bash
helm registry login europe-west3-docker.pkg.dev
helm install web-statistics \
  oci://europe-west3-docker.pkg.dev/bnbdevelopment/webstats/web-statistics \
  --namespace web-statistics \
  --values production.yaml
```

For a local evaluation, the chart can create a database Secret from values by setting `database.createSecret=true`. Helm stores those values in release state, so use `database.existingSecret` for production credentials.

## Required and commonly changed values

| Value | Default | Purpose |
| --- | --- | --- |
| `database.existingSecret` | `""` | Secret containing every database connection field. If empty, the expected name is `<release>-web-statistics-database`. |
| `database.secretKeys.*` | See `values.yaml` | Maps environment variables to keys in the database Secret. |
| `backend.image.*` | Artifact Registry backend | Backend image location and immutable SemVer tag. |
| `frontend.image.*` | Artifact Registry frontend | Frontend image location and immutable SemVer tag. |
| `imagePullSecrets` | `[]` | Registry credentials for both workloads. |
| `backend.env.prefix` | `/api/v1` | API base path. The bundled frontend currently calls this path. |
| `ingress.enabled` | `false` | Creates a standard Kubernetes `Ingress`. |
| `traefik.ingressRoute.enabled` | `false` | Creates a Traefik `IngressRoute`; requires the Traefik CRD. |
| `serviceMonitor.enabled` | `false` | Scrapes backend metrics from `/metrics`; requires Prometheus Operator. |
| `backend.migration.enabled` | `true` | Runs `main migrate` in an init container before each backend pod starts. |
| `observability.otelCollector.enabled` | `false` | Adds an OpenTelemetry Collector sidecar that exports backend log files over OTLP/HTTP. |
| `networkPolicy.enabled` | `false` | Applies the supplied ingress and egress rules to chart pods. |

See [values.yaml](values.yaml) for probes, resources, autoscaling, disruption budgets, scheduling, extra environment variables, extra volumes, and arbitrary sidecars.

## Database migrations

The backend image supports a dedicated `migrate` command. With migrations enabled, an init container connects using the same Secret as the application and applies GORM schema changes before the backend starts. PostgreSQL must be reachable from the pod, and the database user must have schema migration privileges.

Rolling updates can start more than one migration init container. GORM migrations used by this application are idempotent, but production database permissions and lock timeouts should still be chosen with concurrent startup in mind.

## Metrics and logs

The backend exposes Prometheus metrics at `/metrics`. Enable `serviceMonitor.enabled` when Prometheus Operator is installed, or scrape the backend Service directly.

Application logs always go to container stdout. When `observability.otelCollector.enabled=true`, the backend also writes to a shared volume and an OpenTelemetry Collector sidecar tails that file. Set `observability.otelCollector.endpoint` to an OTLP/HTTP receiver. For authenticated exporters, put the required environment variables in a Secret, reference it with `observability.otelCollector.existingSecret`, and set `observability.otelCollector.config` to an exporter configuration that reads those variables.

You can replace the complete Collector configuration with `observability.otelCollector.config`. Both workloads also accept arbitrary `sidecars`, `extraVolumes`, and `extraVolumeMounts`.

## GeoIP database

GeoIP is optional. Enable it and mount an existing PVC:

```yaml
geoip:
  enabled: true
  existingClaim: geoip-database
  path: /geodb/GeoDB.mmdb
```

You can instead supply any native volume source under `geoip.volume`, for example a ConfigMap, Secret, or CSI volume.

## Native CRD resources

Standard Ingress and Traefik IngressRoute are separate switches. Enable only the routing resource used in the cluster. Likewise, ServiceMonitor is opt in. Helm rendering does not query the cluster for CRDs, so an enabled custom resource requires its CRD to be installed before this chart.

## Security defaults

Containers run as UID/GID 1001, drop Linux capabilities, disallow privilege escalation, use the runtime default seccomp profile, and use read only root filesystems. Service account token mounting is disabled. Pod disruption budgets and resource requests are enabled by default. NetworkPolicy stays disabled until explicit ingress, DNS, database, telemetry, and internet egress rules are supplied.
