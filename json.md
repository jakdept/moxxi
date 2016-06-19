JSON requests should be dumped into the body of a request, and have the (go) format:

```go
struct {
  IntHost      string
  IntIP        string
  IntPort      int
  Encrypted    bool
  StripHeaders []string
}
```

`json` format should look like:

```json
{
  "IntHost": string,
  "IntIP": string,
  "IntPort": string,
  "Encrypted": bool.
  "StripHeaders": []string
}
```

Out of these items, only `host` and `ip` are actually required.

The body of an example request is provided below:

```json
{
  "IntHost": "hostname",
  "IntIP": "serverIP",
  "IntPort": 443,
  "Encrypted": true,
  "StripHeaders": [
    "X-Frame-Options",
    "Accept-Encoding"
  ]
}
{
  "IntHost": "hostbaitor.com",
  "IntIP": "72.52.161.205",
  "IntPort": 80,
  "Encrypted": false,
  "StripHeaders": [
    "X-Frame-Options",
    "Accept-Encoding"
  ]
}
{
  "IntHost": "deleteos.com",
  "IntIP": "deleteos.com",
  "IntPort": 443,
  "Encrypted": true,
  "StripHeaders": [
    "X-Frame-Options",
    "Accept-Encoding"
  ]
}
```

It is then expected to run this with something like:

```bash
curl -d @inputFile -o outputFile moxxi.domain.com/appropiate/JSON/url
```

The expected response depends on your `responseTempl`.

It is recommended that you consider using [response.flat.template](/response.flat.template) with JSON handlers.