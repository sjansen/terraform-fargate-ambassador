output "ambassador_repo_name" {
  value = aws_ecr_repository.ambassador.name
}

output "ambassador_repo_url" {
  value = aws_ecr_repository.ambassador.repository_url
}

output "ecr_registry" {
  value = split("/", aws_ecr_repository.ambassador.repository_url)[0]
}
