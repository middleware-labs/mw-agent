from enum import Enum


# Data types of metrics
class MetricDataTypes(Enum):
    SUM = "sum"
    GAUGE = "gauge"
    HISTOGRAM = "histogram"
    EXPONENTIAL_HISTOGRAM = "exponentialHistogram"
    SUMMARY = "summary"
