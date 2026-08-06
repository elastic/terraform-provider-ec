# resource "aws_cloudformation_stack" "elastic" {
#   name = "elastic-stack"
#   template_url = "https://mp-saas-integrations.s3.amazonaws.com/saas-elastic-cloud/main/templates/AgentInstallMain.yaml"
#   capabilities = ["CAPABILITY_NAMED_IAM"]
#   parameters = {
#     DeploymentName = var.elastic_aws_deployment_name
#     EC2HostName = "elastic-agent"
#     EC2InstanceType = "t2.micro"
#     GitBranchName = "main"
#     KeyPairName = "felix-putty"
#     PublicSubnet1ID = "subnet-00f145e4ab29d2d0c"
#     PublicSubnet2ID = "subnet-00f145e4ab29d2d0c"
#     QSS3BucketName = "mp-saas-integrations"
#     QSS3KeyPrefix = "saas-elastic-cloud/"
#     Region = "us-east-1"
#     RemoteAccessCIDR = "172.31.0.0/16"
#     RootVolumeSize = "10"
#     SecretName = "arn:aws:secretsmanager:us-east-1:644184947617:secret:ec_pme_dev_api_key-UCoLzg"
#     VPCID = "vpc-0a8d055883d1a19ac"
#   }

# }