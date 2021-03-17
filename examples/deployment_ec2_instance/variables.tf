variable "region" {
  default = "us-east-1"
}

variable "aws_profile" {
  default = "<aws profile>"
}

variable "ec_api_key" {
  default = "<enter your Elastic Cloud API Key>"
}

variable "keypair" {
  default = "<enter a key pair>"
}

variable "ubuntu18_ami" {
  # This AMI ID is only valid for us-east-1.
  default = "ami-0817d428a6fb68645"
}
