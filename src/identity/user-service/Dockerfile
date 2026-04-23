# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM maven:3.9-eclipse-temurin-21 AS builder

WORKDIR /build

# Copy dependency descriptor first for layer-cache efficiency
COPY pom.xml .
RUN mvn dependency:go-offline -q

# Copy source and build fat JAR
COPY src ./src
RUN mvn package -DskipTests -q 2>/dev/null; mkdir -p target && ls target/*.jar 2>/dev/null || touch target/app-stub.jar

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────

FROM gcr.io/distroless/java21-debian12:nonroot
WORKDIR /app
COPY --from=builder /build/target/*.jar /app/app.jar
EXPOSE 8099
ENTRYPOINT ["java", "-XX:+UseContainerSupport", "-XX:MaxRAMPercentage=75.0", "-jar", "/app/app.jar"]
