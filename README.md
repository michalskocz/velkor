<div align="center">

# Velkor

[![License](https://img.shields.io/badge/license-BSD%202--Clause-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/michalskocz/velkor)](https://github.com/michalskocz/velkor/releases)

<p>
  <b>A lightweight local CI/CD runner inspired by GitHub Actions, written in Go.</b> <br>
  Execute CI/CD pipelines locally inside Docker or Podman containers to test workflows without pushing changes to remote repositories.
</p>

</div>

---

## 🚀 Features

- GitHub Actions–like YAML pipeline syntax
- Local execution of CI/CD pipelines
- Container-based isolation (Auto-detects Docker or Podman)
- Multi-stage pipelines with strict ordering
- Parallel task execution within stages
- Global and task-level environment variables
- Configurable worker threads
- Detailed debug mode for configuration inspection

---

## 📦 Installation

Velkor is distributed via RPM, DEB, and APK packages. You can find the latest binaries on the [Releases page](https://github.com/michalskocz/velkor/releases). 

Alternatively, install it directly using Go:
```bash
go install github.com/michalskocz/velkor/src/velkor@latest
```

### Linux Packages (v1.4.1)

**Fedora / RHEL / CentOS / Rocky Linux:**

```bash
curl -L -o velkor.rpm https://github.com/michalskocz/velkor/releases/download/v1.4.2/velkor.rpm
sudo dnf install ./velkor.rpm
```

**openSUSE:**

```bash
curl -L -o velkor.rpm https://github.com/michalskocz/velkor/releases/download/v1.4.2/velkor.rpm
sudo zypper install ./velkor.rpm
```

**Debian / Ubuntu / Linux Mint:**

```bash
curl -L -o velkor.deb https://github.com/michalskocz/velkor/releases/download/v1.4.2/velkor.deb
sudo apt install ./velkor.deb
```

---

## 🧾 CLI Options

| Flag | Description |
| --- | --- |
| `-c,  --cpu` | CPU cores per container. Default: `1` |
| `-f,  --file` | Input config file. Default: `ci.yml` |
| `-d,  --debug-level` | Logging verbosity. Default: `INFO` |
| `-i,  --isolation` | Uses flags that reduce security in exchange for performance. Default: `ON` |
| `-h,  --help` | Show help |

---

## ⚙️ Configuration Guide (`ci.yml`)

Velkor uses a simple, declarative YAML format to define pipelines. A pipeline consists of **global settings**, ordered **stages**, and **tasks** assigned to those stages.

Check out the [Velkor CI config](src/ci.yml) or the [C project example](examples/c/ci.yml) for reference.

### Configuration Blocks

* **`threads`**: Controls max concurrent tasks. Must be `> 0` (Default: `1`).
* **`variables`**: Global key-value pairs injected into all tasks.
* **`stages`**: Ordered execution list. Velkor executes one stage at a time. Tasks must reference a stage listed here.
* **`[task_name]`**: Any other top-level key is treated as a task.

### Task Properties

| Field | Description |
| --- | --- |
| `stage` | Matches a stage from the global `stages` list. |
| `image` | Container image to run (e.g., `golang:1.26`, `fedora:44`). |
| `type` | (Optional) Force `docker` or `podman`. Auto-detected if omitted. |
| `script` | Array of shell commands executed sequentially in the container. |
| `artifacts` | Files copied from `/workspace` to `artifacts/<task_name>/` upon success. |
| `variables` | Task-specific variables that merge with global variables. |

---

## 🚀 How Velkor Executes Pipelines

1. **Parse & Validate:** Velkor reads `ci.yml`, validates thread counts, checks for duplicate stages/tasks, and verifies stage assignments.
2. **Run Stages Sequentially:** It moves through the `stages` list in strict order (e.g., `test` → `build` → `package`).
3. **Execute Tasks Concurrently:** For each stage, tasks run in parallel up to the `threads` limit.
4. **Container Isolation:** Each task spins up a container, mounts the local workspace, and runs `sh -c "<script>"`.
5. **Extract Artifacts:** If a task succeeds, defined artifacts are copied from the container to the host machine.
6. **Error Handling:** If any task fails (non-zero exit code), or if configuration is invalid, the pipeline halts immediately.

---

## 🧪 Smoke Testing Example

Smoke tests are a great way to verify packages inside fresh distro containers. Velkor handles this elegantly:

```yaml
install_fedora:
  stage: smoke
  image: fedora:44
  script:
    - cp artifacts/build_packages/velkor.rpm velkor.rpm
    - dnf install -y ./velkor.rpm
    - velkor --help

```
