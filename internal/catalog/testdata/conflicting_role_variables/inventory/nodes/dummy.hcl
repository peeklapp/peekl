node "dummy" {
  groups = [
    "web"
  ]

  role "nginx" {
    variables = {
      port = 80
    }
  }
}
