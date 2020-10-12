# Deployment Example

This example shows how to deploy an Elastic Cloud deployment using Terraform.
It will also initialize the instance by using some of the outputs as string variables within a very simple bash script, to create a user and an index.

Lastly, it also creates a traffic filter (which allows all traffic) and attaches that to the deployment as well.

## Running the example
build the provider using `make install` from the main folder
run `terrafrom init` to initialize your terraform cli
run `terraform apply` to see it work.
