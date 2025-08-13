# config-diff

A CLI tool to interact with NOS (Network Operating System) configurations based on YANG schemas.  
It supports loading schemas, validating configurations, computing diffs, and performing various config management operations for targets.

---

## Installation

You can install `config-diff` in several ways:

### 1. One-Liner install 
```bash 
curl -fsSL https://raw.githubusercontent.com/sdcio/config-diff/main/install.sh | bash
```

### 2. Download from GitHub Releases (recommended)
Prebuilt binaries for Linux, macOS, and Windows are available.

1. One-Liner install
2. Visit the [Releases page](https://github.com/sdcio/config-diff/releases).
3. Download the archive for your platform.
4. Extract the binary and place it somewhere in your `PATH`:

```bash
tar -xvf config-diff_<version>_<os>_<arch>.tar.gz
sudo mv config-diff /usr/local/bin/
```

### 3. Build from source
If you have Go installed:

```bash
git clone https://github.com/sdcio/config-diff.git
cd config-diff
go build -o config-diff main.go
```

### 4. Install with `go install`
If you just want the latest main branch build:

```bash
go install github.com/sdcio/config-diff@latest
```

## Usage

The general syntax is:

```bash
config-diff [command] [flags]
```

Use `--help` with any command to see its options:

```bash
config-diff schema load --help
```

---

## Command Reference

### **`config` — Config-based actions**
Manage and inspect device configurations.

#### Load a single config file
```bash
config-diff config load -t <target> --file <path|-> --file-format <format> [--priority 500] [--intent-name <name>]
```
Flags:
- `--file string` – Config file path or `-` for stdin
- `--file-format string` – One of `json`, `json-ietf`, `xml`, `sdc`, etc.
- `--priority int` – Config priority (default `500`)
- `--intent-name string` – Name of the configuration intent

#### Bulk load configs
```bash
config-diff config bulk -t <target> --files file1.yaml,file2.yaml
```
- `--files stringSlice` – List of files to load

#### Blame config changes
```bash
config-diff config blame -t <target> [--include-defaults]
```
- `--include-defaults` – Include schema defaults

#### Show configuration
```bash
config-diff config show -t <target> [-o json] [-a]
```
- `-o, --out-format string` – Output format (`json`, `xml`, etc.)
- `-a, --all` – Show entire config, not just updates

#### Diff config with running
```bash
config-diff config diff -t <target> [--type side-by-side-patch] [--context 2] [--no-color] [-o json]
```
- `--type string` – Diff type
- `--context int` – Context lines (default 2)
- `--no-color` – Disable colored output
- `-o, --out-format string` – Output format

#### Validate configuration
```bash
config-diff config validate -t <target>
```

---

### **`schema` — Schema-based actions**
Manage YANG schema versions and definitions.

#### List schemas
```bash
config-diff schema list
```

#### Load schema
```bash
config-diff schema load -t <target> -f schema.yaml [--cleanup]
```
- `-f, --schema-def string` – Schema definition file (**required**)
- `-c, --cleanup` – Cleanup schema directory after load (default `true`)

#### Remove schema
```bash
config-diff schema remove [-f schema.yaml] [--vendor <vendor>] [--version <version>]
```

---

### **`target` — Target-based actions**
Manage configured targets.

#### Show target details
```bash
config-diff target show -t <target>
```

#### Remove target
```bash
config-diff target remove -t <target>
```

---

## Persistent Flags

Some commands share persistent flags:

- `-t, --target string` – The target to use (**required**)
- `-o, --out-format string` – Output format (`json`, `xml`, etc.)

---

## Examples

**Load a schema:**
```bash
config-diff schema load -t router1 -f schema.yaml
```

**Validate a config:**
```bash
config-diff config validate -t router1
```

**Diff config changes:**
```bash
config-diff config diff -t router1 --type side-by-side-patch --context 3
```

**Show current configuration:**
```bash
config-diff config show -t router1 -o json
```
