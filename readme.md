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

Server Setup
------------


This section predicates that you have already installed `nginx`, `iptables`, `cron`, `systemd`, `moxxi`, and `syncthing`.

To install these things on Debian 8:

```bash
apt-get install nginx iptables
```

Build/download the syncthing and moxxi then:

### User setup ###

```bash
useradd -m moxxi
usermod -aG www-data moxxi
su -c "mkdir -p /home/moxxi/bin /home/moxxi/vhosts.d /home/moxxi/files" moxxi
chgrp www-data /home/moxxi/vhosts.d
```

```
scp moxxi moxxi1:/home/moxxi/bin/moxxi
```

### Nginx setup ###

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
* `ssl.conf`

### Firewall setup ###

```bash
cat <<EOM >/etc/network/if-pre-up.d/iptables
#!/bin/sh
/sbin/iptables-restore < /etc/iptables
EOM
#chmod +x /etc/network/if-pre-up.d/iptables
```

Copy `iptables` to `/etc/iptables`.

### Cron setup ###

To configure the scheduled tasks for this, run:

```bash
cat <<EOM >/etc/cron.d/moxxi.cron
*/15 * * * * root /bin/systemctl reload nginx
0 1 * * * moxxi /usr/bin/find /etc/nginx/proxy.d -type f -mtime +31 -delete
EOM
```


### moxxi setup ###

Copy the following files to `/home/moxxi`

* `proxy.template`
* `response.template`
* `moxxi.service`

Copy the unit files into place for the services, and then start/load them.

```bash
systemctl enable moxxi.service syncthing@moxxi.service
systemctl start moxxi.service syncthing@moxxi.service
```

### syncthing setup ###

The `syncthing` binary should already be running, if it's not, copy it onto the server and get it running.

Use `netstat -tpln` to find the port `syncthing` is running on - likely `8384`.

Use a line like the following to administer it:

```bash
ssh -L 