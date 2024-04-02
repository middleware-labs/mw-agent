# Get the attribute type from the JSON value
class AttributeType:
    @staticmethod
    def get_attribute_type(value: dict):
        attribute_type = {
            "stringValue": "string",
            "intValue": "int",
            "boolValue": "bool",
            "doubleValue": "double",
            "arrayValue": "array",
            "kvlistValue": "kvlist",
            "bytesValue": "bytes"
        }
        return [attribute_type.get(key) for key in list(value.keys()) if key in attribute_type].pop()
