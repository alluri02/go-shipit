# Lesson 02: Images & Layers

> An image looks like one thing (`nginx`, `mysql:8.0`) but it's really a **stack of read-only
> layers** glued together by a **union filesystem**. Understanding layers is the key to fast
> builds, small images, and cheap containers.

---

## Concept Map

| Docker concept | It's basically... | Analogy |
|----------------|-------------------|---------|
| **Image** | Immutable, read-only template | A **class** / a compiled `.jar` / a `.zip` |
| **Layer** | One immutable filesystem diff | A **git commit** — a diff on top of the last |
| **Tag** | Human-friendly pointer to an image | A **git branch/tag** (`nginx:1.27`) |
| **Digest** | Content hash of an image | A **git commit SHA** (`sha256:...`) |
| **Container** | Image + one writable layer, running | An **object** — `new Image()` |

---

## An Image Is a Stack of Layers

Each instruction in a `Dockerfile` that changes the filesystem creates **one layer**. Layers are
**stacked** and presented as a single filesystem via a **union/overlay** driver (`overlay2`).

```
        IMAGE  (read-only)                     CONTAINER  (adds 1 writable layer)
┌────────────────────────────┐        ┌────────────────────────────┐
│ Layer 4:  COPY app  binary  │        │ ✏️  Writable layer (this container) │
├────────────────────────────┤        ├────────────────────────────┤
│ Layer 3:  RUN apt install   │        │ Layer 4  (read-only, shared) │
├────────────────────────────┤   →    ├────────────────────────────┤
│ Layer 2:  COPY go.mod        │        │ Layer 3  (read-only, shared) │
├────────────────────────────┤        ├────────────────────────────┤
│ Layer 1:  FROM alpine (base) │        │ Layer 2  (read-only, shared) │
└────────────────────────────┘        │ Layer 1  (read-only, shared) │
                                        └────────────────────────────┘
```

**The magic:** those read-only layers are **shared** across every container and every image that
uses them. Run 100 `nginx` containers → the image layers exist **once** on disk. Each container
only adds its own tiny writable layer.

---

## Copy-on-Write (CoW)

A container never modifies image layers — they're read-only. When a process **writes** to a file:

1. Docker **copies** the file up from the read-only layer into the container's writable layer.
2. The write happens on that copy.
3. Reads of unchanged files come straight from the shared read-only layers.

```
Read  /etc/nginx/nginx.conf   → served from read-only image layer (no copy)
Write /etc/nginx/nginx.conf   → file copied UP to writable layer, then modified
```

> **Why it matters:** starting a container is **instant and cheap** — nothing is copied up front.
> But it also means **container writes are ephemeral**: delete the container and the writable
> layer is gone. Persistent data belongs in a **volume** (Lesson 07).

---

## Tags vs Digests

```bash
nginx                 # → nginx:latest       (tag, mutable — can point elsewhere tomorrow)
nginx:1.27            # → a specific version (tag, still mutable)
nginx@sha256:abc123…  # → an exact image     (digest, immutable — always the same bytes)
```

| | Tag | Digest |
|-|-----|--------|
| Example | `mysql:8.0` | `mysql@sha256:9f0e...` |
| Mutable? | ✅ can be re-pointed | ❌ content-addressed, never changes |
| Use for | humans, dev | **production pinning, reproducibility** |

> **Production tip:** pin by digest (or an immutable version tag) so a redeploy can't silently
> pull different bytes. `latest` in production is how 3am pages happen.

---

## Inspecting Layers

```bash
# Pull the same base image ShipIt builds on
docker pull golang:1.22-alpine

# See each layer and the command that created it
docker history golang:1.22-alpine

# See the full metadata, including the layer digests
docker inspect golang:1.22-alpine --format '{{ json .RootFS.Layers }}'

# See total size and how much is shared vs unique
docker image ls
docker system df -v          # detailed: shared size, unique size, reclaimable
```

Example `docker history` output (read bottom-up = build order):

```
IMAGE          CREATED BY                                      SIZE
<missing>      COPY app /shipit                                12MB     ← our binary
<missing>      RUN go build ...                                0B
<missing>      COPY go.mod go.sum ./                           2kB
<missing>      FROM alpine:3.19                                7MB      ← base
```

---

## Why Go Images Are Tiny (and Java/C# Aren't)

Layers explain the size difference from the Go track's Lesson 15:

| Base | Why | Typical final image |
|------|-----|---------------------|
| **Go** `FROM scratch` | Static binary, **no runtime layers needed** | **~10 MB** |
| **C#** `aspnet:8.0-alpine` | Needs the .NET runtime layers | ~100 MB |
| **Java** `temurin:21-jre-alpine` | Needs the JVM layers | ~150 MB |

A Go binary is self-contained, so the "runtime" layers simply don't exist. Fewer/smaller layers
= smaller image = faster pulls = faster deploys. **This is a layering win, not magic.**

---

## Layer Caching = Fast Builds

Docker caches each layer. On rebuild, it reuses a cached layer **as long as that instruction and
everything before it are unchanged**. Change one line and **every layer after it is invalidated**.

This is *why* the ShipIt `Dockerfile` copies `go.mod`/`go.sum` and downloads deps **before**
copying the source — deps change rarely, source changes constantly. (Full deep dive in Lesson 03.)

```dockerfile
COPY go.mod go.sum ./     # ← cached until deps change
RUN go mod download       # ← cached (expensive step, rarely re-run)
COPY . .                  # ← invalidated on every code change (cheap)
RUN go build ...
```

---

## Try It

```bash
# Build the ShipIt image and watch layers being created / cached
docker build -t shipit:local .

# Rebuild after changing a .go file — note which layers say "CACHED"
docker build -t shipit:local .

# Inspect the layers you just built
docker history shipit:local

# See how little unique space it uses
docker image ls shipit:local
```

---

## Key Takeaways

1. **An image = a stack of read-only layers**; a container adds **one writable layer** on top.
2. **Layers are shared** — 100 containers from one image store the image layers **once**.
3. **Copy-on-write:** writes copy a file up to the writable layer; container writes are **ephemeral**.
4. **Tags are mutable, digests are immutable** — pin by digest in production.
5. **Fewer/smaller layers = smaller images** — that's why `FROM scratch` Go images are ~10MB.
6. **Order your Dockerfile by change frequency** to maximize layer-cache hits.

---

## Next: [Lesson 03 — Dockerfile Deep Dive](03-dockerfile-deep-dive.md)
We'll dissect this repo's real multi-stage `Dockerfile` line by line.
