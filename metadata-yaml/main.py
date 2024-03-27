import json
import os
import shutil

import yaml

from config.attribute_type import AttributeType
from config.metric_data_types import MetricDataTypes
from config.aggregation_temporality import AggregationTemporality
from config.value_type import ValueType


# JSON file directory
source_directory = "input/todo"
destination_directory = "input/completed"
output_directory = "output"

#  Ensure directory exists
os.makedirs(source_directory, exist_ok=True)
os.makedirs(destination_directory, exist_ok=True)
os.makedirs(output_directory, exist_ok=True)


# Collect the metric attributes
def get_metric_attributes(data_type, metric):
    attributes = {}
    if "dataPoints" in metric[data_type]:
        for dataPoint in metric[data_type]["dataPoints"]:
            if "attributes" in dataPoint:
                for attribute in dataPoint["attributes"]:
                    attributes[attribute["key"]] = {
                        'description': "",
                        'type': AttributeType.get_attribute_type(attribute['value'])
                    }
    return attributes


# Get the value type of the metric
def get_value_type(data_type, metric):
    value_type = ValueType()
    if "dataPoints" in metric[data_type]:
        for dataPoint in metric[data_type]["dataPoints"]:
            return value_type.get_value_type_from_datapoint(dataPoint)


# Get the all the details of the metric data type
def get_metric_type(data_type, metric, attributes):
    metric_attributes = get_metric_attributes(data_type, metric)
    attributes.update(metric_attributes)

    metric_type = {}
    if "aggregationTemporality" in metric[data_type]:
        metric_type.update(
            {
                'aggregation_temporality': AggregationTemporality.get_aggregation_temporality(
                    metric[data_type]['aggregationTemporality']
                )
            }
        )
    if data_type == MetricDataTypes.SUM.value:
        metric_type.update(
            {
                'monotonic': metric[data_type]['isMonotonic'] if "isMonotonic" in metric[data_type] else False
            }
        )

    metric_type.update({'value_type': get_value_type(data_type, metric)})

    metric_data = {
        data_type: metric_type
    }
    if len(metric_attributes.keys()) > 0:
        metric_data.update({
            'attributes': list(metric_attributes.keys())
        })

    return metric_data


# Main execution function
def main():
    # List all files in the source directory
    json_files = [f for f in os.listdir(source_directory) if f.endswith('.json')]
    for file_name in json_files:
        source_file_path = os.path.join(source_directory, file_name)
        destination_file_path = os.path.join(destination_directory, file_name)
        yaml_file_path = os.path.join(output_directory, file_name[:-5] + "_metadata.yaml")  # Modify file name for YAML

        with open(source_file_path, 'r') as json_file:
            data = json.load(json_file)

            resource_attributes = {}
            attributes = {}
            metrics = {}

            # Collect the resource attributes
            for item in data['resourceMetrics']:
                for attribute in item['resource']['attributes']:
                    resource_attributes[attribute['key']] = {
                        'type': AttributeType.get_attribute_type(attribute['value']),
                        'enabled': 'true'
                    }

                # Collect the metrics and attributes
                for scope_metric in item['scopeMetrics']:
                    for metric in scope_metric['metrics']:
                        metrics[metric['name']] = {
                            'description': metric['description'],
                            'unit': metric['unit'],
                            'enabled': True
                        }

                        for data_type in MetricDataTypes:
                            if data_type.value in metric:
                                metrics[metric['name']].update(
                                    get_metric_type(data_type.value, metric, attributes)
                                )
                                break

            # Create final json object for yaml
            metadata_json = {
                'resource_attributes': resource_attributes,
                'metrics': metrics
            }
            if len(attributes) > 0:
                metadata_json['attributes'] = attributes

            # Create the yaml file
            with open(yaml_file_path, 'w') as yaml_file:
                yaml.safe_dump(metadata_json, yaml_file, indent=2, width=float("inf"))

            # Move the completed file to the another directory
            shutil.move(source_file_path, destination_file_path)


if __name__ == '__main__':
    main()
