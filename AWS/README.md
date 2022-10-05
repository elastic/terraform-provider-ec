# elastic-terraform-examples

This example is creating an AWS Monitoring and Enhanced Security environment. It creates all necessary AWS Services as well as the Elastic Cloud Cluster for you. The only thing you need to provide is are AWS account credentials that provide the right permissions as well as the Elastic Cloud API Key. It works both: In [Elastic Cloud directly](https://cloud.elastic.co) or via the [AWS Marketplace option for Elastic Cloud](https://ela.st/aws).

This example will install and configure:
- Elastic Cluster
- AWS EC2 instance with Elastic Agent installed and configured to talk to the Elastic Cluster 
- Elastic Agent will be configured to collect available Metric datasets with zero manual configuration
- The Elastic Cluster will be configured with the following additional capabilities
	- Preloaded all Elastic Security Detection rules and enabled all AWS related rules

## Get started

#### Prepare software dependencies

- [jq](https://stedolan.github.io/jq/download/)
- [terraform](https://www.terraform.io/downloads)


#### Clone the repository

```bash
git clone https://github.com/felix-lessoer/elastic-terraform-examples.git
```

#### Create local env files within the repo

```bash
mkdir local_env
touch terraform.tfvars.json
```

Modify the terraform env settings. The service account explanation you find below

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


#### Create Elastic Cloud ID following this steps.

[Create EC API key](https://registry.terraform.io/providers/elastic/ec/latest/docs#api-key-authentication-recommended)

Set env variable for Elastic Cloud:

```bash
export EC_API_KEY="[PUT YOUR ELASTIC CLOUD API KEY HERE]"
```

#### Create AWS Access credentials

1. Visit the [IAM Management Console](https://us-east-1.console.aws.amazon.com/iam/home) in AWS
2. Navigate to the user you want to use for the setup
3. Click on "Security credentials"
4. Click on "Create access key" and save the credentials in your `terraform.tfvars.json` file

Hint: The credentials you choose here will also be used to authenticate the Elastic Agent against your AWS Environment. In production ready setups you might want to change that. Elastic also offers other authentication mechanisms for the Elastic Agent. This terraform script does not ATM.

### Deploy

##### Initialize within 'terraform' folder

```bash
terraform init
```

##### Check plan

```bash
terraform plan -var-file="../local_env/terraform.tfvars.json"
```

##### Run

```bash
terraform apply -var-file="../local_env/terraform.tfvars.json" -auto-approve
```

### Cleanup

```bash
terraform destroy -var-file="../local_env/terraform.tfvars.json" -auto-approve
```

 
