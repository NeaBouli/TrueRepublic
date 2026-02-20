# Deployment Options

Overview of different deployment strategies for TrueRepublic nodes.

## Table of Contents

1. [Single Node Setup](#single-node-setup)
2. [Docker Compose](#docker-compose)
3. [Kubernetes](#kubernetes)
4. [Cloud Providers](#cloud-providers)
5. [Comparison Matrix](#comparison-matrix)

---

## Single Node Setup

**Best for:** Testing, development, small validators

### Architecture

```
Single Server
├── truerepublicd (blockchain)
├── Prometheus (metrics)
└── Grafana (dashboards)
```

### Deployment

See [Node Setup Guide](Node-Setup) for detailed instructions.

Quick start:

```bash
# Docker
make docker-build && make docker-up

# Native
make build && truerepublicd start
```

**Pros:**
- Simple setup
- Low cost
- Easy to manage

**Cons:**
- Single point of failure
- Limited scalability
- Manual failover

---

## Docker Compose

**Best for:** Small to medium deployments, easy updates

### Architecture

```
docker-compose.yml
├── truerepublic-node (blockchain)
├── prometheus (monitoring)
├── grafana (dashboards)
├── nginx (reverse proxy)
└── postgres (optional, for indexer)
```

### Full docker-compose.yml

```yaml
version: '3.8'

services:
  truerepublic-node:
    build: .
    container_name: truerepublic-node
    restart: unless-stopped
    ports:
      - "26656:26656"  # P2P
      - "26657:26657"  # RPC
      - "1317:1317"    # REST
      - "9090:9090"    # gRPC
    volumes:
      - ./data:/root/.truerepublic
      - ./config:/config
    environment:
      - MONIKER=${MONIKER}
      - CHAIN_ID=${CHAIN_ID}
    command: start
    networks:
      - truerepublic-net

  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    restart: unless-stopped
    ports:
      - "9091:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - truerepublic-net

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning
      - grafana-data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin}
      - GF_USERS_ALLOW_SIGN_UP=false
    networks:
      - truerepublic-net

  nginx:
    image: nginx:alpine
    container_name: nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
    depends_on:
      - truerepublic-node
    networks:
      - truerepublic-net

volumes:
  prometheus-data:
  grafana-data:

networks:
  truerepublic-net:
    driver: bridge
```

### Usage

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Update
docker-compose pull
docker-compose up -d

# Restart single service
docker-compose restart truerepublic-node
```

**Pros:**
- All services together
- Easy updates
- Reproducible
- Isolated environment

**Cons:**
- Still single machine
- Docker overhead
- Manual scaling

---

## Kubernetes

**Best for:** Large deployments, high availability, auto-scaling

### Architecture

```
Kubernetes Cluster
├── StatefulSet (truerepublic-node)
│   ├── Pod 1 (validator)
│   ├── Pod 2 (sentry)
│   └── Pod 3 (sentry)
├── Deployment (prometheus)
├── Deployment (grafana)
└── Service (load balancer)
```

### Kubernetes Manifests

**Namespace:**

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: truerepublic
```

**ConfigMap:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: truerepublic-config
  namespace: truerepublic
data:
  config.toml: |
    # CometBFT configuration
    [p2p]
    external_address = "tcp://0.0.0.0:26656"
    max_num_inbound_peers = 40
    max_num_outbound_peers = 10

    [consensus]
    timeout_commit = "5s"

    [rpc]
    laddr = "tcp://0.0.0.0:26657"

  app.toml: |
    # Application configuration
    [api]
    enable = true
    address = "tcp://0.0.0.0:1317"

    [grpc]
    enable = true
    address = "0.0.0.0:9090"
```

**StatefulSet:**

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: truerepublic-node
  namespace: truerepublic
spec:
  serviceName: truerepublic-node
  replicas: 3
  selector:
    matchLabels:
      app: truerepublic-node
  template:
    metadata:
      labels:
        app: truerepublic-node
    spec:
      containers:
      - name: node
        image: ghcr.io/neabouli/truerepublic:latest
        ports:
        - containerPort: 26656
          name: p2p
        - containerPort: 26657
          name: rpc
        - containerPort: 1317
          name: rest
        - containerPort: 9090
          name: grpc
        volumeMounts:
        - name: data
          mountPath: /root/.truerepublic
        - name: config
          mountPath: /config
        resources:
          requests:
            cpu: 2
            memory: 4Gi
          limits:
            cpu: 4
            memory: 8Gi
        livenessProbe:
          httpGet:
            path: /health
            port: 26657
          initialDelaySeconds: 60
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 26657
          initialDelaySeconds: 30
          periodSeconds: 10
      volumes:
      - name: config
        configMap:
          name: truerepublic-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: fast-ssd
      resources:
        requests:
          storage: 500Gi
```

**Service:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: truerepublic-node
  namespace: truerepublic
spec:
  type: LoadBalancer
  selector:
    app: truerepublic-node
  ports:
  - name: p2p
    port: 26656
    targetPort: 26656
  - name: rpc
    port: 26657
    targetPort: 26657
  - name: rest
    port: 1317
    targetPort: 1317
  - name: grpc
    port: 9090
    targetPort: 9090
```

### Deployment Commands

```bash
# Create namespace
kubectl apply -f namespace.yaml

# Deploy config
kubectl apply -f configmap.yaml

# Deploy StatefulSet
kubectl apply -f statefulset.yaml

# Deploy Service
kubectl apply -f service.yaml

# Check status
kubectl get pods -n truerepublic

# View logs
kubectl logs -f truerepublic-node-0 -n truerepublic

# Scale
kubectl scale statefulset truerepublic-node --replicas=5 -n truerepublic
```

**Pros:**
- High availability
- Auto-scaling
- Self-healing
- Rolling updates
- Resource management

**Cons:**
- Complex setup
- Learning curve
- Higher cost

---

## Cloud Providers

### AWS Deployment

**EC2 Instance:**

```
Instance Type: c6i.2xlarge
vCPUs: 8
RAM: 16 GB
Storage: 1 TB gp3 SSD
Cost: ~$300/month
```

**User Data Script:**

```bash
#!/bin/bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Clone repo
git clone https://github.com/NeaBouli/TrueRepublic.git /opt/truerepublic
cd /opt/truerepublic

# Configure
cp .env.example .env
sed -i "s/MONIKER=.*/MONIKER=aws-node-1/" .env
sed -i "s/EXTERNAL_IP=.*/EXTERNAL_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)/" .env

# Start
docker-compose up -d
```

### Google Cloud Platform

**Compute Engine:**

```
Machine Type: n2-standard-8
vCPUs: 8
RAM: 32 GB
Storage: 1 TB SSD
Cost: ~$350/month
```

**Deployment:**

```bash
# Create instance
gcloud compute instances create truerepublic-node-1 \
  --machine-type=n2-standard-8 \
  --boot-disk-size=1000GB \
  --boot-disk-type=pd-ssd \
  --image-family=ubuntu-2204-lts \
  --image-project=ubuntu-os-cloud \
  --metadata-from-file startup-script=startup.sh
```

### DigitalOcean

**Droplet:**

```
Plan: CPU-Optimized
Size: c-8 (8 vCPUs, 16 GB)
Storage: 1 TB NVMe
Cost: ~$240/month
```

**One-Click Deploy:**

```bash
# Using doctl CLI
doctl compute droplet create truerepublic-node \
  --size c-8 \
  --image ubuntu-22-04-x64 \
  --region nyc1 \
  --user-data-file cloud-init.yaml
```

### Hetzner

**Dedicated Server:**

```
Server: AX41-NVMe
CPU: AMD Ryzen 5 3600
RAM: 64 GB
Storage: 2x 512 GB NVMe RAID
Cost: ~EUR40/month (best value!)
```

---

## Comparison Matrix

| Feature | Single Node | Docker Compose | Kubernetes | Cloud |
|---------|-------------|----------------|------------|-------|
| **Setup Time** | 30 min | 1 hour | 4 hours | 1 hour |
| **Complexity** | Low | Low | High | Medium |
| **HA** | No | No | Yes | Optional |
| **Auto-Scaling** | No | No | Yes | Yes |
| **Cost** | $50-200/mo | $100-300/mo | $500+/mo | $200-500/mo |
| **Maintenance** | Manual | Semi-Auto | Auto | Semi-Auto |
| **Best For** | Dev/Test | Small Validators | Large Ops | Flexibility |

**Recommendations:**

- **Hobbyist:** Single Node (native or Docker)
- **Small Validator:** Docker Compose + monitoring
- **Professional Validator:** Kubernetes or multi-cloud
- **Enterprise:** Multi-region Kubernetes with DR

---

## Next Steps

- [Node Setup](Node-Setup) -- Detailed setup guide
- [Monitoring](Monitoring) -- Set up monitoring
- [Validator Guide](Validator-Guide) -- Become a validator
