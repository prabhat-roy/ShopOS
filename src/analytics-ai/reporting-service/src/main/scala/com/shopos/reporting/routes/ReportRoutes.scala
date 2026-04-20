package com.shopos.reporting.routes

import cats.effect.IO
import com.shopos.reporting.service.ReportingService
import io.circe.Json
import io.circe.syntax._
import org.http4s._
import org.http4s.circe.CirceEntityEncoder._
import org.http4s.dsl.io._

object ReportRoutes {
  def routes(reportingService: ReportingService): HttpRoutes[IO] =
    HttpRoutes.of[IO] {

      case GET -> Root / "healthz" =>
        Ok(Json.obj("status" -> Json.fromString("ok")))

      case GET -> Root / "reports" / "daily-sales" :? DateQueryParam(date) =>
        for {
          report   <- reportingService.generateDailySales(date)
          response <- Ok(report.asJson)
        } yield response

      case GET -> Root / "reports" / "user-metrics" :? DateQueryParam(date) =>
        for {
          metrics  <- reportingService.generateUserMetrics(date)
          response <- Ok(metrics.asJson)
        } yield response

      case GET -> Root / "reports" / "top-products" :? StartQueryParam(start) +& EndQueryParam(end) +& OptLimitQueryParam(limitOpt) =>
        val limit = limitOpt.toOption.flatMap(_.headOption).getOrElse(10)
        for {
          products <- reportingService.generateTopProducts(start, end, limit)
          response <- Ok(products.asJson)
        } yield response
    }

  object DateQueryParam    extends QueryParamDecoderMatcher[String]("date")
  object StartQueryParam   extends QueryParamDecoderMatcher[String]("start")
  object EndQueryParam     extends QueryParamDecoderMatcher[String]("end")
  object OptLimitQueryParam extends OptionalMultiQueryParamDecoderMatcher[Int]("limit")
}
