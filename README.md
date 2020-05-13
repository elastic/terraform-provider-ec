# Terraform Provider for Elastic Cloud

Terraform provider for the Elastic Cloud API, including:

* Elasticsearch Service (ESS).
* Elastic Cloud Enterprise (ECE).
* Elasticsearch Service Private (ESSP).

## Example usage

```hcl
provider "ec" {
  # ECE installation endpoint
  endpoint = "https://my.ece-environment.corp"

  # If the ECE installation has a self-signed certificate
  # setting "insecure" to true is required.
  insecure = true

  # APIKey is the recommended authentication mechanism. When
  # Targeting the Elasticsearch Service, APIKeys are the only
  # valid authentication mechanism.
  apikey = "my-apikey"

  # When targeting ECE installations, username and password
  # authentication is allowed.
  username = "my-username"
  password = "my-password"
}


# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "my_example_deployment"

  # Mandatory fields
  region                 = "us-east-1"
  version                = "7.6.2"
  deployment_template_id = "aws-io-optimized"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highio.i3"
    }
  }

  kibana {
    topology {
      instance_configuration_id = "aws.kibana.r4"
    }
  }
}
```

## Developer Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.12+
- [Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (please check the [requirements](https://github.com/terraform-providers/terraform-provider-aws#requirements) before proceeding).

*Note:* This project uses [Go Modules](https://blog.golang.org/using-go-modules) making it safe to work with it outside of your existing [GOPATH](http://golang.org/doc/code.html#GOPATH). The instructions that follow assume a directory in your home directory outside of the standard GOPATH (i.e `$HOME/development/terraform-providers/`).
