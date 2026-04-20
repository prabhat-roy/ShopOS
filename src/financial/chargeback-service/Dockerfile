# =============================================================
# chargeback-service — multi-stage Docker build
# Stage 1: Build with Maven
# Stage 2: Minimal JRE runtime, non-root user
# =============================================================

# ---- Stage 1: Build ----
FROM maven:3.9-eclipse-temurin-21 AS builder

WORKDIR /build

COPY pom.xml .
RUN mvn dependency:go-offline -B --no-transfer-progress

COPY src ./src
RUN mvn package -DskipTests -B --no-transfer-progress

# ---- Stage 2: Runtime ----

FROM gcr.io/distroless/java21-debian12:nonroot
WORKDIR /app
COPY --from=builder /build/target/chargeback-service-*.jar /app/app.jar
EXPOSE 50191
ENTRYPOINT ["java", "-XX:+UseContainerSupport", "-XX:MaxRAMPercentage=75.0", "-jar", "/app/app.jar"]
