Moxxi
=====

This is a HTTP proxy to allow you to access a site at a specific IP address via an alternate domain name.

The intention is to enable easier migrations.

Please see [design and layout](/design.md) to get an idea of how this should work.

Please see [the setup instructions](/setup.md) for setup information.

JSON Format
-----------

JSON requests should be laid out as follows - just included here:

```json
{
  "host": "hostname",
  "ip": "serverIP",
  "port": 443,
  "tls": true,
  "blockedHeaders": [
    "X-Frame-Options",
    "Accept-Encoding"
  ]
},
{
  "host": "hostbaitor.com",
  "ip": "72.52.161.205",
  "port": 80,
  "tls": false,
  "blockedHeaders": [
    "X-Frame-Options",
    "Accept-Encoding"
  ]
},
{
  "host": "deleteos.com",
  "ip": "deleteos.com",
  "port": 443,
  "tls": true,
  "blockedHeaders": [
    "X-Frame-Options",
    "Accept-Encoding"
  ]
}
```

You would then be expected to run them from CLI by putting that into a file, then running:

```bash
curl -d @inputFile -o outputFile moxxi.domain.com/appropiate/JSON/url
```
