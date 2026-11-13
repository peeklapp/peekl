include "testing" {
  path = "testing.hcl"
}

resource "builtin.pkg" "install_package_nginx" {
  parameters = {
    names = ["nginx"]
  }
}
