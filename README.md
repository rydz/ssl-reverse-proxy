# reverseproxy
Simple HTTPs reverse proxy for an http backend server

## Flags

| Flag                | Default  | Description                                              |
|---------------------|----------|----------------------------------------------------------|
| u                   | :80      | address to bind to on upgrade server                     |
| a                   | :443     | address to bind to on HTTPs server                       |
| cert                | cert.crt | location of certificate file                             |
| key                 | key.key  | location of key file                                     |
| strip-forwarded-for | false    | strip existing forwarded for headers to prevent spoofing |

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

specify server location to proxy to
### TARGET http://webapp:8080

Location of certificate file
### CERT /var/certs/cert.crt

Location of key file
### KEY /var/certs/key.key

Port to host upgrade server on.
### UPGRADE_ADDR :80

Port to host HTTPS server on
### SERVER_ADDR :443

Port to host SSL server on
### STRIP_FORWARDED_FOR "false"
Strip the list of IPs from incoming X-Forwarded-For headers to prevent spoofing