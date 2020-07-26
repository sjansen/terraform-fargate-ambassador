data "aws_iam_policy_document" "ecs-tasks" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}


# ecsTaskExecutionRole
resource "aws_iam_role" "ecsTaskExecutionRole" {
  name               = "ecsTaskExecutionRole"
  assume_role_policy = data.aws_iam_policy_document.ecs-tasks.json
}

resource "aws_iam_role_policy_attachment" "ecsTaskExecutionRole" {
  role       = aws_iam_role.ecsTaskExecutionRole.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}


# task_role
resource "aws_iam_role" "task_role" {
  name               = "${var.ecs_name}-task"
  assume_role_policy = data.aws_iam_policy_document.ecs-tasks.json
}

resource "aws_iam_role_policy" "task_role-sqs" {
  name   = "sqs-access"
  role   = aws_iam_role.task_role.id
  policy = data.aws_iam_policy_document.task_role-sqs.json
}

data "aws_iam_policy_document" "task_role-sqs" {
  statement {
    actions = [
      "sqs:DeleteMessage",
      "sqs:GetQueueUrl",
      "sqs:DeleteMessageBatch",
      "sqs:ReceiveMessage"
    ]
    resources = [
      aws_sqs_queue.queue.arn
    ]
  }
}
