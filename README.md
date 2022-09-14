# elastic-terraform-examples

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

#### Create Elastic Cloud ID following this steps.

[Create EC API key](https://registry.terraform.io/providers/elastic/ec/latest/docs#api-key-authentication-recommended)

- Set env variable for Elastic Cloud:

```bash
export EC_API_KEY="[PUT YOUR ELASTIC CLOUD API KEY HERE]"
```


#### Create Google Cloud service account following this steps.

##### Create json for Google Cloud credentials. Follow the instractions here

Use [Google Cloud Console](https://console.cloud.google.com/iam-admin/serviceaccounts) for the initial creation


##### Set permission for the Google Cloud service account.

```bash
gcloud projects add-iam-policy-binding "[PUT YOUR GOOGLE CLOUD PROJECT NAME HERE]" \
--member=serviceAccount:[PUT YOUR SERVICE ACCOUNT MEMBER HERE] \
--role=roles/resourcemanager.projectIamAdmin
```

```bash
gcloud projects add-iam-policy-binding "[PUT YOUR GOOGLE CLOUD PROJECT NAME HERE]" \
--member=serviceAccount:[PUT YOUR SERVICE ACCOUNT MEMBER HERE] \
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

 
