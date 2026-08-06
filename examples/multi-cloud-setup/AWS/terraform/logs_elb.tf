# -------------------------------------------------------------
# Event trigger
# -------------------------------------------------------------

resource "aws_sqs_queue" "elb-events" {
  name = "s3-elb-event-notification-queue"
  visibility_timeout_seconds = 900
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sqs:SendMessage",
      "Resource": "arn:aws:sqs:*:*:s3-elb-event-notification-queue",
      "Condition": {
        "ArnEquals": { "aws:SourceArn": "${aws_s3_bucket.elastic_bucket.arn}" }
      }
    }
  ]
}
POLICY
}