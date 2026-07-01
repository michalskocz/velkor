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

## 📦 Build

Build the project using the project. The result will be in artifacts/dist/velkor or download [relese](https://github.com/michalskocz/velkor/releases):

```bash
cd src && chmod +x quick.sh
./quick.sh
````

---

## 🧾 CLI options

| Flag                | Description                                                              |
| ------------------- | ------------------------------------------------------------------------ |
| `-c  --cpu`         | CPU cores per container. Default: 1                                      |
| `-f, --file`        | nput config file. Default: ci.yml                                        |
| `-d, --debug-level` | How meny info you got. Default: INFO                                     |
| `-h, --help`        | Show help                                                                |
| `i, --isolation`    | Uses flags that reduce security in exchange for performance. Default: ON |
---

## 🧱 Performence
CI/CD performance when building the project on my laptop [6c/12t]. 
For testing, I'm running part of the CI/CD pipeline (Linux build and test) with an action[Github](.github/workflows/build.yml) [local](src/ci.yml). 

| CLI                   | Time      |
| --------------------- | --------- |
| ./velkor              | 86 [s]    |
| ./velkor -c 6         | 23 [s]    |         
| ./velkor -c 12        | 15 [s]    |
| ./velkor -c 12 -i OFF | 15 [s]    |
---

## ⚙️ Pipeline configuration

Velkor uses a YAML file to define CI/CD pipelines. [Example](examples/c/ci.yml)


---

## 🐳 Containers

Each task runs inside a container:

* working directory mounted into `/workspace`
* executed via shell (`sh -c`)
* supports Docker and Podman

---

## 🔁 Execution flow

1. Load YAML configuration
2. Parse stages and tasks
3. Validate pipeline structure
4. Execute stages sequentially
5. Run tasks concurrently inside each stage
6. Stop pipeline on first failure
