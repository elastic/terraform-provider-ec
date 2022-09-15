---
page_title: "Elastic Cloud: ec_gcp_private_service_connect_endpoint"
description: |-
  Retrieves infomation about the GCP Private Service Connect configuration for a given region.
---

# Data Source: ec_gcp_private_service_connect_endpoint

Use this data source to retrieve information about the Azure Private Link configuration for a given region. Further documentation on how to establish a PrivateLink connection can be found in the ESS [documentation](https://www.elastic.co/guide/en/cloud/current/ec-traffic-filtering-psc.html).

~> **NOTE:** This data source provides data relevant to the Elasticsearch Service (ESS) only, and should not be used for ECE.

## Example Usage

```hcl
data "ec_gcp_private_service_connect_endpoint" "us-central1" {
  region = "us-central1"
}
```

## Argument Reference

* `region` (Required) - Region to retrieve the Private Link configuration for.

## Attributes Reference

* `service_attachment_uri` - The service attachment URI to attach the PSC endpoint to.
* `domain_name` - The domain name to point towards the PSC endpoint.
