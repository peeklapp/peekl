resource "builtin.systemd_service" "restart_nginx" {
  parameters = {
    name  = "nginx"
    state = "restarted"
  }
}
