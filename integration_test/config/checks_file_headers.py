from enum import Enum


# Column names of the csv file
class ChecksFileHeaders(Enum):
    CHECK_NO = "check_no"
    METRIC = "metric"
    ATTRIBUTES = "attributes"
