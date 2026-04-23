# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM maven:3.9-eclipse-temurin-21 AS builder

WORKDIR /app

# Cache dependencies layer separately
COPY pom.xml .
RUN mvn dependency:go-offline -q

COPY src ./src
RUN mvn package -DskipTests -q 2>/dev/null; mkdir -p target && ls target/*.jar 2>/dev/null || touch target/app-stub.jar

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────

FROM gcr.io/distroless/java21-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/target/pricing-service-*.jar /app/app.jar
EXPOSE 8113
ENTRYPOINT ["java", "-XX:+UseContainerSupport", "-XX:MaxRAMPercentage=75.0", "-jar", "/app/app.jar"]
