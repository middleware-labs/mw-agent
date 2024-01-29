import os
from enum import Enum


# Environment configurations
class Config(Enum):
    # Kafka Configuration
    BOOTSTRAP_SERVERS = os.getenv("KAFKA_BOOTSTRAP_SERVER", "localhost:9092")
    POLLING_INTERVAL = int(os.getenv("KAFKA_CONSUMER_POLLING_INTERVAL", "2"))
    KAFKA_TOPIC = os.getenv("KAFKA_TOPIC", "otlp_metrics")

    # Checks file path
    CHECKS_FILE_PATH = os.getenv("CHECKS_FILE_PATH", "./checks/checks.csv")
