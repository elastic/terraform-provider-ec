---
page_title: "Provider: Elastic Cloud (EC)"
---

# Elastic Cloud Provider

The Elastic Cloud can be used to configure and manage Elastic Cloud deployments using the Elastic Cloud
APIs. Documentation regarding the Data Sources and Resources supported by the Elastic Cloud provider can be found in the navigation to the left.

## Example Usage


```hcl
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

The Elastic Cloud provider offers two methods of authentication against the remote API. Depending on the target environment, one or all can be used. By default the `endpoint` which the provider will target is the Public API of the Elasticsearch Service (ESS) offering.

Only one of `apikey` or `username` and `password`  can be specified at one time. The Elasticsearch Service (ESS) offering, only supports API Keys as the authentication mechanism. When targeting an ECE
Installation, `username` and `password` can be used.

!> **Warning:** Hard-coding credentials into any Terraform configuration is not
recommended, and risks secret leakage should this file ever be committed to a
public version control system.

### API Key authentication (Recommended)

API Keys are the recommended authentication method. They can be used when authenticating against the Elasticsearch Service (ESS) or Elastic Cloud Enterprise (ECE).

They can either be hardcoded in the provider `.tf` provider configuration (NOT RECOMMENDED). Or specified via environment variables: `EC_API_KEY`.

```hcl
provider "ec" {
  apikey = "my-api-key"
}
```

### Username and Password login (ECE)

A `username` and `password` combination can be used to authenticate when targeting an ECE environment. 
Note that `username` and `password` is not a supported method when 

They can either be hardcoded in the provider `.tf` provider configuration (NOT RECOMMENDED). Or specified via environment variables: `EC_USERNAME` or `EC_USER` and `EC_PASSWORD` or `EC_PASS`.

```hcl
provider "ec" {
  # ECE installation endpoint
  endpoint = "https://my.ece-environment.corp"

  # If the ECE installation has a self-signed certificate
  # setting "insecure" to true is required.
  insecure = true

  username = "my-username"
  password = "my-password"
}
```

## Argument Reference

In addition to [generic `provider` arguments](https://www.terraform.io/docs/configuration/providers.html)
(e.g. `alias` and `version`), the following arguments are supported in the Elastic Cloud (EC) `provider` block:

* `endpoint` - (Optional) This is the target endpoint. It must be provided only when
  aiming to use the Elastic Cloud provider against an ECE installation or ESS Private.

* `apikey` - (Optional) This is the EC API Key. Required when targeting the Elasticsearch
  Service (ESS) offering, but also valid when targeting an ECE installation. It must be
  provided, but it can also be sourced from the `EC_API_KEY` environment variable.
  Conflicts with `username` and `password` authentication options.

* `username` - (Optional) This is the EC username. It must be provided, but it can also
  be sourced from the `EC_USER` or `EC_USERNAME` environment variables. Conflicts with
  `apikey`. Not recommended.

* `password` - (Optional) This is the EC password. It must be provided, but it can also
  be sourced from the `EC_PASS` or `EC_PASSWORD` environment variables. Conflicts with 
  `apikey`. Not recommended.

* `insecure` - (Optional) This setting allows the provider to skip TLS verification.
  Useful when targeting installation with self-signed certificates. Not recommended when
  targeting the Elasticsearch Service (ESS).

* `insecure` - (Optional) This setting allows the user to set a custom timeout in the
  individual HTTP request level. Defaults to "1m" but might need to be tweaked if timeouts
  are experienced.

* `verbose` - (Optional) When set to true, it'll write a "requests.json" file in the folder
  where terraform is executed with all outgoing HTTP requests and responses. Defaults to "false".

* `verbose_credentials` - (Optional) When set with `verbose`, the contents of the Authorization
header will not be redacted. Defaults to `"false"`.

* `verbose_file` - (Optional) Sets the file where the verbose request / response http flow will
be written to. Defaults to `"request.log"`.
