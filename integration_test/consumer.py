import json
import logging

from kafka import KafkaConsumer

from config import Config


# Consume the metrics messages
class MetricsConsumer:

    # Kafka consumer to consume the metrics messages
    @staticmethod
    def get_messages():
        logging.warning("IN: get_messages")
        offset = "earliest"
        messages = []
        consumer = KafkaConsumer(
            bootstrap_servers=Config.BOOTSTRAP_SERVERS.value,
            consumer_timeout_ms=(Config.POLLING_INTERVAL.value * 1000),
            auto_offset_reset=offset
        )
        logging.warning("\tConsumer Created...!!")
        consumer.subscribe([Config.KAFKA_TOPIC.value])
        logging.warning(f"\tTopic {Config.KAFKA_TOPIC.value} Subscribed...!!")
        for message in consumer:
            logging.info(f"\nloading...{message.value}")
            json_data = json.loads(message.value)
            messages.append(json_data)
            logging.info(len(messages))

        logging.warning("\tClosing the consumer")
        consumer.close()
        logging.warning("OUT: get_messages")
        return messages
