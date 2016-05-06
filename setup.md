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
su -c "mkdir -p /home/moxxi/bin /home/moxxi/vhosts.d /home/moxxi" moxxi
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

Copy, rename, and edit the following files into `/etc/nginx/sites-available`, then symlink then into `/etc/nginx/sites-enabled`:

* `moxxi.parentdomain.com.conf`
* `parentdomain.com.conf`

Drop the following file into `/etc/nginx`:

* `ssl.conf`

You should also probably consider adding access control to the moxxi control vhost - otherwise someone could spam it and create domains. The following cover some different examples:

* `ip_restriction.moxxi.parentdomain.com.conf`

### Firewall setup ###

```bash
cat <<EOM >/etc/network/if-pre-up.d/iptables
#!/bin/sh
/sbin/iptables-restore < /etc/iptables
EOM
chmod +x /etc/network/if-pre-up.d/iptables
```

Copy `iptables` to `/etc/iptables`.

### Cron setup ###

To configure the scheduled tasks for this, add the following to cron:

```bash
*/5 * * * * root /bin/systemctl reload nginx
0 1 * * * moxxi /usr/bin/find /etc/nginx/proxy.d -type f -mtime +31 -delete
```


### moxxi setup ###

Install the `moxxi` binary to `/usr/bin/moxxi`.

Copy the following files to `/home/moxxi`

* `proxy.template`
* `response.template`

Copy the unit file to `/etc/systemd/system/moxxi.service`.

```bash
systemctl enable moxxi.service
systemctl start moxxi.service
```

### syncthing setup ###

Copy the `syncthing` binary to `/usr/bin/syncthing`.

Copy the unit file to `/etc/systemd/system/syncthing@.service`.

```bash
systemctl enable syncthing@moxxi.service
systemctl start syncthing@moxxi.service
```

Use `netstat -tpln` to find the port `syncthing` is running on - likely `8384`.

From your wks, run the following to connect then visit [localhost:8081](localhost:8081):

```bash
ssh -L 8081:127.0.0.1:8384 moxxi1
```