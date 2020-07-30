output "ambassador_repo_name" {
  value = aws_ecr_repository.ambassador.name
}

output "ambassador_repo_url" {
  value = aws_ecr_repository.ambassador.repository_url
}

output "application_repo_name" {
  value = aws_ecr_repository.application.name
}

output "application_repo_url" {
  value = aws_ecr_repository.application.repository_url
}

output "registry" {
  value = split("/", aws_ecr_repository.ambassador.repository_url)[0]
}

output "queue_url" {
  value = aws_sqs_queue.queue.id
}
