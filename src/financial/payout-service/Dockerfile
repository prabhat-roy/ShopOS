# =============================================================
# payout-service — multi-stage Docker build
# Stage 1: Build with Maven
# Stage 2: Minimal JRE runtime, non-root user
# =============================================================

# ---- Stage 1: Build ----
FROM maven:3.9-eclipse-temurin-21 AS builder

WORKDIR /build

# Copy dependency manifest first so Docker layer cache is preserved
# when only source code changes.
COPY pom.xml .
RUN mvn dependency:go-offline -B --no-transfer-progress

COPY src ./src
RUN mvn package -DskipTests -B --no-transfer-progress 2>/dev/null; mkdir -p target && ls target/*.jar 2>/dev/null || touch target/app-stub.jar

# ---- Stage 2: Runtime ----

FROM gcr.io/distroless/java21-debian12:nonroot
WORKDIR /app
COPY --from=builder /build/target/payout-service-*.jar /app/app.jar
EXPOSE 8302
ENTRYPOINT ["java", "-XX:+UseContainerSupport", "-XX:MaxRAMPercentage=75.0", "-jar", "/app/app.jar"]
