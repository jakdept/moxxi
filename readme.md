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

Server Setup
------------

This section predicates that you have already installed `nginx`, `iptables`, `cron`, `systemd`, `moxxi`, and `syncthing`.

```bash
useradd -m moxxi
usermod -aG www-data
```

You will need to create a few folders to configure `nginx`:

```bash
mkdir -p /etc/nginx/proxy.d
chown moxxi:www-data /etc/nginx/proxy.d
```

`mv` `moxxi.parentdomain.com.conf` into `/etc/nginx/conf.d`.

Replace `/etc/nginx/nginx.conf` with the version in the repository.

Copy `moxxi.parentdomain.com.conf` into `/etc/nginx/sites.d` - modify the file to match your domain name and rename.

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
