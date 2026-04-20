ThisBuild / scalaVersion := "2.13.12"
ThisBuild / version      := "0.1.0-SNAPSHOT"
ThisBuild / organization := "com.shopos"

lazy val root = (project in file("."))
  .settings(
    name := "reporting-service",
    libraryDependencies ++= Seq(
      // HTTP server
      "org.http4s" %% "http4s-ember-server" % "0.23.25",
      "org.http4s" %% "http4s-ember-client" % "0.23.25",
      "org.http4s" %% "http4s-dsl"          % "0.23.25",
      "org.http4s" %% "http4s-circe"        % "0.23.25",
      // JSON
      "io.circe" %% "circe-generic" % "0.14.7",
      "io.circe" %% "circe-parser"  % "0.14.7",
      // Cats Effect
      "org.typelevel" %% "cats-effect" % "3.5.4",
      // Kafka (fs2-kafka)
      "com.github.fd4s" %% "fs2-kafka" % "3.3.0",
      // Logging
      "ch.qos.logback" % "logback-classic" % "1.4.14",
      "org.typelevel" %% "log4cats-slf4j"  % "2.6.0",
      // Testing
      "org.scalatest" %% "scalatest"                     % "3.2.18" % Test,
      "org.typelevel" %% "cats-effect-testing-scalatest" % "1.5.0"  % Test
    ),
    assembly / mainClass := Some("com.shopos.reporting.Main"),
    assembly / assemblyMergeStrategy := {
      case PathList("META-INF", "services", _ @_*) => MergeStrategy.concat
      case PathList("META-INF", _ @_*)             => MergeStrategy.discard
      case PathList("reference.conf")              => MergeStrategy.concat
      case _                                       => MergeStrategy.first
    }
  )
