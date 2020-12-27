# Gook - simple, file-based webhook server

[![](https://ci.mo-mar.de/api/badges/moqmar/gook/status.svg)](https://ci.mo-mar.de/moqmar/gook)

**A simple but configurable and secure webhook service for linux servers, written in Go.**

The short explanation: If you create an executable script at `/var/www/example/.webhook` (we'll call this a "gookfile" or a "webhook script") with the prefix set to `/var/www` in `/etc/gook.yaml`, you can run it by requesting http://localhost:8080/example/[key].

The key must be specified in the second line of the script file with `#@gook:[key]`, and that line can also contain multiple flags that change the webhook behaviour, separated by a `+` - e.g. `#@gook+flag1+flag2:[key]`.

Example `.webhook` file:
```
#!/bin/sh
#@gook+stdin:dontusethiskey

git pull

# Using STDIN to check if the POST body contains the word "recreate":
if grep 'recreate'; then
  docker-compose up --force-recreate
else
  docker-compose up
fi
```

## Features

- Query parameters - appending `?hello=world` results in `$gook_hello` being set to `world`
- POST requests - the body is piped to the STDIN of the script
- Working directory of a script is the folder containing the .webhook file
- Port and host can be set using the `PORT` and `HOST` environment variables

## ⚠️ SECURITY CONSIDERATIONS ⚠️
- The software can be considered production-ready, but we don't make any guarantees that it will work or not break anything.
- You can (and should) **generate a secure key** using `echo $(tr -dc A-Za-z0-9 < /dev/urandom | head -c 64)`.
- It is recommended that the webhook script is only readable by the user the Gook server is running under.
- Using a reverse proxy like [Caddy](https://caddyserver.com/) is recommended to ensure the connection to Gook is working securely via HTTPS only.

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

<!-- TODO: ## Environment Variables -->

---

# Installation and Configuration

## Installation with systemd
```
wget https://github.com/moqmar/gook/releases/download/latest/gook -O /usr/local/bin/gook && chmod +x /usr/local/bin/gook
wget https://raw.githubusercontent.com/moqmar/gook/master/gook.service -O /etc/systemd/system/gook.service
systemctl enable gook
systemctl start gook
systemctl status gook
```

The binary file provided on https://get.mo-mar.de/gook is compiled by the CI against the git master, for Linux x86 64-bit, and is statically linked.  
If you want to use Gook on another system, you can build it yourself using Go: `go get github.com/moqmar/gook && cp "$(go env GOPATH)/bin/gook" /usr/local/bin/gook`

## Configuration file

The configuration files parsed are `/etc/gook.yaml`, `~/.config/gook.yaml` and `./gook.yaml`.

```yaml
# Only files under the prefix can be requested.
# http://localhost:8080/something/[key] would look for /var/www/something/.webhook in this case.
prefix: /var/www
# Only those exact directories are allowed - the recommended alternative to `ignore`.
# WARNING: if the whitelist is empty, every .webhook file is allowed which can cause security issues!
whitelist:
- something # /var/www/something/.webhook
# .gitignore-like definition of folders to ignore. The default contains the following folders:
# /proc, /sys, /dev, .git/, node_modules/
# DEPRECATED for security issues!
ignore: |
  .git/
  node_modules/
```
