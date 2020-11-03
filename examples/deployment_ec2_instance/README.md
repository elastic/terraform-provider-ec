# Deployment with an EC2 Instance example

This example code shows how easy it is to build an application's infrastrucute using ec2 instances and an Elastic Cloud deployment communicating securly using traffic filters
The code creates an ec2 instance in your default vpc & subnet, but uses the instance's public IP address to configure a traffic filter connecting it back to the Elastic Cloud deployment.
Such communication can also be done through this terraform provider using an AWS Private link.

## Running the example
Build the provider using `make install` from the main folder
run `terrafrom init` to initialize your terraform cli
run `terraform apply` to see it work.
