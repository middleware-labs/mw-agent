package utils

import "os"

func InitializeEnv() {
	_, hasTarget := os.LookupEnv("TARGET")
	if !hasTarget {
		os.Setenv("TARGET", "us1-v1-grpc.melt.so:443")
	}

	_, hasCollectionType := os.LookupEnv("MELT_COLLECTION_TYPE")
	if !hasCollectionType {
		os.Setenv("MELT_COLLECTION_TYPE", "all")
	}
}
