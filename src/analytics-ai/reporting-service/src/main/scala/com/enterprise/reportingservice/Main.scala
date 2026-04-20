package com.enterprise.reportingservice

object Main extends App {

  val port: Int = sys.env.getOrElse("PORT", "50000").toInt

  println(s"reporting-service starting on port $port")

  // TODO: Start HTTP server (e.g., Akka HTTP or http4s) and expose /healthz
  //   Example with Akka HTTP:
  //     import akka.actor.ActorSystem
  //     import akka.http.scaladsl.Http
  //     import akka.http.scaladsl.server.Directives._
  //     implicit val system = ActorSystem("reporting-service")
  //     val route = path("healthz") { get { complete("""{"status":"ok"}""") } }
  //     Http().newServerAt("0.0.0.0", port).bind(route)

  // TODO: Start Kafka consumer subscribing to analytics.* topics
  //   Example with Alpakka Kafka:
  //     val consumerSettings = ConsumerSettings(system, new StringDeserializer, new StringDeserializer)
  //       .withBootstrapServers(sys.env.getOrElse("KAFKA_BROKERS", "kafka:9092"))
  //       .withGroupId("reporting-service-group")
  //     Consumer.committableSource(consumerSettings, Subscriptions.topicPattern("analytics\\..*"))
  //       .runForeach(msg => println(s"Received: ${msg.record.value()}"))

  // Graceful shutdown hook
  sys.addShutdownHook {
    println("reporting-service shutting down")
  }

  // Keep the JVM alive (replace with server Await when HTTP server is wired)
  Thread.currentThread().join()
}
