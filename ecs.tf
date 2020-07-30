resource "aws_ecs_cluster" "app" {
  name               = var.ecs_name
  capacity_providers = ["FARGATE", "FARGATE_SPOT"]
  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    base              = 1
    weight            = 1
  }
  default_capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    base              = 0
    weight            = 4
  }
  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

resource "aws_ecs_service" "app" {
  name                 = var.ecs_name
  launch_type          = "FARGATE"
  platform_version     = "1.4.0"
  cluster              = aws_ecs_cluster.app.id
  task_definition      = aws_ecs_task_definition.app.arn
  force_new_deployment = true
  desired_count        = var.desired_count
  network_configuration {
    assign_public_ip = false
    security_groups = [
      aws_security_group.egress-only.id,
    ]
    subnets = [
      aws_subnet.private.id,
    ]
  }
  depends_on = [aws_cloudwatch_log_group.app]
  lifecycle {
    ignore_changes = [desired_count]
  }
}

resource "aws_ecs_task_definition" "app" {
  family                   = var.ecs_name
  requires_compatibilities = ["FARGATE"]
  execution_role_arn       = aws_iam_role.ecsTaskExecutionRole.arn
  task_role_arn            = aws_iam_role.task_role.arn
  cpu                      = 256
  memory                   = 512
  network_mode             = "awsvpc"
  container_definitions    = <<EOF
[{
    "name": "ambassador",
    "dependsOn": [{
      "containerName": "application",
      "condition": "START"
    }],
    "environment": [
      {"name": "AMBASSADOR", "value": "${var.ambassador_url}"},
      {"name": "APPLICATION", "value": "${var.application_url}"},
      {"name": "DEBUG", "value": "${var.debug ? "enabled" : ""}"},
      {"name": "QUEUE", "value": "${var.queue_name}"}
    ],
    "essential": true,
    "healthCheck": {
      "command": ["CMD", "./main", "check-health"],
      "interval": 60,
      "timeout": 5,
      "retries": 3,
      "startPeriod": null
    },
    "image": "${aws_ecr_repository.ambassador.repository_url}:latest",
    "logConfiguration": {
      "logDriver": "awslogs",
      "options": {
        "awslogs-region": "${var.aws_region}",
        "awslogs-group": "/ecs/${var.ecs_name}",
        "awslogs-stream-prefix": "ecs",
        "awslogs-datetime-format": "%Y-%m-%dT%H:%M:%S%LZ"
      }
    },
    "readonlyRootFilesystem": true,
    "startTimeout": 120,
    "stopTimeout": 120
}, {
    "name": "application",
    "environment": [
      {"name": "AMBASSADOR", "value": "${var.ambassador_url}"},
      {"name": "DEBUG", "value": "${var.debug ? "enabled" : ""}"}
    ],
    "essential": true,
    "healthCheck": {
      "command": ["CMD", "./main", "check-health"],
      "interval": 60,
      "timeout": 5,
      "retries": 3,
      "startPeriod": null
    },
    "image": "${aws_ecr_repository.application.repository_url}:latest",
    "logConfiguration": {
      "logDriver": "awslogs",
      "options": {
        "awslogs-region": "${var.aws_region}",
        "awslogs-group": "/ecs/${var.ecs_name}",
        "awslogs-stream-prefix": "ecs"
      }
    },
    "readonlyRootFilesystem": true,
    "startTimeout": 120,
    "stopTimeout": 120
}]
EOF
}
