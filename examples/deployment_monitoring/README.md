# Deployment Example

This examples shows how to deploy an Elastic Cloud deployment using Terraform only.
The created resources are:
    
    * Deployment with 2 resources:
        * 4Gb single-node Elasticsearch cluster.
        * 1Gb Kibana Instance.
    * (Monitoring) Deployment with 1 resource:
        * 1Gb single-node Elasticsearch cluster, where the other single-node Elasticsearch
          cluster sends its monitoring info to.

## Running the example

run `terraform apply` to see it work.