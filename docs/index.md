---
page_title: "Provider: Elastic Cloud"
---

# Elastic Cloud Provider

The Elastic Cloud Terraform provider can be used to configure and manage Elastic Cloud deployments using the Elastic Cloud
APIs. Use the navigation to the left to read about data sources and resources supported by the Elastic Cloud provider. Elastic Cloud APIs are available for:

* Elasticsearch Service (ESS).
* Elastic Cloud Enterprise (ECE).
* Elasticsearch Service Private (ESSP).

## Releases

Interested in the provider's latest features, or want to make sure you're up to date? [Check out the provider changelog](https://github.com/elastic/terraform-provider-ec/blob/master/CHANGELOG.md).

## Authentication

The Elastic Cloud Terraform provider offers two methods of authentication against the remote API: `apikey` or a combination of `username` and `password`. Depending on the environment, you may choose one over the other. The Public API of Elasticsearch Service (ESS) is the default `endpoint` that the provider will target.

Elasticsearch Service (ESS) only supports `apikey`. Elastic Cloud Enterprise (ECE) supports `apikey` or a combination of `username` and `password`.

!> **Warning:** Hard-coding credentials into a Terraform configuration is not recommended, and risks secret leakage should this file ever be committed to a public version control system.

### API key authentication (recommended)

API keys are the recommended authentication method. They can be used to authenticate against Elasticsearch Service or Elastic Cloud Enterprise.

#### Generating an Elasticsearch Service (ESS) API Key

To generate an API key, follow these steps:

  1. Open you browser and navigate to <https://cloud.elastic.co/login>.
  2. Log in with your email and password.
  3. Click on [Elasticsearch Service](https://cloud.elastic.co/deployments).
  4. Navigate to [Features > API Keys](https://cloud.elastic.co/deployment-features/keys) and click on **Generate API Key**.
  5. Choose a name for your API key.
  6. Save your API key somewhere.

#### Using your API Key on the Elastic Cloud terraform provider

After you've generated your API Key, you can make it available to the Terraform provider by exporting it as the environment variable `EC_API_KEY` (recommended), or hardcoded in the provider `.tf` configuration file (supported but not recommended).

```sh
$ export EC_API_KEY="<apikey value>"
```

Or set the `apikey` field in the "ec" provider to the value of your generated API key.

```hcl
provider "ec" {
  apikey = "<apikey value>"
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
