# Deployment Architecture

## MVP Deployment (Local Development)

```yaml
Deployment Model: Single Binary
Components:
    - Embedded SQLite database
    - Local container runtime (Docker)
    - File-based configuration
    - Local template storage

Installation:
    - Download single binary
    - Docker daemon requirement
    - Minimal configuration required
    - Automatic updates available
```

## Cloud Deployment Architecture

```yaml
Platform: Kubernetes
Services:
    - API Gateway (Ingress + Load Balancer)
    - Analysis Orchestrator (Deployment)
    - Worker Nodes (Job Queue)
    - Database (PostgreSQL)
    - Cache (Redis)
    - Object Storage (S3-compatible)

Scaling:
    - Horizontal Pod Autoscaler
    - Cluster Autoscaler for compute nodes
    - Database read replicas
    - CDN for static content delivery

Monitoring:
    - Prometheus for metrics collection
    - Grafana for visualization
    - ELK stack for log aggregation
    - Jaeger for distributed tracing
```

## Container Orchestration

```yaml
Analysis Containers:
    Resource Requests:
        memory: "1Gi"
        cpu: "500m"
    Resource Limits:
        memory: "2Gi"
        cpu: "2000m"

Security Context:
    runAsNonRoot: true
    runAsUser: 65534
    readOnlyRootFilesystem: true
    allowPrivilegeEscalation: false

Network Policy:
    - Deny all ingress
    - Allow egress to Git repositories only
    - Allow egress to vulnerability databases
```
