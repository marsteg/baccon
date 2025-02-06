#!/bin/bash

set -e

# Create a namespace for your prometheus deployment
k create ns prom

helm install prom --namespace prom prometheus-community/prometheus -f prom-values.yaml

helm install grafana --namespace prom grafana/grafana -f grafana-values.yaml

