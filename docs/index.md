---
page_title: "Provider: Elastic Cloud"
---

# Elastic Cloud Provider

The Elastic Cloud Terraform provider can be used to configure and manage Elastic Cloud deployments using the Elastic Cloud
APIs. Use the navigation to the left to read about data sources and resources supported by the Elastic Cloud provider.

## Example Usage


```hcl
terraform {
  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "0.1.0"
    }
  }
}

provider "ec" {
  apikey = "my-api-key"
}

resource "ec_deployment" "my_deployment" {
  name = "my example deployment"

  version                = "7.8.1"
  region                 = "us-east-1"
  deployment_template_id = "aws-io-optimized-v2"

  elasticsearch {
    topology {
      instance_configuration_id = "aws.data.highio.i3"
    }
  }

  kibana {
    topology {
      instance_configuration_id = "aws.kibana.r5d"
    }
  }
}
```

## Authentication

The Elastic Cloud Terraform provider offers two methods of authentication against the remote API: `apikey` or a combination of `username` and `password`. Depending on the environment, one or both can be used. The Public API of Elasticsearch Service (ESS) is the default `endpoint` that the provider will target.

Elasticsearch Service (ESS) only supports `apikey`. Elastic Cloud Enterprise (ECE) supports both `apikey` and a combination of `username` and `password`.

!> **Warning:** Hard-coding credentials into any Terraform configuration is not
recommended, and risks secret leakage should this file ever be committed to a
public version control system.

### API key authentication (recommended)

API keys are the recommended authentication method. They can be used to authenticate against Elasticsearch Service or Elastic Cloud Enterprise.

They can either be specified with the `EC_API_KEY` environment variable (recommended) or hardcoded in the provider `.tf` configuration file (supported but not recommended).

```hcl
provider "ec" {
  apikey = "my-api-key"
}
```

### Username and password login (ECE)

If you are targeting an ECE environment, you can also use a combination of `username` and `password` as authentication method. 

They can either be hardcoded in the provider `.tf` configuration (not recommended), or specified with the following environment variables: `EC_USERNAME` or `EC_USER` and `EC_PASSWORD` or `EC_PASS`.

```hcl
provider "ec" {
  # ECE installation endpoint
  endpoint = "https://my.ece-environment.corp"

  # If the ECE installation has a self-signed certificate
  # you must set insecure to true.
  insecure = true

  username = "my-username"
  password = "my-password"
}
```

## Argument Reference

In addition to [generic `provider` arguments](https://www.terraform.io/docs/configuration/providers.html)
(for ex. `alias` and `version`), the following arguments are supported in the Elastic Cloud `provider` block:

* `endpoint` - (Optional) This is the target endpoint. It must be provided only when
   you use the Elastic Cloud provider with an ECE installation or ESS Private.

* `apikey` - (Optional) This is the Elastic Cloud API key. It is required with ESS, but it is also valid with ECE. It must be
  provided, but it can also be sourced from the `EC_API_KEY` environment variable.
  Conflicts with `username` and `password` authentication options.

* `username` - (Optional) This is the Elastic Cloud username. It must be provided, but it can also
  be sourced from the `EC_USER` or `EC_USERNAME` environment variables. Conflicts with
  `apikey`. Not recommended.

* `password` - (Optional) This is the Elastic Cloud password. It must be provided, but it can also
  be sourced from the `EC_PASS` or `EC_PASSWORD` environment variables. Conflicts with 
  `apikey`. Not recommended.

* `insecure` - (Optional) This setting allows the provider to skip TLS verification.
  Useful when targeting installation with self-signed certificates. Not recommended when
  targeting ESS.

* `timeout` - (Optional) This setting allows the user to set a custom timeout in the
  individual HTTP request level. Defaults to 1 minute (`"1m"`), but might need to be tweaked if timeouts
  are experienced.

* `verbose` - (Optional) When set to `true`, it writes a `requests.json` file in the folder
  where Terraform runs with all the outgoing HTTP requests and responses. Defaults to `false`.

* `verbose_credentials` - (Optional) When set with `verbose`, the contents of the Authorization
header will not be redacted. Defaults to `false`.

* `verbose_file` - (Optional) Sets the file where the verbose request and response HTTP flow will
be written to. Defaults to `request.log`.
