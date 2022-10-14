# -------------------------------------------------------------
# Terraform provider configuration
# -------------------------------------------------------------
terraform {
  required_version = ">= 1.0.2"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = ">= 0.4.1"
    }
	  azurerm = {
      source  = "hashicorp/azurerm"
      version = ">=3.0.0"
    }
  }
}




