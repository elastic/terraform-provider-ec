terraform {
  required_version = ">= 0.12.29"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.12.3"
    }

    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.45, != 5.71.0"
    }
  }
}

provider "ec" {
  apikey = var.ec_api_key
}

provider "aws" {
  region  = var.region
  profile = var.aws_profile
}
