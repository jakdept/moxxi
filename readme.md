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

Install nginx
-------------

Grab your release name with:

```bash
hostnamectl
```

Add the following lines to `/etc/apt/sources.list` (replace codename in these lines)

```text
deb http://nginx.org/packages/ubuntu/ codename nginx
deb-src http://nginx.org/packages/ubuntu/ codename nginx
```

Install:

```bash
apt-get update
apt-get install -y dpkg-dev
mkdir -p /opt/rebuildnginx
apt-get source -y nginx
apt-get build-dep -y nginx
```

Edit `/opt/rebuildnginx/version/debian/rules` adding the configure flags you need.

* `--with-http_sub_module`
* `--with-http_auth_request_module`

```bash
cd /opt/rebuildnginx/$version
```

```bash
dpkg-buildpackage -b
```

```bash
dpkg --install packagename
```

Server Setup
------------


This section predicates that you have already installed `nginx`, `iptables`, `cron`, `systemd`, `moxxi`, and `syncthing`.

To install these things on Debian 8:

```bash
apt-get install nginx iptables
```

Build/download the syncthing and moxxi then:

```bash
useradd -m moxxi
usermod -aG www-data moxxi
mkdir -p /home/moxxi/bin /home/moxxi/vhosts.d
chgrp www-data /home/moxxi/vhosts.d
```

```
scp moxxi moxxi1:/home/moxxi/bin/moxxi
```

Remove some boilerplate nginx stuff:

```
unlink /etc/nginx/sites-enabled/*
rm -rf /etc/nginx/sites-enabled/ /etc/nginx/sites-available/
mkdir -p /etc/nginx/sites.d
```

Replace `/etc/nginx/nginx.conf` with the version in the repository.

Copy, rename, and edit the following files into /etc/nginx/conf.d/

* `moxxi.parentdomain.com.conf`
* `parentdomain.com.conf`

To configure the scheduled tasks for this, run:

```bash
cat <<EOM >/etc/cron.d/moxxi.cron
*/15 * * * * root /bin/systemctl reload nginx
0 1 * * * moxxi /usr/bin/find /etc/nginx/proxy.d -type f -mtime +31 -delete
EOM
```

Copy the unit files into place for the services, and then start/load them.

```bash
systemctl enable /path/to/config/file
```
