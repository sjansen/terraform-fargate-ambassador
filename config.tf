variable "aws_region" {
  type = string
  default = "us-east-2"
}

variable "queue_name" {
  type = string
  default = "terraform-fargate-ambassador"
}
