import asyncio
import logging
import uvicorn

from pipeline.api import app, set_store
from pipeline.config import settings
from pipeline.consumer import KafkaPipelineConsumer
from pipeline.store import InMemoryPipelineStore

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)-8s %(name)s - %(message)s",
)
logger = logging.getLogger(__name__)


async def start_kafka_consumer(store: InMemoryPipelineStore) -> None:
    consumer = KafkaPipelineConsumer(store)
    try:
        await consumer.start()
        await consumer.consume()
    except Exception as exc:
        logger.error("Kafka consumer error: %s", exc, exc_info=True)
    finally:
        await consumer.stop()


async def main() -> None:
    store = InMemoryPipelineStore()
    set_store(store)

    config = uvicorn.Config(
        app,
        host="0.0.0.0",
        port=settings.HTTP_PORT,
        log_level="info",
    )
    server = uvicorn.Server(config)

    kafka_task = asyncio.create_task(start_kafka_consumer(store))
    http_task  = asyncio.create_task(server.serve())

    logger.info("data-pipeline-service starting on port %s", settings.HTTP_PORT)

    await asyncio.gather(http_task, kafka_task, return_exceptions=True)


if __name__ == "__main__":
    asyncio.run(main())