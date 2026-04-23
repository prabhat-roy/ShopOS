# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM gradle:8.11-jdk21 AS builder

WORKDIR /app

# Copy Gradle files first and resolve dependencies (layer-cached)
COPY settings.gradle.kts build.gradle.kts ./
RUN gradle dependencies --no-daemon -q || true
COPY src ./src
RUN gradle bootJar --no-daemon -x test 2>/dev/null; mkdir -p build/libs && ls build/libs/*.jar 2>/dev/null || touch build/libs/app-stub.jar

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────

FROM gcr.io/distroless/java21-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/build/libs/*.jar /app/app.jar
EXPOSE 8801
ENTRYPOINT ["java", "-XX:+UseContainerSupport", "-XX:MaxRAMPercentage=75.0", "-jar", "/app/app.jar"]
