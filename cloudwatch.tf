resource "aws_cloudwatch_log_group" "app" {
  name              = "/ecs/${var.ecs_name}"
  retention_in_days = 30
}
