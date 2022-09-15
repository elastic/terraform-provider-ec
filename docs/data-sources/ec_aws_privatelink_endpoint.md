---
page_title: "Elastic Cloud: ec_aws_privatelink_endpoint"
description: |-
  Retrieves infomation about the AWS Private Link configuration for a given region.
---

# Data Source: ec_aws_privatelink_endpoint

Use this data source to retrieve information about the AWS Private Link configuration for a given region. Further documentation on how to establish a PrivateLink connection can be found in the ESS [documentation](https://www.elastic.co/guide/en/cloud/current/ec-traffic-filtering-vpc.html).

~> **NOTE:** This data source provides data relevant to the Elasticsearch Service (ESS) only, and should not be used for ECE.

## Example Usage

```hcl
data "ec_aws_privatelink_endpoint" "us-east-1" {
  region = "us-east-1"
}
```

## Argument Reference

* `region` (Required) - Region to retrieve the Private Link configuration for.

## Attributes Reference

* `vpc_service_name` - The VPC service name used to connect to the region.
* `domain_name` - The domain name to used in when configuring a private hosted zone in the VPCE connection.
* `zone_ids` - The IDs of the availability zones hosting the VPC endpoints.
