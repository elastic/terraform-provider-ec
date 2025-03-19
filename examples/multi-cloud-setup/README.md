# Elastic Terraform Examples to build an Multi Cloud Monitoring environment

The project in this repository is creating an Elastic Cloud environment in order to getting started with monitoring and protecting your Cloud Service Providers(CSP) environment in Google, AWS and/or Azure. It is creating all necessary components within the CSPs as well as the in Elastic Cloud using terraform. The whole process will be done in less than 1h. 

You can either install every Cloud Environment separatly or choose the MultiCloud project to install everything at once. By choosing MultiCloud the terraform script will also configure the necessary connection between the clusters in order to do Cross Cluster Search(CSS). Because of that each cluster can live in its own Cloud Provider environment (GCP cluster in GCP, AWS cluser in AWS and so on). This will guarantee a low cost footprint when collecting the relevant data from the providers. But because of CCS every cluster can get queried by one main cluster. 

## The AWS environment

The AWS example is creating an AWS Monitoring and Enhanced Security environment. It creates all necessary AWS Services as well as the Elastic Cloud Cluster for you. The only thing you need to provide is are AWS account credentials that provide the right permissions as well as the Elastic Cloud API Key. It works both: In [Elastic Cloud directly](https://cloud.elastic.co) or via the [AWS Marketplace option for Elastic Cloud](https://ela.st/aws).

This example will install and configure:
- Elastic Cluster
- AWS EC2 instance with Elastic Agent installed and configured to talk to the Elastic Cluster 
- Elastic Agent will be configured to collect available Metric datasets with zero manual configuration
- Elastic SAR app will be used to install the elastic serverless forwarder to collect Logs from S3 and CloudWatch Log Groups 
- The Elastic Cluster will be configured with the following additional capabilities
	- Preloaded all Elastic Security Detection rules and enabled all AWS related rules

## The Google Cloud Environment

The Google Cloud example is creating a Google Cloud Monitoring and Enhanced Security environment. It creates all necessary Google Cloud Services as well as the Elastic Cloud Cluster for you. The only thing you need to provide is an appropriate Google Cloud Service account that has the right permissions and the Elastic Cloud API Key. It works both: In [Elastic Cloud directly](https://cloud.elastic.co) or via the [Google Cloud Marketplace option for Elastic Cloud](https://ela.st/google).

This example will install and configure:
- Elastic Cluster
- Google Cloud Compute engine with Elastic Agent installed and configured to talk to the Elastic Cluster
- Google Cloud Log routers (Log sinks) with the appropriate filters for Audit, Firewall, VPC Flow, DNS and Loadbalancer Logs. 
- Google Cloud PubSub topics to collects the log types above
- Elastic Agent will be configured to collect all the logs and all available Google Cloud Metric datasets with zero manual configuration
- The Elastic Cluster will be configured with the following additional capabilities
	- Single pane of glass Google Cloud Dashboard
	- Google Cloud Cost optimizer dashboard
	- Google Cloud Storage bucket analyzer dashboard
	- Elastic transforms to prepare the data for the installed dashboards
	- Preloaded all Elastic Security Detection rules and enabled all Google Cloud related rules

## Getting started

You can decide if you like to install the environment for all Cloud Providers at once or each once independently from each other. No matter what you prefer you need to deploy it within the [MultiCloud](MultiCloud) folder. Before you do that you need to prepare your environment. You will find the comprehensive Getting Started description also within the [MultiCloud](MultiCloud) folder.

# More Elasticsearch terraform examples

Other terraform + elastic examples can be found here:
- [Patent Search](https://github.com/MarxDimitri/solution-accelerators/tree/main/patent-search) using Google Cloud BigQuery public dataset

Kibana Dashboards and other Elastic extensions can be found here
- [Elastic Content Share](https://elastic-content-share.eu/)
- [AWS Cloudformation template](https://elastic-content-share.eu/blog/how-to-create-elastic-cloud-cluster-via-aws-cloud-formation-template/)

 
