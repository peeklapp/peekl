# certs

This module is meant to hold everything related to the creation of the certificate authority within Peekl, as well as the certificates.

## SSL folder structure

On both the agent and the server, a SSL directory will exist. The goal of this SSL directory is to hold everything about SSL.

### Server SSL folder

This folder is located here : `/etc/peekl/server/ssl`

And it's content is a follow

| Name | Type | Description |
| ---- | ---- | ----------- |
| ca | directory | Contains the CA certificate, the CA private key as well as the CRL |
| signed | directory | A directory that contains all the signed certificates |
| pending | directory | A directory that contains all the certificate that still needs to be signed |

### Agent SSL folder

This folder is located here : `/etc/peekl/agent/ssl`

| Name | Type | Description |
| ---- | ---- | ----------- |
| ca | directory | Contains the CA certificate as well as the CRL |
| certificate | directory | A directory that contains the signed certificate of the node, as well as it's private key |



