data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}


# execution_role
resource "aws_iam_role" "execution_role" {
  name               = var.ecs_name
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "execution_role" {
  role       = aws_iam_role.execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}


# task_role
resource "aws_iam_role" "task_role" {
  name               = "${var.ecs_name}-task"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
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
