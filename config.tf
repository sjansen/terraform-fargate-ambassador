variable "app_url" {
  type    = string
  default = "http://127.0.0.1:8080"
}

variable "autoscale_target" {
  type    = number
  default = 0.75
}

variable "aws_region" {
  type    = string
  default = "us-east-2"
}

variable "debug" {
  type    = bool
  default = false
}

variable "desired_count" {
  type    = number
  default = 0
}

variable "ecr_prefix" {
  type    = string
  default = "terraform-fargate-ambassador"
}

variable "ecs_name" {
  type    = string
  default = "terraform-fargate-ambassador"
}

variable "queue_name" {
  type    = string
  default = "terraform-fargate-ambassador"
}

variable "vpc_name" {
  type    = string
  default = "terraform-fargate-ambassador"
}
