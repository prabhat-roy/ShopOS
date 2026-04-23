FROM gradle:8.10-jdk21 AS builder

WORKDIR /app

COPY build.gradle.kts settings.gradle.kts* gradle.properties* ./
RUN gradle dependencies --no-daemon || true

COPY src src
RUN gradle bootJar --no-daemon -x test

FROM eclipse-temurin:21-jre-alpine

RUN adduser -D -u 1001 nonroot
USER 1001

WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app/build/libs/*.jar app.jar

EXPOSE 50321

ENTRYPOINT ["java", "-jar", "app.jar"]
