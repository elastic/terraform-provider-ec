# Deployment Example

This example shows how to deploy an Elastic Cloud deployment using Terraform.
Secondly, it will show you how to initialize the instance. This is done by using some of the outputs as string variables within a very simple bash script to create a user and an index.

Lastly, it will creates a traffic filter (which allows all traffic) and attach it to the deployment.

## Running the example
Build the provider using `make install` from the main folder
run `terrafrom init` to initialize your terraform cli
run `terraform apply` to see it work.
