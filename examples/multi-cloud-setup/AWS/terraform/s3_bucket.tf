resource "aws_s3_bucket" "elastic_bucket" {
  bucket = var.bucket_name

  tags = {
    Name        = "Elastic SAR Data"
  }
}

# resource "aws_s3_bucket_acl" "private_access" {
#   bucket = aws_s3_bucket.elastic_bucket.id
#   acl    = "private"
# }

resource "aws_s3_bucket_notification" "all_notifications" {
  bucket = aws_s3_bucket.elastic_bucket.id
  queue {
    queue_arn     = aws_sqs_queue.cloudtrail-events.arn
    events        = ["s3:ObjectCreated:*"]
    filter_prefix       = "AWSLogs/${data.aws_caller_identity.current.account_id}/CloudTrail/"
  }

  queue {
    queue_arn     = aws_sqs_queue.vpc-events.arn
    events        = ["s3:ObjectCreated:*"]
    filter_prefix       = "AWSLogs/${data.aws_caller_identity.current.account_id}/vpcflowlogs/"
  }

  queue {
    queue_arn     = aws_sqs_queue.elb-events.arn
    events        = ["s3:ObjectCreated:*"]
    filter_prefix       = "AWSLogs/${data.aws_caller_identity.current.account_id}/elasticloadbalancing/"
  }

  queue {
    queue_arn     = aws_sqs_queue.s3-events.arn
    events        = ["s3:ObjectCreated:*"]
    filter_prefix       = "AWSLogs/${data.aws_caller_identity.current.account_id}/s3/"
  }
}

resource "aws_s3_bucket_policy" "cloudtrail" {
  bucket = aws_s3_bucket.elastic_bucket.id
  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AWSCloudTrailAclCheck",
            "Effect": "Allow",
            "Principal": {
              "Service": "cloudtrail.amazonaws.com"
            },
            "Action": "s3:GetBucketAcl",
            "Resource": "${aws_s3_bucket.elastic_bucket.arn}"
        },
        {
            "Sid": "AWSLogDeliveryAclCheck",
            "Effect": "Allow",
            "Principal": {
                "Service": "delivery.logs.amazonaws.com"
            },
            "Action": "s3:GetBucketAcl",
            "Resource": "${aws_s3_bucket.elastic_bucket.arn}",
            "Condition": {
                "StringEquals": {
                    "aws:SourceAccount": ${data.aws_caller_identity.current.account_id}
                }
            }
        },
        {
            "Sid": "AWSLogWrite",
            "Effect": "Allow",
            "Principal": {
              "Service": [
                "cloudtrail.amazonaws.com",
                "vpc-flow-logs.amazonaws.com",
                "delivery.logs.amazonaws.com"
              ]
            },
            "Action": "s3:PutObject",
            "Resource": "${aws_s3_bucket.elastic_bucket.arn}/AWSLogs/${data.aws_caller_identity.current.account_id}/*",
            "Condition": {
                "StringEquals": {
                    "s3:x-amz-acl": "bucket-owner-full-control"
                }
            }
        },
        {
            "Sid": "S3PolicyStmt-DO-NOT-MODIFY-1664972344369",
            "Effect": "Allow",
            "Principal": {
                "Service": "logging.s3.amazonaws.com"
            },
            "Action": "s3:PutObject",
            "Resource": "arn:aws:s3:::elastic-sar-bucket/*"
        }
    ]
}
POLICY
}