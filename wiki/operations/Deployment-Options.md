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

The authoritative stack is the repository's
[`docker-compose.yml`](https://github.com/NeaBouli/TrueRepublic/blob/main/docker-compose.yml);
do not copy an older inline version. It defines `truerepublic-node`,
`client-web`, `nginx`, `prometheus` and `grafana`, with every published port
bound to the host loopback:

| Host binding | Service |
|---|---|
| `127.0.0.1:${P2P_PORT:-26656}` | node P2P |
| `127.0.0.1:8080` | nginx (`/rpc/`, `/api/` reverse proxy) |
| `127.0.0.1:9091` | Prometheus |
| `127.0.0.1:3001` | client-web |
| `127.0.0.1:3000` | Grafana (`GRAFANA_PASSWORD` required, no default) |

RPC and REST listen only inside the node network namespace and gRPC is disabled.
The node runs as the non-root `truerepublic` user with its home in the
`node-data` volume at `/home/truerepublic/.truerepublic`.

### Usage

```bash
# Start all services
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down

# Update: the node image is built locally from source (no published image);
# pull the reviewed source revision, then rebuild
git pull
make docker-build
docker compose up -d

# Restart single service
docker compose restart truerepublic-node
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
    laddr = "tcp://PUBLIC_ROLE_INTERFACE:26656"
    external_address = "tcp://PUBLIC_ROLE_ADDRESS:26656"
    max_num_inbound_peers = 40
    max_num_outbound_peers = 10

    [consensus]
    timeout_commit = "5s"

    [rpc]
    laddr = "tcp://127.0.0.1:26657"
    unsafe = false

  app.toml: |
    # Application configuration
    [api]
    enable = true
    address = "tcp://127.0.0.1:1317"
    enabled-unsafe-cors = false

    [grpc]
    enable = true
    address = "127.0.0.1:9090"
```

These loopback listeners require a reviewed same-pod proxy sidecar. Validate
the rendered node home against the
[role-based network policy](../../docs/node-operators/configuration/network-policy.md);
never publish the upstream ports directly.

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
        # No image is published; build from source and push to your own
        # registry under a reviewed, immutable tag
        image: <your-registry>/truerepublic:<reviewed-tag>
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
          mountPath: /home/truerepublic/.truerepublic
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
          exec:
            command:
            - /usr/local/bin/truerepublicd
            - healthcheck
            - live
            - --timeout
            - 2s
          initialDelaySeconds: 60
          periodSeconds: 30
        readinessProbe:
          exec:
            command:
            - /usr/local/bin/truerepublicd
            - healthcheck
            - ready
            - --timeout
            - 2s
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

**Service:** expose only the roles your validated
[network policy](../../docs/node-operators/configuration/network-policy.md)
allows. A validator publishes no public port; a seed/sentry/RPC role publishes
P2P, and RPC only through the reviewed sidecar proxy. Never publish 26657, 1317
or 9090 directly (see above). The repository ships no Kubernetes manifests or
ingress; the snippets here are illustrations you must review for your cluster.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: truerepublic-node-p2p
  namespace: truerepublic
spec:
  type: ClusterIP   # publish externally only for a qualified seed/sentry role
  selector:
    app: truerepublic-node
  ports:
  - name: p2p
    port: 26656
    targetPort: 26656
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

# Configure (GRAFANA_PASSWORD is required; supply it from your secret store)
cp .env.example .env
sed -i "s/MONIKER=.*/MONIKER=aws-node-1/" .env
sed -i "s/^GRAFANA_PASSWORD=.*/GRAFANA_PASSWORD=${GRAFANA_PASSWORD:?set GRAFANA_PASSWORD}/" .env

# Start
make docker-build
docker compose up -d
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
