from checks import Checks
from consumer import MetricsConsumer


def main():
    # Read Checks from csv file
    print("Integration Testing Started...!!!")
    checks = Checks()
    print("checks.checks_metrics_attributes")

    # Read messages from kafka 'otlp_metrics' topic for particular time interval
    messages = MetricsConsumer.get_messages()

    print("Validation Stage started...")
    for message in messages:
        checks.validate_metrics(message)
    print("Validation Stage Completed...")

    # Generate the final report and act accordingly.
    report = checks.generate_report()
    if report:
        raise Exception(report)
    else:
        print("All the checks have been run successfully")
    print("Integration testing completed...")


if __name__ == '__main__':
    main()
