node "dummy" {
  groups = ["web"]
  tags   = ["tag_from_node"]

  variables = {
    environment = "production"
    replicas    = 3
  }

  role "test" {}
}
