# Nexus Repository Manager â€” Artifact Proxy & Package Registry

## Role in ShopOS

Nexus Repository Manager is the universal artifact proxy and private package registry for
ShopOS. It sits between CI pipelines and the public internet, caching dependencies from Maven
Central, npm, PyPI, Go proxy, and Docker Hub. This eliminates redundant internet downloads across
130 service builds, dramatically improving CI build speeds and providing resilience against upstream
registry outages.

Key capabilities:
- Proxy repositories â€” caches upstream packages locally on first download
- Hosted repositories â€” stores ShopOS's own published artifacts (Java JARs, Docker images)
- Repository groups â€” exposes a single URL that aggregates multiple repos (proxy + hosted)
- RBAC â€” restricts publish access to CI robot accounts
- Component metadata â€” BOM generation, license scanning integration

---

## Repository Inventory

| Repository Name | Type | Format | Remote / Purpose |
|---|---|---|---|
| `maven-central` | Proxy | Maven 2 | https://repo1.maven.org/maven2/ |
| `npm-registry` | Proxy | npm | https://registry.npmjs.org |
| `pypi-proxy` | Proxy | PyPI | https://pypi.org |
| `go-proxy` | Proxy | Go | https://goproxy.io |
| `docker-hub` | Proxy | Docker | https://registry-1.docker.io |
| `shopos-releases` | Hosted | Maven 2 | Internal Java/Kotlin JARs (released versions) |
| `shopos-docker` | Hosted | Docker | Internal Docker images (alternative to Harbor) |

---

## Dependency Caching Benefits for CI

### Without Nexus (Direct Internet)

```
Build 1: mvn install â†’ downloads 200MB from Maven Central (40s)
Build 2: mvn install â†’ downloads 200MB again from Maven Central (40s)
Build N: same download repeated every cold CI runner
```

### With Nexus (Proxy Cache)

```
Build 1: mvn install â†’ Nexus fetches from Maven Central, caches locally (40s)
Build 2: mvn install â†’ Nexus serves from local cache (2s)
Build N: same 2s â€” regardless of internet connectivity
```

| Metric | Without Nexus | With Nexus |
|---|---|---|
| Maven build (cold) | ~40s download | ~2s (cache hit) |
| npm install (cold) | ~25s download | ~1s (cache hit) |
| pip install (cold) | ~15s download | ~1s (cache hit) |
| Resilience to outages | None | Builds succeed even if upstream is down |
| Bandwidth usage | 100% internet | Reduced by ~90% after warmup |

---

## Per-Language Configuration

### Maven (Java / Kotlin services)

```xml
<!-- settings.xml â€” configure Nexus as mirror -->
<mirror>
  <id>nexus</id>
  <mirrorOf>*</mirrorOf>
  <url>http://nexus:8081/repository/maven-central/</url>
</mirror>
```

### npm (Node.js services)

```bash
npm config set registry http://nexus:8081/repository/npm-registry/
```

### pip (Python services)

```bash
pip install --index-url http://nexus:8081/repository/pypi-proxy/simple/ -r requirements.txt
```

### Go modules

```bash
export GOPROXY=http://nexus:8081/repository/go-proxy/,direct
```

### Docker (pull-through cache)

```json
// /etc/docker/daemon.json
{
  "registry-mirrors": ["http://nexus:8081"]
}
```

---

## Publishing ShopOS Artifacts

Java/Kotlin services publish their JARs to `shopos-releases` after a successful CI pipeline run.
This enables other services to consume shared libraries as Maven dependencies rather than copying
source.

```xml
<!-- pom.xml distributionManagement -->
<distributionManagement>
  <repository>
    <id>nexus-releases</id>
    <url>http://nexus:8081/repository/shopos-releases/</url>
  </repository>
</distributionManagement>
```

---

## Connection Details

| Parameter | Value |
|---|---|
| HTTP Port | 8081 |
| Context Path | `/` |
| Admin User | `admin` |
| Default Admin Password | Set on first boot (stored in Vault in production) |
| Data Volume | `/nexus-data` |
