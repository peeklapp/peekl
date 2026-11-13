group "web" {
  role "nginx" {
    variables = {
      port = 8080
    }
  }
}
