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

You can decide if you like to install the environment for all Cloud Providers at once or each once independently from each other. No matter what you prefer you need to deploy it within the [MultiCloud](MultiCloud) folder. Before you do that you need to prepare your environment.

### Prepare software dependencies

- [jq](https://stedolan.github.io/jq/download/)
- [terraform](https://www.terraform.io/downloads)

### Clone the repository

```bash
git clone https://github.com/felix-lessoer/elastic-terraform-examples.git
```

### Create Elastic Cloud ID following this steps

[Create EC API key](https://registry.terraform.io/providers/elastic/ec/latest/docs#api-key-authentication-recommended)

Set env variable for Elastic Cloud:

```bash
export EC_API_KEY="[PUT YOUR ELASTIC CLOUD API KEY HERE]"
```

### Create local env files within the repo

```bash
mkdir local_env
touch terraform.tfvars.json
```

Modify the terraform environment settings to prepare your local env.

#### For AWS
More AWS configuration remarks you can find in the [AWS](AWS) folder.

Minimal config:
```json
{
	"aws_region" : "eu-west-2",	 
	"aws_access_key" : "<YOUR ACCESS KEY>",
	"aws_secret_key" : "<YOUR SECRET KEY>"
}
```

List of other optional parameters that can be added to terraform.tfvars.json 
| Parameter Name  | Default value | Example | Description |
| ------------- | ------------- | ------------- | ------------- |
| elastic_version  | latest  | 8.4.1  | Used to define the Elastic Search version  |
| elastic_region  | aws-eu-west-2  | aws-eu-west-2  | Used to set the Elastic Cloud region for the AWS deployment  |
| elastic_deployment_name  | AWS Observe and Protect  | AWS Observe and Protect  | Used to define the name for the Elastic deployment  |

#### For Google Cloud
More Google CLoud configuration remarks you can find in the [Google Cloud](GoogleCloud) folder.

Minimal config:
```json
{
	"google_cloud_project" : "<Google Project>",
	"google_cloud_service_account_path" : "/path/to/service/account/file"
}
```

List of other optional parameters that can be added to terraform.tfvars.json 
| Parameter Name  | Default value | Example | Description |
| ------------- | ------------- | ------------- | ------------- |
| elastic_version  | latest  | 8.4.1  | Used to define the Elastic Search version  |
| elastic_region  | gcp-europe-west3  | gcp-europe-west3  | Used to set the Elastic Cloud region for the Google Cloud deployment  |
| elastic_deployment_name  | Google Cloud Observe and Protect  | Google Cloud Observe and Protect  | Used to define the name for the Elastic deployment  |
| google_cloud_region  | europe-west3  | europe-west3  | Used to change the region where the Google Cloud objects getting installed  |
| google_cloud_network  | default | my-network  | Used to change the network the Elastic Agent VM is installed in. (Network needs to be existent)  |

## Deploy

For the  setup you need to init and apply the terraform configuration in the [Multi Cloud](MultiCloud) root module and start in the terraform folder. Before the apply you need to provide credentials for Elastic Cloud as well as for every Cloud Provider that you want to deploy. Terraform needs access to perform actions in your name.

After you prepared the settings for each cloud provider you've choosen you should be able to execute the deployment process.

### All in one aka Multi Cloud

If you prefer you install everything at once you need to configure all Cloud Providers. This is the default configuration. 

### Each example separately

To install each setup independenly from each other you can disable the creation of the unnecessary clusters also within the [Multi Cloud](MultiCloud) folder. Each module can run on its own. 
If you want to add more environments later you just need to change the configuration.


List of parameters to de/activate one or more cloud provider environments completly:
| Parameter Name  | Default value | Example | Description |
| ------------- | ------------- | ------------- | ------------- |
| deploy_gc  | true  | false  | Used to de/activate the Google Cloud Environment  |
| deploy_aws  | true  | false  | Used to de/activate the AWS Environment   |


### Run terraform

#### Initialize within 'terraform' folder in the Multi Cloud module

```bash
terraform init
```

#### Check plan to see what will be created by terraform

```bash
terraform plan -var-file="../local_env/terraform.tfvars.json"
```

#### Run with auto-approve will install everything

First run:
```bash
terraform apply -var-file="../local_env/terraform.tfvars.json" -auto-approve
```

The replace part is necessary if you deploy the AWS environment. Without that the Cloud Formation template that is used usually have issues on re apply 
```bash
terraform apply -var-file="../local_env/terraform.tfvars.json" -replace module.aws_environment[0].aws_serverlessapplicationrepository_cloudformation_stack.esf_cf_stack -auto-approve
```

#### Cleanup (Deletes every component that was created by terraform)

```bash
terraform destroy -var-file="../local_env/terraform.tfvars.json" -auto-approve
```

# More Elasticsearch terraform examples

Other terraform + elastic examples can be found here:
- [Patent Search](https://github.com/MarxDimitri/solution-accelerators/tree/main/patent-search) using Google Cloud BigQuery public dataset

Kibana Dashboards and other Elastic extensions can be found here
- [Elastic Content Share](https://elastic-content-share.eu/)
- [AWS Cloudformation template](https://elastic-content-share.eu/blog/how-to-create-elastic-cloud-cluster-via-aws-cloud-formation-template/)

 
