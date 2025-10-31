# interchange

A lightweight, customizable reverse proxy.

Currently in active development

## Features

- [x] Reverse Proxying
- [x] Configuration from a single file
- [x] Static File Hosting
- [ ] Rate Limiting
- [ ] IP Blacklist/Whitelist
- [ ] Load Balancing

## Usage

Create a `interchange.toml` in the same directory as `interchange`

Here is an example configuration:
```toml
hostAddress = "127.0.0.1"
port = 8000

[services.static]
mode = "staticFS"
route = "/static"
directory = "./"
showDirectoryBrowser = false

[services.app]
mode = "reverseProxy"
route = "/"
target = "https://127.0.0.1:5000"
```

## License

interchange is licensed under the MIT license