# Limitations

This document aims to document the limitations of the terraform provider

## Version field diff

When upgrading the version of a deployment from A to B, the diff will look like:

```diff
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # ec_deployment.example_minimal will be updated in-place
  ~ resource "ec_deployment" "example_minimal" {
        deployment_template_id = "aws-io-optimized"
        id                     = "a941d93be8abf207b0169d0da565c8d1"
        name                   = "my_example_deployment"
        region                 = "us-east-1"
      ~ version                = "7.6.2" -> "7.7.0"

        elasticsearch {
            ref_id      = "main-elasticsearch"
            region      = "us-east-1"
            resource_id = "3443cf5a7df74d358939bfeaff7dbbf5"
            version     = "7.6.2"

            topology {
                instance_configuration_id = "aws.data.highio.i3"
                memory_per_node           = "4g"
                node_count_per_zone       = 0
                node_type_data            = true
                node_type_ingest          = true
                node_type_master          = true
                node_type_ml              = false
                zone_count                = 1
            }
        }

        kibana {
            elasticsearch_cluster_ref_id = "main-elasticsearch"
            ref_id                       = "main-kibana"
            region                       = "us-east-1"
            resource_id                  = "1bbd119426724995b2cae0d1c8d9f22e"
            version                      = "7.6.2"

            topology {
                instance_configuration_id = "aws.kibana.r4"
                memory_per_node           = "1g"
                node_count_per_zone       = 0
                zone_count                = 1
            }
        }
    }
```

vs 

```diff
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  # ec_deployment.example_minimal will be updated in-place
  ~ resource "ec_deployment" "example_minimal" {
        deployment_template_id = "aws-io-optimized"
        id                     = "a941d93be8abf207b0169d0da565c8d1"
        name                   = "my_example_deployment"
        region                 = "us-east-1"
      ~ version                = "7.6.2" -> "7.7.0"

        elasticsearch {
            ref_id      = "main-elasticsearch"
            region      = "us-east-1"
            resource_id = "3443cf5a7df74d358939bfeaff7dbbf5"
          ~ version     = "7.6.2" -> "7.7.0"

            topology {
                instance_configuration_id = "aws.data.highio.i3"
                memory_per_node           = "4g"
                node_count_per_zone       = 0
                node_type_data            = true
                node_type_ingest          = true
                node_type_master          = true
                node_type_ml              = false
                zone_count                = 1
            }
        }

        kibana {
            elasticsearch_cluster_ref_id = "main-elasticsearch"
            ref_id                       = "main-kibana"
            region                       = "us-east-1"
            resource_id                  = "1bbd119426724995b2cae0d1c8d9f22e"
          ~ version                      = "7.6.2" -> "7.7.0"

            topology {
                instance_configuration_id = "aws.kibana.r4"
                memory_per_node           = "1g"
                node_count_per_zone       = 0
                zone_count                = 1
            }
        }
    }
```

The terraform issue can be read here <https://github.com/hashicorp/terraform-plugin-sdk/issues/459>

## Elasticsearch resource

### Cluster monitoring

Stopping monitoring on a cluster, doesn't seem to be possible through the API. Reported in <https://github.com/elastic/cloud/issues/57821>.

This would happen when the block `monitoring_settings` of an Elasticsearch resource is removed.
