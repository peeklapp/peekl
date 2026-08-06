name   = "dummy"
groups = [
  "web"
]
roles  = [
  "test",
  "nginx"
]
tags   = [
  "tag_from_node"
]

# Resources
resource "builtin.pkg" "install_package_htop" {
  parameters = {
    names = [
      "htop"
    ]
  }
}
