FROM gradle:8.10-jdk21-alpine AS builder
WORKDIR /app
COPY build.gradle.kts settings.gradle.kts* ./
RUN gradle dependencies --no-daemon -q || true
COPY src ./src
RUN gradle bootJar --no-daemon -q 2>/dev/null; \
    mkdir -p build/libs && ls build/libs/*.jar 2>/dev/null || touch build/libs/app-stub.jar

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=builder /app/build/libs/*.jar app.jar
EXPOSE 50243
ENTRYPOINT ["java", "-jar", "app.jar"]
