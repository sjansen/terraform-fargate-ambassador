output "ambassador_repo_name" {
  value = module.app.ambassador_repo_name
}

output "ambassador_repo_url" {
  value = module.app.ambassador_repo_url
}

output "application_repo_name" {
  value = module.app.application_repo_name
}

output "application_repo_url" {
  value = module.app.application_repo_url
}

output "registry" {
  value = split("/", module.app.registry)[0]
}

output "queue_url" {
  value = module.app.queue_url
}
