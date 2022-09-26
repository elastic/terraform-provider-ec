# elastic-terraform-examples

The first example in the repository is creating a Google Cloud Monitoring and Enhanced Security environment. It creates all necessary Google Cloud Services as well as the Elastic Cloud Cluster for you. The only thing you need to provide is an appropriate Google Cloud Service account that has the right permissions and the Elastic Cloud API Key. It works both: In Elastic Cloud directly or via the Google Cloud Marketplace option for Elastic Cloud.

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

## Prepare

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
	"google_cloud_project" : "<Google Project>",
	"google_cloud_service_account_path" : "/path/to/service/account/file"
}
```

List of other optional parameters that can be added to terraform.tfvars.json 
| Parameter Name  | Default value | Example | Description |
| ------------- | ------------- | ------------- | ------------- |
| elastic_version  | latest  | 8.4.1  | Used to define the Elastic Search version  |
| elastic_gc_region  | gcp-europe-west3  | gcp-europe-west3  | Used to set the Elastic Cloud region for the Google Cloud deployment  |
| elastic_gc_deployment_name  | Google Cloud Observe and Protect  | Google Cloud Observe and Protect  | Used to define the name for the Elastic deployment  |
| google_cloud_region  | europe-west3  | europe-west3  | Used to change the region where the Google Cloud objects getting installed  |
| google_cloud_network  | default | my-network  | Used to change the network the Elastic Agent VM is installed in. (Network needs to be existent)  |


#### Create Elastic Cloud ID following this steps.

[Create EC API key](https://registry.terraform.io/providers/elastic/ec/latest/docs#api-key-authentication-recommended)

- Set env variable for Elastic Cloud:

```bash
export EC_API_KEY="[PUT YOUR ELASTIC CLOUD API KEY HERE]"
```


#### Create Google Cloud service account following this steps.

##### Create json for Google Cloud credentials. Follow the instractions here

Use [Google Cloud Console](https://console.cloud.google.com/iam-admin/serviceaccounts) for the initial creation


##### Set permission for the Google Cloud service account
We are using this service also to connect the Elastic Agent to your Google Cloud Project.
Because of that you should also take care that your Service Account is following the Elastic Agent Integration docs.
Meaning the service account need to have the following roles as well as the roles for creating the terraformed services

- Elastic Agent integration roles needed
	- pubsub.subscriptions.consume
	- pubsub.subscriptions.create 
	- pubsub.subscriptions.get
	- pubsub.topics.attachSubscription

- Terraform installation roles need
	- resourcemanager.projectIamAdmin
	- roles/compute.instanceAdmin.v1 (To create compute instances)
	- roles/logging.admin (To create log sinks)
	- pubsub.editor (This one usually includes the roles the Elastic Agent needs)
	
Example roles assignment via `gcloud`

```bash
gcloud projects add-iam-policy-binding "[PUT YOUR GOOGLE CLOUD PROJECT NAME HERE]" \
--member=serviceAccount:[PUT YOUR SERVICE ACCOUNT MEMBER HERE] \
--role=roles/[PUT THE ROLE NAME IN HERE]
```

Example

```bash
gcloud projects add-iam-policy-binding "my-project-name" \
--member=serviceAccount:terraform@elastic-product.iam.gserviceaccount.com \
--role=roles/pubsub.editor
```

- Verify permissions
```bash
gcloud projects get-iam-policy "[PUT YOUR GOOGLE CLOUD PROJECT NAME HERE]" \
--flatten="bindings[].members" \
--format='table(bindings.role)' \
--filter="bindings.members:[PUT YOUR SERVICE ACCOUNT MEMBER HERE]"
```

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

 
