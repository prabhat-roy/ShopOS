# ============================================================
# Stage 1: Build
# ============================================================
FROM maven:3.9-eclipse-temurin-21 AS builder

WORKDIR /build

# Copy dependency manifests first for layer caching
COPY pom.xml .
RUN mvn dependency:go-offline -q

# Copy sources and build the fat-jar
COPY src ./src
RUN mvn package -DskipTests -q

# ============================================================
# Stage 2: Runtime
# ============================================================

FROM gcr.io/distroless/java21-debian12:nonroot
WORKDIR /app
COPY --from=builder /build/target/kyc-aml-service-*.jar /app/app.jar
EXPOSE 8307
ENTRYPOINT ["java", "-XX:+UseContainerSupport", "-XX:MaxRAMPercentage=75.0", "-jar", "/app/app.jar"]
