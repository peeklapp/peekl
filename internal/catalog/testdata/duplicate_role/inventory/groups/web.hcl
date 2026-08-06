name  = "web"
roles = [
  "nginx"
]
tags  = [
  "tag_from_group"
]

resource "builtin.pkg" "install_package_cowsay" {
  parameters = {
    names = [
      "cowsay"
    ]
  }
}
