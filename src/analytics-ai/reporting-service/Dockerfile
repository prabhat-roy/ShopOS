FROM eclipse-temurin:21-jdk AS builder

WORKDIR /build

RUN apt-get update && apt-get install -y --no-install-recommends curl bash && \
    curl -fsSL "https://github.com/sbt/sbt/releases/download/v1.10.2/sbt-1.10.2.tgz" -o sbt.tgz && \
    tar xzf sbt.tgz -C /usr/local --strip-components=1 && \
    rm sbt.tgz && \
    rm -rf /var/lib/apt/lists/*

COPY build.sbt ./
COPY project ./project

RUN sbt update

COPY src ./src

RUN sbt assembly

FROM gcr.io/distroless/java21-debian12:nonroot
WORKDIR /app
COPY --from=builder /build/target/scala-2.13/reporting-service-assembly-*.jar /app/app.jar
EXPOSE 8701
ENTRYPOINT ["java", "-XX:+UseContainerSupport", "-XX:MaxRAMPercentage=75.0", "-jar", "/app/app.jar"]
