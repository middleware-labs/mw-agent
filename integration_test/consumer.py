import json

from kafka import KafkaConsumer

from config import Config


# Consume the metrics messages
class MetricsConsumer:

    # Kafka consumer to consume the metrics messages
    @staticmethod
    def get_messages():
        print("IN: get_messages")
        offset = "earliest"
        messages = []
        consumer = KafkaConsumer(
            bootstrap_servers=Config.BOOTSTRAP_SERVERS.value,
            consumer_timeout_ms=(Config.POLLING_INTERVAL.value * 1000),
            # auto_offset_reset=offset
        )
        print("\tConsumer Created...!!")
        consumer.subscribe([Config.KAFKA_TOPIC.value])
        print(f"\tTopic {Config.KAFKA_TOPIC.value} Subscribed...!!")
        for message in consumer:
            json_data = json.loads(message.value)
            messages.append(json_data)
            print(len(messages))

        print("\tClosing the consumer")
        consumer.close()
        print("OUT: get_messages")
        return messages
