resource "aws_ecr_repository" "ambassador" {
  name                 = var.ecr_prefix == "" ? "ambassador" : "${var.ecr_prefix}-ambassador"
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "application" {
  name                 = var.ecr_prefix == "" ? "application" : "${var.ecr_prefix}-application"
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_lifecycle_policy" "for" {
  for_each = {
    ambassador = aws_ecr_repository.ambassador.name
    application = aws_ecr_repository.application.name
  }

  repository = each.value
  policy = <<EOF
{
    "rules": [
        {
            "rulePriority": 10,
            "description": "Keep last 3 dev images",
            "selection": {
                "tagStatus": "tagged",
                "tagPrefixList": ["dev-"],
                "countType": "imageCountMoreThan",
                "countNumber": 10
            },
            "action": {
                "type": "expire"
            }
        },
        {
            "rulePriority": 20,
            "description": "Keep last 3 stg images",
            "selection": {
                "tagStatus": "tagged",
                "tagPrefixList": ["stg-"],
                "countType": "imageCountMoreThan",
                "countNumber": 3
            },
            "action": {
                "type": "expire"
            }
        },
        {
            "rulePriority": 30,
            "description": "Keep last 3 prod images",
            "selection": {
                "tagStatus": "tagged",
                "tagPrefixList": ["prod-"],
                "countType": "imageCountMoreThan",
                "countNumber": 3
            },
            "action": {
                "type": "expire"
            }
        },
        {
            "rulePriority": 500,
            "description": "Expire untagged images older than 3 days",
            "selection": {
                "tagStatus": "untagged",
                "countType": "sinceImagePushed",
                "countUnit": "days",
                "countNumber": 3
            },
            "action": {
                "type": "expire"
            }
        },
        {
            "rulePriority": 1000,
            "description": "Expire images with unrecognized tags older than 3 days",
            "selection": {
                "tagStatus": "any",
                "countType": "sinceImagePushed",
                "countUnit": "days",
                "countNumber": 3
            },
            "action": {
                "type": "expire"
            }
        }
    ]
}
EOF
}
