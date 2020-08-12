module "app" {
  source = "./terraform"

  aws_region = var.aws_region
  debug      = var.debug
  fill_disk  = var.fill_disk
}
