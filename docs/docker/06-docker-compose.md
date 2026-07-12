# Lesson 06: Docker Compose

> Running one container is `docker run`. Running **a whole stack** — an app plus its database plus
> a queue — with one command is **Docker Compose**. This repo already ships a real
> [`docker-compose.yml`](../../docker-compose.yml) that boots ShipIt's dependencies. We'll read it
> line by line.

---

## The Problem Compose Solves

Without Compose, starting ShipIt's local dependencies means remembering this:

```bash
docker run -d --name shipit-mysql -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=devpass -e MYSQL_DATABASE=shipit \
  -v mysql_data:/var/lib/mysql mysql:8.0
docker run -d --name shipit-azurite -p 10000:10000 -p 10001:10001 -p 10002:10002 \
  -v azurite_data:/data mcr.microsoft.com/azure-storage/azurite
```

With Compose, all of that becomes **`docker compose up`**. The `docker run` flags become
**declarative YAML, checked into git.**

| Concept | It's basically... |
|---------|-------------------|
| `docker-compose.yml` | Your `docker run` flags, **as version-controlled code** |
| A **service** | One container definition (image + config) |
| `docker compose up` | Start the whole stack |
| A Compose **project** | A named group of services sharing a network |

---

## This Repo's `docker-compose.yml`

```yaml
services:
  # ─── MySQL 8 — deployments, environments, history ───
  mysql:
    image: mysql:8.0
    container_name: shipit-mysql
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: devpass
      MYSQL_DATABASE: shipit
    volumes:
      - mysql_data:/var/lib/mysql
      - ./schema/deployments.sql:/docker-entrypoint-initdb.d/01-deployments.sql:ro
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 5s
      timeout: 3s
      retries: 10

  # ─── Azurite — Azure Storage emulator (queue/blob/table) ───
  azurite:
    image: mcr.microsoft.com/azure-storage/azurite
    container_name: shipit-azurite
    ports:
      - "10000:10000"  # Blob
      - "10001:10001"  # Queue
      - "10002:10002"  # Table
    volumes:
      - azurite_data:/data

volumes:
  mysql_data:
  azurite_data:
```

---

## Reading It Key by Key

| Key | What it does | Maps to `docker run` flag |
|-----|--------------|---------------------------|
| `image:` | Which image to run | *(the image name)* |
| `container_name:` | Fixed, friendly container name | `--name shipit-mysql` |
| `ports:` | Publish container ports to the host | `-p 3306:3306` |
| `environment:` | Set env vars inside the container | `-e MYSQL_ROOT_PASSWORD=devpass` |
| `volumes:` | Mount named volumes / bind mounts | `-v mysql_data:/var/lib/mysql` |
| `healthcheck:` | How Docker probes liveness | `--health-cmd ...` |

### The two volume lines mean different things

```yaml
- mysql_data:/var/lib/mysql                  # NAMED VOLUME — persistent DB data
- ./schema/deployments.sql:/docker-entrypoint-initdb.d/01-deployments.sql:ro  # BIND MOUNT (read-only)
```

- **Named volume** (`mysql_data`): Docker-managed storage that **survives `down`** → your data
  persists across restarts. (Deep dive in Lesson 07.)
- **Bind mount** (`./schema/...`): maps a **file from the repo** into the container. MySQL runs any
  `.sql` in `/docker-entrypoint-initdb.d/` on first boot — so **ShipIt's schema auto-loads**.
  `:ro` = read-only (the container can't modify your source file).

---

## Deep Dive: Healthchecks & Startup Ordering

```yaml
healthcheck:
  test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
  interval: 5s      # probe every 5s
  timeout: 3s       # fail a probe after 3s
  retries: 10       # unhealthy after 10 straight failures
```

MySQL takes a few seconds to accept connections. The healthcheck lets Docker (and other services)
know when it's **actually ready**, not just started. A dependent service can wait for it:

```yaml
# Example: make an app service wait until MySQL is HEALTHY, not just running
depends_on:
  mysql:
    condition: service_healthy
```

> **Why this matters for ShipIt:** the app needs MySQL reachable before it runs migrations/queries.
> `depends_on` + `condition: service_healthy` prevents the classic "connection refused on startup"
> race. This ties to the Go track's **Lesson 11 (context)** — connect with a timeout and retry.

---

## Service Discovery: Services Talk by Name

Compose puts all services on **one private network** and gives each a DNS name equal to its
**service name**. So inside the Compose network, ShipIt would reach MySQL at `mysql:3306` — **not**
`127.0.0.1:3306`.

```
host machine        →  127.0.0.1:3306   (via the published port)
another container   →  mysql:3306       (via Compose DNS, service name)
```

> That's why `ports:` (host access) and inter-service hostnames are **different**. Published ports
> are for *you*; service names are for *container-to-container* traffic.

---

## Essential Commands

```bash
docker compose up -d          # start the whole stack in the background
docker compose ps             # list services + health status
docker compose logs -f mysql  # follow one service's logs
docker compose exec mysql mysql -uroot -pdevpass shipit   # shell into a service
docker compose stop           # stop containers (keep them)
docker compose down           # stop AND remove containers + network
docker compose down -v        # ...also delete named volumes (wipes DB data!)
```

---

## Try It (end-to-end with ShipIt)

```bash
# 1. Boot MySQL + Azurite
docker compose up -d

# 2. Wait for MySQL to be healthy
docker compose ps        # STATUS should show "healthy"

# 3. Point ShipIt at the local stack and run it (PowerShell)
$env:SHIPIT_DATABASE_URL = "root:devpass@tcp(127.0.0.1:3306)/shipit?parseTime=true"
$env:SHIPIT_QUEUE_CONNECTION_STRING = "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1"
.\shipit.exe serve

# 4. Verify the schema auto-loaded via the bind mount
docker compose exec mysql mysql -uroot -pdevpass shipit -e "SHOW TABLES;"

# 5. Tear down (keep data) vs wipe everything
docker compose down          # keeps mysql_data volume
# docker compose down -v     # deletes volumes too
```

---

## Key Takeaways

1. **Compose = your `docker run` flags as declarative, version-controlled YAML.**
2. **One `docker compose up`** starts a whole multi-container stack.
3. **Named volume vs bind mount:** persistent Docker storage vs mapping a repo file in.
4. **The schema bind mount** auto-loads ShipIt's tables on MySQL's first boot.
5. **Healthchecks + `depends_on: service_healthy`** fix startup race conditions.
6. **Services reach each other by service name** (`mysql:3306`), not `localhost`.
7. **`down -v` deletes volumes** — the fast way to lose your dev database.

---

## Next: [Lesson 07 — Networking & Volumes](07-networking-and-volumes.md)
We'll go deeper on the two things Compose set up for us: networks and persistent storage.
