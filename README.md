<div align="center">

# Velkor

[![License](https://img.shields.io/badge/license-BSD%202--Clause-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/michalskocz/velkor)](https://github.com/michalskocz/velkor/releases)


---

</div>

<div align="center">

Velkor is a lightweight **local CI/CD runner inspired by GitHub Actions**, written in Go.  
It allows you to execute CI/CD pipelines locally inside containers (Docker or Podman), making it easy to test workflows without pushing changes to remote CI systems.

</div>

---

## 🚀 Features

- GitHub Actions–like YAML pipeline syntax
- Local execution of CI/CD pipelines
- Container-based isolation (Docker / Podman)
- Multi-stage pipelines
- Parallel job execution within stages
- Global and job-level environment variables
- Configurable worker threads
- Debug mode for configuration inspection

---

## 📦 Install

Velkor can be installed by rpm, deb and apk. Required files are available in the release [release](https://github.com/michalskocz/velkor/releases). You can also install it by using go:
```bash
go install github.com/michalskocz/velkor/src/velkor@latest
```

### Fedora, RHEL, CentOS, Rocky Linux ...
```bash
curl -L -o velkor.rpm  https://github.com/michalskocz/velkor/releases/download/v1.4.1/velkor.rpm
sudo dnf install ./velkor.rpm
```
### SUSE
```bash
curl -L -o velkor.rpm  https://github.com/michalskocz/velkor/releases/download/v1.4.1/velkor.rpm
sudo zypper install ./velkor.rpm
```

### Debian, Ubuntu, Zorin OS, ...
```bash
curl -L -o velkor.deb  https://github.com/michalskocz/velkor/releases/download/v1.4.1/velkor.deb
sudo apt install ./velkor.deb
```


---

## 🧾 CLI options

| Flag                | Description                                                              |
| ------------------- | ------------------------------------------------------------------------ |
| `-c  --cpu`         | CPU cores per container. Default: 1                                      |
| `-f, --file`        | input config file. Default: ci.yml                                       |
| `-d, --debug-level` | How much information you get. Default: INFO                              |
| `-h, --help`        | Show help                                                                |
| `-i, --isolation`    | Uses flags that reduce security in exchange for performance. Default: ON|
---

## ⚙️ Pipeline configuration

Velkor uses a YAML file to define CI/CD pipelines. [Example](examples/c/ci.yml)
