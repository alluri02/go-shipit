# Lesson 08: Best Practices & Security

> You now understand how Docker works, builds, isolates, ships, and connects. This final lesson is
> the **production checklist** — how to make images **small, safe, and trustworthy**. ShipIt's
> `FROM scratch` image already nails several of these; we'll make the rest explicit.

---

## The Production Checklist

| # | Practice | ShipIt status | Lesson |
|---|----------|---------------|--------|
| 1 | Small base image (multi-stage / `scratch`) | ✅ `FROM scratch`, ~10MB | 03 |
| 2 | Pin base images by version/digest | ⚠️ `golang:1.22-alpine` (pin digest for prod) | 05 |
| 3 | Run as **non-root** | ⚠️ add a `USER` | this lesson |
| 4 | `.dockerignore` to shrink context | ⚠️ add one | 03 |
| 5 | Least-privilege at runtime (drop caps) | ⚠️ `--cap-drop=ALL` | 04 |
| 6 | Scan images for CVEs | ⚠️ add to CI | this lesson |
| 7 | Sign & verify images | ⚠️ optional (Cosign) | this lesson |
| 8 | No secrets baked into images | ✅ env vars at runtime | this lesson |

---

## 1. Smaller Is Safer (and Faster)

A smaller image = **fewer packages = smaller attack surface = fewer CVEs = faster pulls.**

| Base | Size | Attack surface |
|------|------|----------------|
| `ubuntu` | ~78 MB | Full distro, shell, package manager |
| `alpine` | ~7 MB | Minimal distro, `sh`, `apk` |
| `gcr.io/distroless/static` | ~2 MB | **No shell, no package manager** |
| `scratch` | 0 MB | **Nothing** — just your binary |

ShipIt uses `scratch`: no shell means an attacker who lands RCE has **no `sh`, no `curl`, no
package manager** to pivot with. Minimalism *is* a security control.

> **Trade-off recap:** no shell also means no `docker exec sh` for debugging. If you need it, use
> `distroless` (still no package manager, but you can attach a debug sidecar).

---

## 2. Run as Non-Root

By default, a container process runs as **root** (UID 0). If it escapes isolation, it's root-ish on
the host. **Don't run as root.**

`scratch` has no `/etc/passwd`, so you can't `useradd`. Instead, create the user in the **builder**
stage and copy it forward, then run under a numeric UID:

```dockerfile
FROM golang:1.22-alpine AS builder
# ...build shipit...
RUN echo 'shipit:x:10001:10001::/nonexistent:/sbin/nologin' > /etc/passwd.min

FROM scratch
COPY --from=builder /app/shipit /shipit
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd.min /etc/passwd
USER 10001                       # ← run as non-root
ENTRYPOINT ["/shipit"]
CMD ["serve"]
```

Or enforce it at runtime without changing the image:

```bash
docker run --user 10001:10001 shipit:local
```

> **Note:** non-root can't bind ports below 1024. ShipIt listens on **8080** (not 80), so it works
> as non-root out of the box. That's an intentional 12-factor choice.

---

## 3. Least Privilege at Runtime

Even as non-root, drop every Linux capability you don't need (Lesson 04):

```bash
docker run \
  --cap-drop=ALL \                     # drop all root powers
  --security-opt=no-new-privileges \   # can't gain privileges via setuid
  --read-only \                        # root filesystem is read-only
  --tmpfs /tmp \                       # writable scratch in RAM only
  --user 10001:10001 \
  shipit:local
```

`--read-only` is powerful and easy with ShipIt: the app writes nothing to its own filesystem
(state lives in MySQL and the queue), so a read-only root FS **just works**.

---

## 4. Keep Secrets Out of Images

**Anyone who can pull your image can read every layer** — including a secret you thought a later
`RUN rm` deleted (it's still in the earlier layer!).

```dockerfile
# ❌ NEVER — baked into a layer forever, visible in `docker history`
ENV SHIPIT_SLACK_TOKEN=xoxb-super-secret
COPY id_rsa /root/.ssh/id_rsa
```

```bash
# ✅ Inject at RUN time via env vars (how ShipIt already does it)
docker run -e SHIPIT_SLACK_TOKEN=$SLACK_TOKEN shipit:local

# ✅ Or a secrets file / orchestrator secret
docker run --env-file ./secrets.env shipit:local
```

| Secret source | Good for |
|---------------|----------|
| `-e` / `--env-file` | Local dev |
| Docker/Swarm/K8s **secrets** | Production orchestrators |
| Cloud KeyVault / Secrets Manager | Cloud deploys (ShipIt → Azure) |
| BuildKit `--secret` mounts | Secrets needed **only at build time** |

> ShipIt already reads config from `SHIPIT_*` env vars (Go track **Lesson 09**) — so it's secret-safe
> by design. Never move those into the Dockerfile.

---

## 5. Scan for Vulnerabilities

Base images accumulate CVEs over time. Scan on every build:

```bash
# Docker's built-in (Snyk-powered) scanner
docker scout cves shipit:local

# Trivy — popular open-source scanner
trivy image shipit:local
```

Add it to the pipeline (extends the Go track's Lesson 15 CI):

```yaml
# In .github/workflows/ci.yml, after the docker build step
- name: Scan image for vulnerabilities
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: ghcr.io/${{ github.repository }}:${{ github.sha }}
    severity: CRITICAL,HIGH
    exit-code: '1'          # fail the build on High/Critical CVEs
```

> `scratch` images have almost nothing to scan — another win. Most CVEs come from the base distro,
> which ShipIt simply doesn't ship.

---

## 6. Sign & Verify Images (Supply Chain)

The `* Optional: image signature verification (Notary v2 / Cosign)` note in the diagram is this:
prove an image **really came from your pipeline** and wasn't tampered with.

```bash
# Sign after pushing (Cosign, keyless via OIDC in CI)
cosign sign ghcr.io/alluri02/go-shipit@sha256:9f0e...

# Verify before deploying
cosign verify ghcr.io/alluri02/go-shipit@sha256:9f0e...
```

The daemon's optional **"verify image signature"** step (diagram stage ②) is where this is
enforced with Docker Content Trust / policy controllers in production clusters.

---

## 7. Pin Everything

```dockerfile
FROM golang:1.22-alpine@sha256:abc123...   # digest-pinned base — reproducible, tamper-evident
```

```yaml
# CI already tags by immutable git SHA — deploy that, not :latest
ghcr.io/${{ github.repository }}:${{ github.sha }}
```

Mutable tags (`latest`, even `1.22-alpine`) can change under you. Digests can't. Pin base images
by digest and deploy app images by SHA.

---

## Try It

```bash
# Build and check the size + layers
docker build -t shipit:local .
docker image ls shipit:local

# Scan it (needs Docker Desktop / docker scout)
docker scout quickview shipit:local
docker scout cves shipit:local

# Run it hardened: non-root, no new privs, read-only FS, no capabilities
docker run --rm -p 8080:8080 \
  --user 10001:10001 \
  --cap-drop=ALL \
  --security-opt=no-new-privileges \
  --read-only --tmpfs /tmp \
  shipit:local

# Confirm no secrets are baked in
docker history --no-trunc shipit:local | Select-String -Pattern "TOKEN|PASSWORD|KEY"
```

---

## Key Takeaways

1. **Small = safe:** `scratch`/distroless shrink the attack surface to near zero.
2. **Never run as root** — set `USER` (numeric UID) or `--user`; ShipIt on port 8080 is non-root ready.
3. **Least privilege at runtime:** `--cap-drop=ALL`, `--no-new-privileges`, `--read-only`.
4. **Secrets never go in images** — inject via env/secret at runtime (ShipIt uses `SHIPIT_*` env).
5. **Scan every build** (`docker scout` / `trivy`) and **fail CI on High/Critical CVEs**.
6. **Sign and verify** images (Cosign/Notary) for supply-chain integrity.
7. **Pin by digest / deploy by SHA** — mutable tags are a reproducibility and security risk.

---

## 🎉 You've Completed the Docker Track!

You can now explain, end-to-end:

- **How Docker works** — CLI → `dockerd` → `containerd` → `runc` → kernel *(D01)*
- **What an image is** — read-only layers + copy-on-write *(D02)*
- **How to build one well** — multi-stage, `FROM scratch`, layer caching *(D03)*
- **How isolation works** — namespaces, cgroups, capabilities *(D04)*
- **Where images live** — registries, push/pull, digests *(D05)*
- **How to run a stack** — Compose with MySQL + Azurite *(D06)*
- **How containers connect & remember** — networking + volumes *(D07)*
- **How to harden it** — small, non-root, scanned, signed *(D08)*

Combined with the [Go track](../../README.md), you can now **build a production Go service *and*
ship it as a secure, minimal container.**

---

## Back to: [Docker Track Index](README.md) · [Main README](../../README.md)
