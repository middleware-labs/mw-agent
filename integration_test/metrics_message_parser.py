from config import MetricDataTypes


class Metrics:

    # Get the metrics and respective data from the received json metric message
    def get_metrics(self, metrics_data):
        metrics_attributes = {}
        if isinstance(metrics_data, dict):
            if "resourceMetrics" in metrics_data:
                for resourceMetric in metrics_data["resourceMetrics"]:

                    # Collect the resource attributes
                    resource_attributes = []
                    if "resource" in resourceMetric:
                        if "attributes" in resourceMetric["resource"]:
                            for attribute in resourceMetric["resource"]["attributes"]:
                                resource_attributes.append(attribute["key"])

                    if "scopeMetrics" in resourceMetric:
                        for scopeMetric in resourceMetric["scopeMetrics"]:
                            if "metrics" in scopeMetric:

                                # Collect the metric attributes
                                for metric in scopeMetric["metrics"]:
                                    name = metric["name"]
                                    attributes = []
                                    if MetricDataTypes.SUM.value in metric:
                                        attributes = self.get_metric_attributes(MetricDataTypes.SUM.value, metric)
                                    elif MetricDataTypes.GAUGE.value in metric:
                                        attributes = self.get_metric_attributes(MetricDataTypes.GAUGE.value, metric)
                                    elif MetricDataTypes.HISTOGRAM.value in metric:
                                        attributes = self.get_metric_attributes(MetricDataTypes.HISTOGRAM.value, metric)
                                    elif MetricDataTypes.EXPONENTIAL_HISTOGRAM.value in metric:
                                        attributes = self.get_metric_attributes(
                                            MetricDataTypes.EXPONENTIAL_HISTOGRAM.value, metric)
                                    elif MetricDataTypes.SUMMARY.value in metric:
                                        attributes = self.get_metric_attributes(MetricDataTypes.SUMMARY.value, metric)

                                    # Merge resource attributes and metric attributes
                                    metrics_attributes[name] = list(set(attributes + resource_attributes))

        return metrics_attributes

    # Get the metric attributes based on the @MetricDataTypes
    @staticmethod
    def get_metric_attributes(data_type, metric):
        attributes = []
        if "dataPoints" in metric[data_type]:
            for dataPoint in metric[data_type]["dataPoints"]:
                if "attributes" in dataPoint:
                    for attribute in dataPoint["attributes"]:
                        attributes.append(attribute["key"])
        return attributes
