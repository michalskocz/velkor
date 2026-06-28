<div align="center">

# Velkor

[![License](https://img.shields.io/badge/license-BSD%202--Clause-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/michalskocz/velkor)](https://github.com/twoj-login/twoje-repo/releases)


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

## ▶️ Usage

### Run pipeline

```bash
./velkor -f ci.yml
```

### Run with debug output

```bash
./velkor -f ci.yml -d DEBUG
```

### Help

```bash
./velkor --help
```

---

## 🧾 CLI options

| Flag                | Description                               |
| ------------------- | ----------------------------------------- |
| `-c  --cpu`         | Cores per container                       |
| `-f, --file`        | Path to pipeline file (default: `ci.yml`) |
| `-d, --debug-level` | `INFO` or `DEBUG`                         |
| `-h, --help`        | Show help                                 |

---

## ⚙️ Pipeline configuration

Velkor uses a YAML file to define CI/CD pipelines. [Example](examples/c/ci.yml)

---

## 🧱 Pipeline model

Velkor executes pipelines in two layers:

* **Stages** → executed sequentially
* **Tasks (jobs)** → executed in parallel within a stage (limited by `threads`)

```
Pipeline
 ├── Stage: build   → jobs run in parallel
 ├── Stage: test    → jobs run in parallel
 └── Stage: run     → jobs run in parallel
```

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

---

## 🌍 Variables

### Global variables

```yaml
variables:
  - ENV: "production"
```
### Local variables:
```yaml
build_client:
  stage: build
  image: ubuntu:latest
  variables:
    - CFLAGS: "-O3"
  artifacts: ["out/client"]
  script:
    - mkdir out
    - gcc $CFLAGS -o out/client  
```


Variables are injected into container environment.

---

## 🧵 Concurrency

Controlled via:

```yaml
threads: 4
```

* limits number of parallel jobs per stage
* default: `1`

---

## 🐞 Debug mode

Enable configuration preview:

```bash
./velkor -f ci.yml -d DEBUG
```

Shows:

* parsed stages
* tasks
* images
* scripts
* variables
