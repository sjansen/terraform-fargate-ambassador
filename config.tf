variable "aws_region" {
  type = string
  default = "us-east-2"
}

variable "ecr_prefix" {
  type = string
  default = "terraform-fargate-ambassador"
}

variable "queue_name" {
  type = string
  default = "terraform-fargate-ambassador"
}
