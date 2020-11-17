# Deployment Example

This example shows how to deploy multiple Elastic Cloud deployments, and reference one another through Cross Cluster Search using Terraform only.
The example creates two single node Elasticsearch clusters acting as sources for another Elasticsearch deployment with Cross Cluster Search enabled.

## Running the example

Build the provider using `make install` from the main folder. From within the example's directory, run `terraform init` to initialize Terraform, and `terraform apply` to apply the changes.
