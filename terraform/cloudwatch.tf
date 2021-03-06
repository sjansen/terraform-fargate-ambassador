resource "aws_cloudwatch_log_group" "app" {
  name              = "/ecs/${var.ecs_name}"
  retention_in_days = 30
}

resource "aws_cloudwatch_log_group" "containerinsights" {
  name              = "/aws/ecs/containerinsights/${var.ecs_name}/performance"
  retention_in_days = 1
}
