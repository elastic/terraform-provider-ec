# AWS environment

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


#### Create AWS Access credentials

1. Visit the [IAM Management Console](https://us-east-1.console.aws.amazon.com/iam/home) in AWS
2. Navigate to the user you want to use for the setup
3. Click on "Security credentials"
4. Click on "Create access key" and save the credentials in your `terraform.tfvars.json` file

Hint: The credentials you choose here will also be used to authenticate the Elastic Agent against your AWS Environment. In production ready setups you might want to change that. Elastic also offers other authentication mechanisms for the Elastic Agent. This terraform script does not ATM.


 
