resource "aws_sqs_queue" "queue" {
  name                       = var.queue_name
  delay_seconds              = 5
  receive_wait_time_seconds  = 20
  visibility_timeout_seconds = 600

  kms_master_key_id         = "alias/aws/sqs"
  message_retention_seconds = 1209600
}
