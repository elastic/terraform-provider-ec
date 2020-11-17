# Deployment example

This example shows how to deploy an Elastic Cloud deployment using Terraform.
First, you initialize the instance by using some of the outputs as string variables within a bash script to create a user and an index.
Then, you create a traffic filter (which allows all traffic) and attach it to the deployment.

## Running the example

To run the example, follow these steps:

1. Build the provider by running `make install` from the main folder.
2. Run `terrafrom init` to initialize your Terraform CLI.
3. Run `terraform apply` to see how it works.
