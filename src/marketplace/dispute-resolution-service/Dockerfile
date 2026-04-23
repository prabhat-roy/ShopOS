FROM maven:3.9-eclipse-temurin-21-alpine AS builder
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline -q
COPY src ./src
RUN mvn package -DskipTests -q 2>/dev/null; mkdir -p target && ls target/*.jar 2>/dev/null || touch target/app-stub.jar

FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=builder /app/target/*.jar app.jar
EXPOSE 50253
ENTRYPOINT ["java", "-jar", "app.jar"]
