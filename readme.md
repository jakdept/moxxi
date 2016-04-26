Moxxi
=====

This is a HTTP proxy to allow you to access a site at a specific IP address via an alternate domain name.

The intention is to enable easier migrations.

Please see [the setup instructions](/setup.md) for setup information.

Design
------

This system consists of a few things:

* `nginx` runs on 443 and 80, and has a lot of server blocks to listen as
* `moxxi` - a custom written binary - listens on 8080 and generates configs with given input from each request - putting all those configs in one directory.
* `cron` deletes all configs older than a threshold in that directory.
* `syncthing` keeps that one directory in sync across the different servers in the cluster.
* A loadbalancer sits in front of it all, splitting traffic among the servers and providing redudancy.

`nginx` has two static server blocks:

* A default server block that simply serves one static file, a notification that the domain requested is expired, not configured, or broken.
* A proxy for contorl - requests htiting this server block are forwarded to `moxxi`.

`nginx` then also includes all the confs in a given folder, and is set to reload (reparse all configs) every 5 minutes.

`iptables` with incoming filtering is used to filter incoming traffic - only the SSH port, 80, and 443 are open. No filtering is done outgoing as this can hit any remote port on a remote server.