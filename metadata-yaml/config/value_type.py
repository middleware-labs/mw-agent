# Map the metric data type the string
class ValueType:
    value_types = {
        "asInt": "int",
        "asDouble": "double",
        "asString": "string",
        "asBool": "bool",
        "asEmpty": "empty",
        "asMap": "map",
        "asSlice": "slice",
        "asBytes": "bytes"
    }

    # Get the value type from the string
    def get_value_type(self, value):
        return self.value_types.get(value)

    # Get the value type from the datapoint object
    def get_value_type_from_datapoint(self, datapoint: dict):
        return [self.get_value_type(key) for key in self.value_types.keys() if key in datapoint].pop()
