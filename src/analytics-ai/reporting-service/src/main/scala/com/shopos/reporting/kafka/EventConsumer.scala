package com.shopos.reporting.kafka

import cats.effect.IO
import com.shopos.reporting.Config
import com.shopos.reporting.service.ReportingService
import fs2.kafka._
import org.typelevel.log4cats.slf4j.Slf4jLogger

object EventConsumer {

  def stream(config: Config, reportingService: ReportingService): fs2.Stream[IO, Unit] = {
    val logger = Slf4jLogger.getLogger[IO]

    val consumerSettings: ConsumerSettings[IO, String, String] =
      ConsumerSettings[IO, String, String]
        .withAutoOffsetReset(AutoOffsetReset.Earliest)
        .withBootstrapServers(config.kafkaBrokers)
        .withGroupId(config.kafkaGroupId)

    KafkaConsumer.stream(consumerSettings).flatMap { consumer =>
      fs2.Stream.eval(consumer.subscribeTo(config.kafkaTopics.head, config.kafkaTopics.tail: _*)) >>
        consumer.records
          .evalMap { committable =>
            val record = committable.record
            val topic  = record.topic
            val value  = record.value

            reportingService
              .processEvent(topic, value)
              .handleErrorWith { err =>
                logger.error(err)(s"Error processing event from topic $topic: ${err.getMessage}")
              } >>
              committable.offset.commit
          }
    }
  }
}
