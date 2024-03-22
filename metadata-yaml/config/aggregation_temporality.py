# Map the value of aggregation temporality to the respective string
class AggregationTemporality:
    @staticmethod
    def get_aggregation_temporality(aggregation):
        aggregation_map = {
            0: "unknown",
            1: "delta",
            2: "cumulative"
        }
        if aggregation in aggregation_map:
            return aggregation_map[aggregation]
        else:
            raise ValueError("Invalid Aggregation!")
