# Deployment Guide

## Table of Contents

1. [Deployment Overview](#deployment-overview)
2. [Environment Configuration](#environment-configuration)
3. [Database Deployment](#database-deployment)
4. [Application Deployment](#application-deployment)
5. [Docker Deployment](#docker-deployment)
6. [Kubernetes Deployment](#kubernetes-deployment)
7. [Monitoring & Health Checks](#monitoring--health-checks)
8. [Troubleshooting](#troubleshooting)

---

## Deployment Overview

The UGCL platform can be deployed in multiple ways:
- **Docker Compose** - For development and small deployments
- **Kubernetes** - For production-scale deployments
- **Systemd** - For traditional Linux server deployments

### Prerequisites

- Go 1.25+
- PostgreSQL 15+
- Redis 7+
- MinIO (or S3-compatible storage)
- Docker (for containerized deployments)
- Kubernetes (for K8s deployments)

---

## Environment Configuration

See [development-guide.md](development-guide.md) for detailed environment setup.

**Production environment variables:**
```bash
# Server
SERVER_ENV=production
SERVER_PORT=8080

# Database
DB_HOST=prod-db.internal
DB_PORT=5432
DB_NAME=ugcl_prod
DB_USER=ugcl_prod_user
DB_PASSWORD=<from-secrets-manager>
DB_SSL_MODE=require

# Use secrets management for sensitive values
```

---

## Database Deployment

**Run migrations:**
```bash
atlas migrate apply --env production
```

**Verify:**
```bash
atlas schema inspect --env production
```

---

## Application Deployment

**Build:**
```bash
make build ENV=production
```

**Deploy:**
```bash
# Docker
docker-compose up -d

# Kubernetes
kubectl apply -f k8s/

# Systemd
sudo systemctl start ugcl
```

---

## Monitoring & Health Checks

**Health endpoints:**
- `GET /health` - Overall health
- `GET /ready` - Readiness probe
- `GET /metrics` - Prometheus metrics

---

## Troubleshooting

Check logs:
```bash
# Docker
docker-compose logs -f

# Kubernetes
kubectl logs -f deployment/ugcl-app

# Systemd
journalctl -u ugcl -f
```

For detailed deployment procedures, see individual platform documentation.
