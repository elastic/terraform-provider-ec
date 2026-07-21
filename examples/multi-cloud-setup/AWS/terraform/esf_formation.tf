resource "aws_s3_object" "config_file" {
  bucket = var.bucket_name
  key    = "sar_config.yaml"
  content = templatefile("${path.module}/config.tftpl", 
  {
    s3-sqs-objs = [
      {arn= aws_sqs_queue.vpc-events.arn, datastream = "logs-aws.vpcflow-esf"},
      {arn= aws_sqs_queue.cloudtrail-events.arn, datastream = "logs-aws.cloudtrail-esf"},
      {arn= aws_sqs_queue.s3-events.arn, datastream = "logs-aws.s3access-esf"},
      {arn= aws_sqs_queue.elb-events.arn, datastream = "logs-aws.elb-esf"}
    ]
    cw-logs-objs = data.aws_cloudwatch_log_groups.all.arns
    elasticsearch_url  = ec_deployment.elastic_deployment.elasticsearch[0].https_endpoint
    elasticsearch_user  = ec_deployment.elastic_deployment.elasticsearch_username
    elasticsearch_password  = ec_deployment.elastic_deployment.elasticsearch_password
  }
  )
}

data "aws_serverlessapplicationrepository_application" "esf_sar" {
  application_id = "arn:aws:serverlessrepo:eu-central-1:267093732750:applications/elastic-serverless-forwarder"
}

resource "aws_serverlessapplicationrepository_cloudformation_stack" "esf_cf_stack" {
  name             = "terraform-elastic-serverless-forwarder"
  application_id   = data.aws_serverlessapplicationrepository_application.esf_sar.application_id
  semantic_version = data.aws_serverlessapplicationrepository_application.esf_sar.semantic_version
  capabilities     = data.aws_serverlessapplicationrepository_application.esf_sar.required_capabilities

parameters = {
    ElasticServerlessForwarderS3ConfigFile         = "s3://${var.bucket_name}/sar_config.yaml"  ## FILL WITH THE VALUE OF THE S3 URL IN THE FORMAT "s3://bucket-name/config-file-name" POINTING TO THE CONFIGURATION FILE FOR YOUR DEPLOYMENT OF THE ELASTIC SERVERLESS FORWARDER

    ElasticServerlessForwarderSSMSecrets           = ""  ## FILL WITH A COMMA DELIMITED LIST OF AWS SSM SECRETS ARNS REFERENCED IN THE CONFIG YAML FILE (IF ANY).

    ElasticServerlessForwarderKMSKeys              = ""  ## FILL WITH A COMMA DELIMITED LIST OF AWS KMS KEYS ARNS TO BE USED FOR DECRYPTING AWS SSM SECRETS REFERENCED IN THE CONFIG YAML FILE (IF ANY).

    ElasticServerlessForwarderSQSEvents            = ""  ## FILL WITH A COMMA DELIMITED LIST OF DIRECT SQS QUEUES ARNS TO SET AS EVENT TRIGGERS FOR THE LAMBDA (IF ANY).

    ElasticServerlessForwarderS3SQSEvents          = "${aws_sqs_queue.vpc-events.arn},${aws_sqs_queue.cloudtrail-events.arn},${aws_sqs_queue.s3-events.arn},${aws_sqs_queue.elb-events.arn}"  ## FILL WITH A COMMA DELIMITED LIST OF S3 SQS EVENT NOTIFICATIONS ARNS TO SET AS EVENT TRIGGERS FOR THE LAMBDA (IF ANY).

    ElasticServerlessForwarderKinesisEvents        = ""  ## FILL WITH A COMMA DELIMITED LIST OF KINESIS DATA STREAM ARNS TO SET AS EVENT TRIGGERS FOR THE LAMBDA (IF ANY).

    ElasticServerlessForwarderCloudWatchLogsEvents = "${join(",",data.aws_cloudwatch_log_groups.all.arns)}"  ## FILL WITH A COMMA DELIMITED LIST OF CLOUDWATCH LOGS LOG GROUPS ARNS TO SET SUBSCRIPTION FILTERS ON THE LAMBDA FOR (IF ANY).

    ElasticServerlessForwarderS3Buckets            = "${aws_s3_bucket.elastic_bucket.arn}"  ## FILL WITH A COMMA DELIMITED LIST OF S3 BUCKETS ARNS THAT ARE THE SOURCES OF THE S3 SQS EVENT NOTIFICATIONS (IF ANY).
  }

  depends_on = [
    data.aws_cloudwatch_log_groups.all
  ]
}