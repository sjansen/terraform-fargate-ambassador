resource "aws_appautoscaling_target" "app" {
  max_capacity       = 10
  min_capacity       = var.desired_count
  resource_id        = "service/${aws_ecs_cluster.app.name}/${aws_ecs_service.app.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "app" {
  name               = var.ecs_name
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.app.resource_id
  scalable_dimension = aws_appautoscaling_target.app.scalable_dimension
  service_namespace  = aws_appautoscaling_target.app.service_namespace

  target_tracking_scaling_policy_configuration {
    target_value       = var.autoscale_target
    disable_scale_in   = false
    scale_in_cooldown  = 180
    scale_out_cooldown = 120
    customized_metric_specification {
      namespace   = "AWS/SQS"
      metric_name = "ApproximateNumberOfMessagesVisible"
      statistic   = "Average"
      unit        = "Count"
      dimensions {
        name  = "QueueName"
        value = aws_sqs_queue.queue.name
      }
    }
  }
}
