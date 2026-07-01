# -------------------------------------------------------------
# Data Collection
# -------------------------------------------------------------

data "aws_vpcs" "all" {}

resource "aws_flow_log" "vpc" { 
  for_each             = toset(data.aws_vpcs.all.ids)

  iam_role_arn         = aws_iam_role.vpcflow.arn
  log_destination      = aws_s3_bucket.elastic_bucket.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  log_format           = "$${version} $${account-id} $${interface-id} $${srcaddr} $${dstaddr} $${srcport} $${dstport} $${protocol} $${packets} $${bytes} $${start} $${end} $${action} $${log-status} $${vpc-id} $${subnet-id} $${instance-id} $${tcp-flags} $${type} $${pkt-srcaddr} $${pkt-dstaddr} $${region} $${az-id} $${sublocation-type} $${sublocation-id} $${pkt-src-aws-service} $${pkt-dst-aws-service} $${flow-direction} $${traffic-path}"
  
  vpc_id               = each.value
}

# -------------------------------------------------------------
# Role assignment
# -------------------------------------------------------------

resource "aws_iam_role" "vpcflow" {
  name = "vpcflow_writer"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "vpc-flow-logs.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "vpcflow" {
  name = "vpcflow_writer"
  role = aws_iam_role.vpcflow.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogDelivery",
        "logs:DeleteLogDelivery"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

# -------------------------------------------------------------
# Event trigger
# -------------------------------------------------------------

resource "aws_sqs_queue" "vpc-events" {
  name = "s3-vpc-event-notification-queue"
  visibility_timeout_seconds = 900
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sqs:SendMessage",
      "Resource": "arn:aws:sqs:*:*:s3-vpc-event-notification-queue",
      "Condition": {
        "ArnEquals": { "aws:SourceArn": "${aws_s3_bucket.elastic_bucket.arn}" }
      }
    }
  ]
}
POLICY
}