data "aws_caller_identity" "current" {}

resource "aws_cloudtrail" "management" {
  name                          = "tf-trail-elastic"
  s3_bucket_name                = aws_s3_bucket.elastic_bucket.id
  include_global_service_events = true
  is_multi_region_trail         = true
  enable_logging                = true

  event_selector {
    read_write_type           = "All"
    include_management_events = true

    data_resource {
      type   = "AWS::Lambda::Function"
      values = ["arn:aws:lambda"]
    }

    data_resource {
      type   = "AWS::S3::Object"
      values = ["arn:aws:s3"]
    }
  }

  depends_on = [aws_s3_bucket_policy.cloudtrail]
}

# -------------------------------------------------------------
# Event trigger
# -------------------------------------------------------------

resource "aws_sqs_queue" "cloudtrail-events" {
  name = "s3-cloudtrail-event-notification-queue"
  visibility_timeout_seconds = 900
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sqs:SendMessage",
      "Resource": "arn:aws:sqs:*:*:s3-cloudtrail-event-notification-queue",
      "Condition": {
        "ArnEquals": { "aws:SourceArn": "${aws_s3_bucket.elastic_bucket.arn}" }
      }
    }
  ]
}
POLICY
}