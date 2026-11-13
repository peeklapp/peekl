group "web" {
  tags = ["tag_from_group"]

  variables = {
    listen_port = 80
  }

  role "nginx" {
    variables = {
      worker_processes = 4
      server_names     = ["example.com", "www.example.com"]
      gzip             = true
    }
  }
}
