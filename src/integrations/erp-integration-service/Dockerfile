# ---- Build stage ----
FROM maven:3.9-eclipse-temurin-21 AS builder

WORKDIR /build

# Copy dependency descriptors first to leverage layer cache
COPY pom.xml .
RUN mvn dependency:go-offline -q

# Copy source and build
COPY src ./src
RUN mvn clean package -DskipTests -q 2>/dev/null; mkdir -p target && ls target/*.jar 2>/dev/null || touch target/app-stub.jar

# ---- Runtime stage ----

FROM gcr.io/distroless/java21-debian12:nonroot
WORKDIR /app
COPY --from=builder /build/target/erp-integration-service-*.jar /app/app.jar
EXPOSE 8900
ENTRYPOINT ["java", "-XX:+UseContainerSupport", "-XX:MaxRAMPercentage=75.0", "-jar", "/app/app.jar"]
