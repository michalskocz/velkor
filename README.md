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

<div align="center">

<table>
<tr>
<td width="33%">

### 🚀 Pipelines
- GitHub Actions–like YAML syntax
- Multi-stage execution
- Strict ordering

</td>
<td width="33%">

### ⚡ Execution
- Local CI/CD runs
- Parallel tasks
- Configurable workers

</td>
<td width="33%">

### 🔒 Isolation
- Docker support
- Podman support
- Environment variables

</td>
</tr>
</table>

</div>

---

## 📦 Installation

Velkor is distributed via RPM, DEB, and APK packages. You can find the latest binaries on the [Releases page](https://github.com/michalskocz/velkor/releases). 

Alternatively, install it directly using Go:
```bash
go install github.com/michalskocz/velkor/src/velkor@latest
```

### Linux Packages (v1.4.2)

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


<div align="center">

## 🚀 How Velkor Executes Pipelines

<table>
<tr>
<td width="33%" valign="top">

### 🔍 1. Parse & Validate

Velkor loads `ci.yml` and checks:

- ✅ Valid configuration
- ✅ Thread limits
- ✅ Duplicate stages/tasks
- ✅ Stage assignments

</td>
<td width="33%" valign="top">

### 🔄 2. Sequential Stages

Pipelines follow strict stage order:

```
test → build → package
```

Each stage starts only after the previous one completes.

</td>
<td width="33%" valign="top">

### ⚡ 3. Parallel Tasks

Tasks inside a stage run concurrently:

- Configurable worker threads
- Faster execution
- Controlled resource usage

</td>
</tr>

<tr>
<td width="33%" valign="top">

### 📦 4. Container Isolation

Every task runs inside its own container:

- 🐳 Docker support
- 📦 Podman support
- 📂 Workspace mounting
- 🖥️ `sh -c "<script>"` execution

</td>
<td width="33%" valign="top">

### 📤 5. Artifact Extraction

After successful execution:

- Artifacts are collected
- Files are copied from containers
- Results are stored locally

</td>
<td width="33%" valign="top">

### 🛑 6. Error Handling

Pipeline stops immediately when:

- ❌ Task exits with non-zero code
- ❌ Configuration validation fails
- ❌ Execution cannot continue

</td>
</tr>
</table>

</div>
