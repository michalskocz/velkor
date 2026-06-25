
# Velkor

Velkor is a lightweight **local CI/CD runner inspired by GitHub Actions**, written in Go.  
It allows you to execute CI/CD pipelines locally inside containers (Docker or Podman), making it easy to test workflows without pushing changes to remote CI systems.

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

Build the binary using Go:

```bash
cd src
go build -o velkor ./exe/*
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
| `-f, --file`        | Path to pipeline file (default: `ci.yml`) |
| `-d, --debug-level` | `INFO` or `DEBUG`                         |
| `-h, --help`        | Show help                                 |

---

## ⚙️ Pipeline configuration

Velkor uses a YAML file to define CI/CD pipelines.

### Example pipeline

```yaml
stages:
  - build
  - test
  - run

threads: 4

variables:
  - APP_NAME: "velkor"
  - GREETING: "Hello from Velkor"

build_app:
  stage: build
  image: golang:1.26
  type: docker # or podman
  script:
    - go version
    - echo "Building $APP_NAME"

unit_tests:
  stage: test
  image: golang:1.26
  type: docker # or podman
  script:
    - go test ./...

integration_echo:
  stage: run
  image: alpine:latest
  type: docker # or podman
  script:
    - echo "$GREETING"
    - echo "Pipeline finished successfully"
```

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

Example behavior:

```bash
docker run --rm -v "$PWD:/workspace" -w /workspace golang:1.26 sh -c "<script>"
```

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

### Usage

```bash
echo $ENV
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

---

## 📁 Project structure

```
src/
├── configuration   # YAML parsing + pipeline model
├── exe             # CLI + pipeline runner
├── internal        # queue + debug + constants
└── example.yml     # sample pipeline
```

---

## ❗ Error handling

* invalid YAML → fatal error
* unknown stage → pipeline abort
* task failure → stops pipeline immediately
* duplicate stages/tasks → rejected during parsing

---

## 🧠 Design goals

* simplicity over complexity
* fast local CI feedback loop
* reproducible container-based execution
* minimal configuration overhead

---

## 📌 Example output

```
mchal@localhost:~/Dokumenty/forgejo/LocalAction/src$ ./velkor -d DEBUG

=== CONFIGURATION ===
Threads: 4

[Global Variables]
  APP_NAME = velkor
  GREETING = Hello from Velkor

[Stages and Tasks]
► Stage 1: build
    ◇ Task: build_app
      Image:  golang:1.22 (Podman)
      Script:
        > go version
        > echo "Building $APP_NAME"
► Stage 2: test
    ◇ Task: unit_tests
      Image:  golang:1.26 (Podman)
      Script:
        > go test ./...
► Stage 3: run
    ◇ Task: integration_echo
      Image:  alpine:latest (Podman)
      Script:
        > echo "$GREETING"
        > echo "Pipeline finished successfully"
====================

Starting CI/CD pipeline...

--- Running Stage: build ---
[Task: build_app] Running in container: golang:1.22
go version go1.22.12 linux/amd64
Building velkor
[Task: build_app] Completed successfully.

--- Running Stage: test ---
[Task: unit_tests] Running in container: golang:1.26
go: downloading gopkg.in/yaml.v3 v3.0.1
go: downloading github.com/akamensky/argparse v1.4.0
?   	velkor/configuration	[no test files]
?   	velkor/exe	[no test files]
?   	velkor/internal	[no test files]
[Task: unit_tests] Completed successfully.

--- Running Stage: run ---
[Task: integration_echo] Running in container: alpine:latest
Hello from Velkor
Pipeline finished successfully
[Task: integration_echo] Completed successfully.

Success! All stages completed successfully.
mchal@localhost:~/Dokumenty/forgejo/LocalAction/src$
```

---

## 📈 Future improvements

* artifact passing between stages
* caching support
* conditional jobs (if/when rules)
* matrix builds
* remote execution agents
* secrets management
