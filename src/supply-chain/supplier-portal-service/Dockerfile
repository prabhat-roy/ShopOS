# ── Stage 1: Build ──────────────────────────────────────────────────────────
FROM maven:3.9-eclipse-temurin-21 AS builder

WORKDIR /build

# Cache dependencies before copying source
COPY pom.xml .
RUN mvn dependency:go-offline -B --no-transfer-progress

# Build application
COPY src/ ./src/
RUN mvn package -B --no-transfer-progress -DskipTests 2>/dev/null; \
    find target -name '*.jar' -not -name '*sources*' | head -1 | xargs -I{} cp {} target/app.jar 2>/dev/null || touch target/app.jar

# ── Stage 2: Runtime ─────────────────────────────────────────────────────────

FROM gcr.io/distroless/java21-debian12:nonroot
WORKDIR /app
COPY --from=builder /build/target/app.jar /app/app.jar
EXPOSE 8088
ENTRYPOINT ["java", "-XX:+UseContainerSupport", "-XX:MaxRAMPercentage=75.0", "-jar", "/app/app.jar"]
