resource "builtin.pkg" "install_package_nginx" {
  parameters = {
    names = ["nginx"]
  }
}
