package com.shopos.reporting

import cats.effect.{IO, IOApp, Resource}
import com.comcast.ip4s._
import com.shopos.reporting.kafka.EventConsumer
import com.shopos.reporting.routes.ReportRoutes
import com.shopos.reporting.service.ReportingService
import com.shopos.reporting.store.InMemoryReportStore
import org.http4s.ember.server.EmberServerBuilder
import org.http4s.server.middleware.Logger
import org.typelevel.log4cats.slf4j.Slf4jLogger

object Main extends IOApp.Simple {

  def run: IO[Unit] = {
    val logger = Slf4jLogger.getLogger[IO]

    for {
      _      <- logger.info("Starting reporting-service...")
      config  = Config.fromEnv()
      _      <- logger.info(s"HTTP port: ${config.httpPort}, Kafka brokers: ${config.kafkaBrokers}")
      store  <- InMemoryReportStore()
      svc     = ReportingService(store)
      _      <- serverResource(config, svc).use { _ =>
                  val kafkaFiber =
                    EventConsumer.stream(config, svc)
                      .compile
                      .drain
                      .handleErrorWith { err =>
                        logger.error(err)(s"Kafka consumer error: ${err.getMessage}")
                      }
                  logger.info("Server and Kafka consumer started") >>
                    kafkaFiber
                }
    } yield ()
  }

  private def serverResource(config: Config, svc: ReportingService): Resource[IO, org.http4s.server.Server] = {
    val routes    = ReportRoutes.routes(svc)
    val loggedApp = Logger.httpRoutes[IO](logHeaders = true, logBody = false)(routes).orNotFound

    val port = Port.fromInt(config.httpPort).getOrElse(port"8701")

    EmberServerBuilder
      .default[IO]
      .withHost(host"0.0.0.0")
      .withPort(port)
      .withHttpApp(loggedApp)
      .build
  }
}
