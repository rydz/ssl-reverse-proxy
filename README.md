# ssl-reverse-proxy
Simple HTTPs reverse proxy for an http backend server

## Installing

`go get -u github.com/rydz/ssl-reverse-proxy`


## Example
If this proxy is the first in your chain, you should enable the -strip-forwarded-for flag which will strip incoming `X-Forwarded-For` headers from the incoming requests.

X-Forwarded-For headers allow servers behind proxies to know the origin IP address of the request.

```sh
ssl-reverse-proxy \
  -target "http://localhost:8080" \
  -strip-forwarded-for \
  -cert "certificate.crt" \
  -key "key.key" \
```


## Flags

| Flag                | Default  | Description                                              |
|---------------------|----------|----------------------------------------------------------|
| u                   | :80      | address to bind to on upgrade server                     |
| a                   | :443     | address to bind to on HTTPs server                       |
| cert                | cert.crt | location of certificate file                             |
| key                 | key.key  | location of key file                                     |
| strip-forwarded-for | false    | strip existing forwarded for headers to prevent spoofing |
| formatter           | text     | logrus formatted for logs, one of 'text' or 'json'       |
| color               | false    | enable color in the text formatter                       |

## Docker

### Example usage
`docker build -t reverse-proxy .`

```sh
docker run \
  --network=webappnetwork \
  -e "TARGET=http://webapp" \
  -p "80:80" \
  -p "443:443" \
  -v "./certs":"/var/certs/" \
  -d \
  --name reverse-proxy \
  reverse-proxy
```

Environment variables

### TARGET http://webapp:8080
Location of target server

### CERT /var/certs/cert.crt
Location of certificate file

### KEY /var/certs/key.key
Location of key file

### UPGRADE_ADDR :80
Port to host upgrade server on.

### SERVER_ADDR :443
Port to host SSL server on

### STRIP_FORWARDED_FOR "false"
Strip the list of IPs from incoming X-Forwarded-For headers to prevent spoofing
