# Gook - simple, file-based webhook server

[![](https://ci.mo-mar.de/api/badges/moqmar/gook/status.svg)](https://ci.mo-mar.de/moqmar/gook)

A very simple webhook service for linux servers, written in Go.  
If you create an executable script at `/var/www/.webhook` (we'll call this a "gookfile" or a "webhook script"), you can run it by requesting http://localhost:8080/var/www/[key].  
The key must be specified in the second line of the file with `#@gook:[key]`, and can contain any number of flags, separated by a `+` - e.g. `#@gook+flag1+flag2:[key]`.

You can **generate a secure key** using `echo $(tr -dc A-Za-z0-9 < /dev/urandom | head -c 64)`.

Example `.webhook` file:
```
#!/bin/sh
#@gook+stdin:dontusethiskey

git pull
docker-compose up
```

**WARNING:** It is highly recommended to use a reverse proxy like [Caddy](https://caddyserver.com/) to ensure the connection to Gook is working securely via HTTPS only!

## Features

- POST requests - the body is piped to the STDIN of the script
- Query parameters - appending `?hello=world` results in `$gook_hello` being set to `world`
- Working directory of a script is the folder containing the .webhook file
- Port and host can be set using the `PORT` and `HOST` environment variables

---

# Documentation

## How the gookfile is selected
If `/<path>/<key>` is requested (key can be empty, although not recommended, but the slash is required in that case), the gookfile is located at `<prefix>/<path>/.webhook`.

## Preconditions for a gookfile
- Must be named `.webhook`
- Must be executable
- Must start with a shebang (e.g. `#!/bin/sh`)
- The second line ("gookline") must have the format `#@gook[+<flags>]:[key]`
- The capitalization of "gook" and the flags doesn't matter, but the key is case-sensitive
- Every flag must be valid

## Flags
 Flag  |  Meaning
------ | ---------
`stdin`| Pipe HTTP POST body to the script's STDIN.

## HTTP Status Codes
 Code  |  Meaning
------ | ---------
 `200` | Script exited with exit code `0`.
 `424` | Script exited with a non-zero exit code. The exit code is available in the `Gook-Status` header.
 `418` | Script exited with an error (e.g. if the process is killed). More information is available in the `Gook-Error` header.
 `403` | The given key is invalid.
 `404` | In the specified directory is no `.webhook` file, or the file is not available. More information is available in the Gook server logs.
 `500` | The `.webhook` file is invalid or couldn't be read.

---

# Installation and Configuration

## Installation with systemd
```
wget https://get.mo-mar.de/gook -O /usr/local/bin/gook && chmod +x /usr/local/bin/gook
wget https://raw.githubusercontent.com/moqmar/gook/master/gook.service -O /etc/systemd/system/gook.service
systemctl enable gook
systemctl start gook
systemctl status gook
```

## Configuration file

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
