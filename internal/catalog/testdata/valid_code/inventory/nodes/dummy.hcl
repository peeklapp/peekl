name   = "dummy"
roles  = [
  "test"
]
groups = [
  "web"
]
tags   = [
  "tag_from_node"
]

resource "builtin.pkg" "install_package_htop" {
  parameters = {
    names = ["htop"]  
  }
}
