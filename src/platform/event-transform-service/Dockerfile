FROM clojure:temurin-21-lein-alpine AS builder

WORKDIR /app

COPY project.clj .
RUN lein deps

COPY src src
RUN lein uberjar

FROM eclipse-temurin:21-jre-alpine

RUN adduser -D -u 1001 nonroot
USER 1001

WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app/target/uberjar/event-transform-service.jar app.jar

EXPOSE 50354

ENTRYPOINT ["java", "-jar", "app.jar"]
