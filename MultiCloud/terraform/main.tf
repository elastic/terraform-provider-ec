terraform {
  required_version = ">= 1.0.2"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = ">= 0.4.1"
    }
	  google = {
      source = "hashicorp/google"
      version = ">= 4.35.0" 
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = ">= 1.7.0"
    }
  }
}

