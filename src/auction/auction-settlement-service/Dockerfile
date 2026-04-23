FROM maven:3.9-eclipse-temurin-21 AS builder

WORKDIR /app

COPY pom.xml .
RUN mvn dependency:go-offline -B

COPY src src
RUN mvn package -DskipTests -B

FROM eclipse-temurin:21-jre-alpine

RUN adduser -D -u 1001 nonroot
USER 1001

WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app/target/*.jar app.jar

EXPOSE 50313

ENTRYPOINT ["java", "-jar", "app.jar"]
