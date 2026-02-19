# peekl

> [!CAUTION]
> This project is still under active development. Expect breaking changes at any time. Use on your servers at your own risk.

Peekl is a simple yet powerful configuration management tool ! It's a more modern and highly opinionated solution at how tools such as Ansible and Puppet works.

1. [Getting started](#getting-started)
	1. [Installing the server](#installing-the-server)
	2. [Installing the agent](#installing-the-agent)
2. [Contributing](#contributing)
3. [License](#license)

## Getting started

Peekl is composed of two main components :
- The agent, that you install on each node that you want to manage.
- The server, that the agent will communicate with to get their catalog.

### Installing the server

The first step you will have to follow is to install the server. For that you can download the latest version of Peekl on GitHub.

Installing from Debian packages
```bash
export PEEKL_VERSION=0.1.0
export PEEKL_ARCH=amd64
wget https://github.com/peeklapp/peekl/releases/download/${PEEKL_VERSION}/peekl-server_${PEEKL_VERSION}_linux_${PEEKL_ARCH}.deb
apt install ./peekl-server_${PEEKL_VERSION}_linux_${PEEKL_ARCH}.deb
```

Once the server is installed, you'll have to bootstrap it.
```bash
peekl-server bootstrap
```

And then you're ready to start the server. If you've installed it using the Debian packages, you can start it with systemd.

```bash
systemctl enable peekl-server.service && systemctl start peekl-server.service
```

### Installing the agent

You can repeat most of the same step that you did for the server

Install using Debian packages
```bash
export PEEKL_VERSION=0.1.0
export PEEKL_ARCH=amd64
wget https://github.com/peeklapp/peekl/releases/download/${PEEKL_VERSION}/peekl-agent_${PEEKL_VERSION}_linux_${PEEKL_ARCH}.deb
apt install ./peekl-agent_${PEEKL_VERSION}_linux_${PEEKL_ARCH}.deb
```

Then you also have to bootstrap the agent
```bash
peekl-agent bootstrap
```

But this time you will also have to run the command to sign the certificates on the server
In order to find which certificate to sign, you will have to run the following command on the server

```bash
peekl-server ca list pending
```

And use the found name to sign the certificate
```bash
peekl-server ca sign --certname name_of_the_node
```

Then the agent should automatically download it's signed certificate, and be ready to get started

You can start the agent like we did for the server
```bash
systemctl enable peekl-agent.service && systemctl start peekl-agent.service
```

You could also execute the agent by hand
```bash
peekl-agent run
```
# Contributing

Contributions are more than welcome. But please note that supporting every little use-case is not something feasible. In order to keep the project into a maintainable state, we might refuse improvements or features suggestions, if they're deemed incompatible with the ideas of what Peekl should be : simple and efficient.

Asking for help is always ok, and reporting bug even more.

# License

This software is distributed with the [MIT License](https://opensource.org/license/mit)
