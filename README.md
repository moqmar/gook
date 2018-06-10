# gook - simple, file-based webhook server

A very simple webhook service for linux servers, written in Go.  
If you create an executable script at `/var/www/.webhook`, you can run it by requesting http://localhost:8080/var/www/[key].  
The key must be specified in the second line of the file with `#@GOOK:[key]` - you can generate a secure one using `echo $(tr -dc A-Za-z0-9 < /dev/urandom | head -c 64)`.

Example `.webhook` file:
```
#!/bin/sh
#@GOOK:dontusethiskey

git pull
docker-compose up
```

**WARNING:** It is highly recommended to use a reverse proxy like [Caddy](https://caddyserver.com/) to ensure the connection to Gook is working securely via HTTPS only!

## Features

- POST requests - the body is piped to the STDIN of the script
- Query parameters - appending `?hello=world` results in `$gook_hello` being set to `world`
- Working directory of a script is the folder containing the .webhook file
- Port and host can be set using the `PORT` and `HOST` environment variables

## Setup
```
wget https://static.mo-mar.de/bin/gook -O /usr/local/bin/gook && chmod +x /usr/local/bin/gook
wget https://raw.githubusercontent.com/moqmar/gook/master/gook.service -O /etc/systemd/system/gook.service
systemctl enable gook
systemctl start gook
systemctl status gook
```

## Configuration

The configuration files parsed are `/etc/gook.yaml`, `~/.config/gook.yaml` and `./gook.yaml`.

```yaml
# Only files under the prefix can be requested.
# http://localhost:8080/something/[key] would look for /var/www/something/.webhook in this case.
prefix: /var/www
# .gitignore-like definition of folders to ignore. The default contains the following folders:
# /proc, /sys, /dev, .git/, node_modules/
ignore: |
  .git/
  node_modules/
```

# HTTP Status Codes
 Code  |  Meaning
====== | =========
 `200` | Script exited with exit code `0`.
 `418` | Script exited with a non-zero exit code. More information is available in the `Gook-Error` header.
 `403` | The given key is invalid.
 `404` | In the specified directory is no `.webhook` file, or the file is not available.
 `500` | The `.webhook` file is invalid or couldn't be read.
