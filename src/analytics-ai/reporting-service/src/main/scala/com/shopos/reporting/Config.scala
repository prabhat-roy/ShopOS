package com.shopos.reporting

final case class Config(
    httpPort: Int,
    kafkaBrokers: String,
    kafkaTopics: List[String],
    kafkaGroupId: String,
    cassandraHosts: List[String],
    cassandraKeyspace: String
)

object Config {
  def fromEnv(): Config = {
    val httpPort        = sys.env.getOrElse("HTTP_PORT", "8701").toInt
    val kafkaBrokers    = sys.env.getOrElse("KAFKA_BROKERS", "localhost:9092")
    val topicsRaw       = sys.env.getOrElse("KAFKA_TOPICS", "analytics.page.viewed,analytics.product.clicked")
    val kafkaTopics     = topicsRaw.split(",").map(_.trim).toList
    val kafkaGroupId    = sys.env.getOrElse("KAFKA_GROUP_ID", "reporting-service")
    val cassandraRaw    = sys.env.getOrElse("CASSANDRA_HOSTS", "localhost")
    val cassandraHosts  = cassandraRaw.split(",").map(_.trim).toList
    val cassandraKeyspace = sys.env.getOrElse("CASSANDRA_KEYSPACE", "reporting")

    Config(
      httpPort        = httpPort,
      kafkaBrokers    = kafkaBrokers,
      kafkaTopics     = kafkaTopics,
      kafkaGroupId    = kafkaGroupId,
      cassandraHosts  = cassandraHosts,
      cassandraKeyspace = cassandraKeyspace
    )
  }
}
