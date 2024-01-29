import ast
import csv

from config import Config, ChecksFileHeaders
from metrics_message_parser import Metrics


class Checks:

    def __init__(self):
        self.checks_metrics_attributes = self.get_checks()

    # Validate metrics
    def validate_metrics(self, message):
        metrics_obj = Metrics()
        # Collect metrics
        metric_attributes = metrics_obj.get_metrics(message)
        # Run checks
        self.run_checks(metric_attributes)

    # Get the list of checks from the respective check file
    @staticmethod
    def get_checks():
        checks_metrics_attributes = {}
        with open(Config.CHECKS_FILE_PATH.value, 'r') as file:
            csv_reader = csv.DictReader(file)

            # Iterate over rows
            for row in csv_reader:
                required_attributes = ast.literal_eval(row[ChecksFileHeaders.ATTRIBUTES.value])
                metric_attributes = {}
                for attribute in required_attributes:
                    metric_attributes[attribute] = False

                metric_properties = {
                    'check_no': row[ChecksFileHeaders.CHECK_NO.value],
                    'is_available': False,
                    'attributes': metric_attributes
                }
                checks_metrics_attributes[row[ChecksFileHeaders.METRIC.value]] = metric_properties
        return checks_metrics_attributes

    # Run all the checks from check list
    def run_checks(self, metric_attributes):
        for metric in self.checks_metrics_attributes.keys():
            if metric in metric_attributes:
                self.checks_metrics_attributes[metric]["is_available"] = True
                for attribute in self.checks_metrics_attributes[metric][ChecksFileHeaders.ATTRIBUTES.value].keys():
                    if attribute in metric_attributes[metric]:
                        self.checks_metrics_attributes[metric][ChecksFileHeaders.ATTRIBUTES.value][attribute] = True

    # Generate the report based on all checks
    def generate_report(self):
        report = []
        for metric in self.checks_metrics_attributes:
            check_no = self.checks_metrics_attributes[metric][ChecksFileHeaders.CHECK_NO.value]
            if not self.checks_metrics_attributes[metric]["is_available"]:
                report.append(f"Check {check_no}: Metric '{metric}' not found")
            else:
                attributes = []
                for attribute in self.checks_metrics_attributes[metric][ChecksFileHeaders.ATTRIBUTES.value]:
                    if not self.checks_metrics_attributes[metric][ChecksFileHeaders.ATTRIBUTES.value][attribute]:
                        attributes.append(attribute)
                if attributes:
                    report.append(f"Check {check_no}: Attribute(s) {attributes} were not found in Metric '{metric}'")

        return report
