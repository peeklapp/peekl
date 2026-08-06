name = "agent"
tags = [
  "public"
]
groups = [
  "common",
  "webserver"
]
roles = [
  "firewall",
  "my_app"
]

# Resources
resource "builtin.pkg" "extra_packages" {
  present    = true
  parameters = {
    names    = [
      "htop",
      "ncdu"
    ]
  } 
}
