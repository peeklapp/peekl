include "restart_nginx" {
  path = "restart_nginx.hcl"
}

resource "builtin.pkg" "install_package_nginx" {
  parameters = {
    names = ["nginx"]
  }
}
