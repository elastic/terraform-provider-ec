terraform {
  required_version = ">= 0.12.29"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.12.4"
    }

    aws = {
      source  = "hashicorp/aws"
      version = "0.12.4"
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
