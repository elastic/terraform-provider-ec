# elastic-terraform-examples

The project in this repository is creating an Elastic Cloud environment in order to getting started with monitoring and protecting your Cloud Service Providers(CSP) environment in Google, AWS and/or Azure. It is creating all necessary components within the CSPs as well as the in Elastic Cloud using terraform. The whole process will be done in less than 1h. 

You can either install every Cloud Environment separatly or choose the MultiCloud project to install everything at once. By choosing MultiCloud the terraform script will also configure the necessary connection between the clusters in order to do Cross Cluster Search(CSS). Because of that each cluster can live in its own Cloud Provider environment (GCP cluster in GCP, AWS cluser in AWS and so on). This will guarantee a low cost footprint when collecting the relevant data from the providers. But because of CCS every cluster can get queried by one main cluster. 

## Getting started

In order to set up your Multi Cloud environment you need to configure access for each Cloud Provider first. The best way to do that is following the instructions in the Cloud Specific modules. But do not deploy on the Cloud Provider level but rather in this folder.

After you prepared the settings for each cloud provider you should be able to execute the deployment process.

### Deploy

##### Initialize within 'terraform' folder

```bash
terraform init
```

##### Check plan

```bash
terraform plan -var-file="../../AWS/local_env/terraform.tfvars.json" -var-file="../../GoogleCloud/local_env/terraform.tfvars.json"
```

##### Run

```bash
terraform apply -var-file="../../AWS/local_env/terraform.tfvars.json" -var-file="../../GoogleCloud/local_env/terraform.tfvars.json" -auto-approve
```

### Cleanup

```bash
terraform destroy -var-file="../../AWS/local_env/terraform.tfvars.json" -var-file="../../GoogleCloud/local_env/terraform.tfvars.json" -auto-approve
```
