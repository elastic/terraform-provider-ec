# Elastic Terraform Examples to build an Multi Cloud Monitoring environment

The project in this repository is creating an Elastic Cloud environment in order to getting started with monitoring and protecting your Cloud Service Providers(CSP) environment in Google, AWS and/or Azure. It is creating all necessary components within the CSPs as well as the in Elastic Cloud using terraform. The whole process will be done in less than 1h. 

You can either install every Cloud Environment separatly or choose the MultiCloud project to install everything at once. By choosing MultiCloud the terraform script will also configure the necessary connection between the clusters in order to do Cross Cluster Search(CSS). Because of that each cluster can live in its own Cloud Provider environment (GCP cluster in GCP, AWS cluser in AWS and so on). This will guarantee a low cost footprint when collecting the relevant data from the providers. But because of CCS every cluster can get queried by one main cluster. 

## Getting started

First of all you need to decide which setup you would like to use. Installing each needed example separatly or all in one.

### All in one

For the all in setup you need to init and apply the terraform configuration in the [MultiCloud](MultiCloud) folder. This folder also contains more description on how to prepare your local environment for that.

### Each example separately

To install each setup independenly from each other you can go into the folder of your preferred Cloud Provider. Each module can run on its own. 
If you want to switch to the all in one setup later you may need to import / destroy some of the objects that where created independenly.

## More examples

Other terraform + elastic examples can be found here:
- [Patent Search](https://github.com/MarxDimitri/solution-accelerators/tree/main/patent-search) using Google Cloud BigQuery public dataset

Kibana Dashboards and other Elastic extensions can be found here
- [Elastic Content Share](https://elastic-content-share.eu/)
- [AWS Cloudformation template](https://elastic-content-share.eu/blog/how-to-create-elastic-cloud-cluster-via-aws-cloud-formation-template/)

 
