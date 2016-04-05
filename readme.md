Moxxi Migration Proxy Server
============================

To set this up, you need a few things:

* Debian (or your worse choice)
* `iptables`
* `nginx`
* `ngx_http_sub_module` - you will have to [custom build the module](https://serversforhackers.com/compiling-third-party-modules-into-nginx) [or try this](https://www.digitalocean.com/community/tutorials/how-to-add-ngx_pagespeed-to-nginx-on-ubuntu-14-04) [or maybe this](http://serverfault.com/questions/227480/installing-optional-nginx-modules-with-apt-get)
* `moxxi` - the binary (can be downloaded here)
* `cron` or some version of it
* `systemd` or write your own init scripts
* An LDAP server?

Better Instructions on setting up nginx
---------------------------------------

Building the moxxi server yourself
----------------------------------

First, [install go](https://golang.org/doc/install) somewhere (not on the target system).

Download the source:

```bash
go get github.com/JackKnifed/moxxi
```

Run the tests:

```bash
go test github.com/JackKnifed/moxxiconf
```

Build the binary (instructions are for a target Linux 64bit system):

```bash
GOARCH=amd64 GOOS=linux go install github.com/JackKnifed/moxxi
```

Finally, copy that binary to `/usr/local/bin` on the target system. It should currently be at `$GOPATH/bin/moxxi`.

Nginx Configs
-------------

You will need to create a few folders to configure Nginx:

```bash
mkdir -p /etc/nginx/proxy.d
```

Copy all configs from the nginx folder of the repository to `/etc/nginx/conf.d`. Adjust the domain name in there as needed.

Cron task Configuration
-----------------------

