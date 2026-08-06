resource "aws_s3_bucket_logging" "elastic" {
  bucket = aws_s3_bucket.elastic_bucket.id

  target_bucket = aws_s3_bucket.elastic_bucket.id
  target_prefix = "AWSLogs/${data.aws_caller_identity.current.account_id}/s3/"
}

# -------------------------------------------------------------
# Event trigger
# -------------------------------------------------------------

resource "aws_sqs_queue" "s3-events" {
  name = "s3-s3-event-notification-queue"
  visibility_timeout_seconds = 900
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sqs:SendMessage",
      "Resource": "arn:aws:sqs:*:*:s3-s3-event-notification-queue",
      "Condition": {
        "ArnEquals": { "aws:SourceArn": "${aws_s3_bucket.elastic_bucket.arn}" }
      }
    }
  ]
}
POLICY
}