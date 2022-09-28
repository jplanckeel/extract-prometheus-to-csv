## Describe 

[Golang] Script to extract a prometheus metric to csv.

Use case: We need to extract a metric over a wide range of times to ingest into a data platform to compare with a data metric.

Problem: This metric has a lot of cardinality and it is not possible to extract a year without killing the prometheus platform :( 

# Run

```
URL="https://prometheus_url" go run main.go

