# Deployment example

This example shows how to deploy an Elastic Cloud deployment using Terraform only.
The created resources are a single-node Elasticsearch cluster with a Kibana, APM and Enterprise Search instances.

## Running the example

Build the provider using `make install` from the main folder. From within the example's directory run `terrafrom init` to initialize Terrafrom, and `terraform apply` to see it work.
