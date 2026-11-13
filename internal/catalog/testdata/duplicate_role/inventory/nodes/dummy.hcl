node "dummy" {
  tags = [
    "tag_from_node"
  ]
  groups = [
    "web"
  ]

  role "test" {}
  role "nginx" {}
}
