from checks import Checks
from consumer import MetricsConsumer
import logging


def main():
    # Read Checks from csv file
    logging.warning("Integration Testing Started...!!!")
    checks = Checks()
    logging.warning("checks.checks_metrics_attributes")

    # Read messages from kafka 'otlp_metrics' topic for particular time interval
    messages = MetricsConsumer.get_messages()

    logging.warning("Validation Stage started...")
    for message in messages:
        checks.validate_metrics(message)
    logging.warning("Validation Stage Completed...")

    # Generate the final report and act accordingly.
    report = checks.generate_report()

    # Write the report content in file
    with open('/usr/bin/report.txt', 'w') as file:
        if report:
            logging.warning(f"Checks failed: {report}")
            for check in report:
                file.write(check + '\n')
        else:
            logging.warning("Checks passed: All the checks have been run successfully")
            file.write("All the checks have been run successfully")

    # logging.warning("Integration testing completed...")


if __name__ == '__main__':
    main()
