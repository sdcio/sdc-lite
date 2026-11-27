# sdc-lite

![sdc logo](https://docs.sdcio.dev/assets/logos/SDC-transparent-withname-100x133.png)


`sdc-lite` is a CLI tool to interact with network operating system (NOS) configurations based on YANG schemas.  
It provides powerful capabilities for working with configurations — from schema management to validation — and also allows configuration format conversion.

With `sdc-lite`, you can:
- Load YANG schemas
- Validate configurations against loaded schemas
- Compare and inspect configuration differences
- Blame resulting config to see the contributing intents
- Convert configurations between formats (e.g., load a config in `json_ietf` format, then output it as `xml`)

---

## Installation

You can install `sdc-lite` in several ways:

### 1. One-Liner install 
```bash 
curl -fsSL https://raw.githubusercontent.com/sdcio/sdc-lite/main/install.sh | bash
```

### 2. Download from GitHub Releases (recommended)
Prebuilt binaries for Linux, macOS, and Windows are available.

1. One-Liner install
2. Visit the [Releases page](https://github.com/sdcio/sdc-lite/releases).
3. Download the archive for your platform.
4. Extract the binary and place it somewhere in your `PATH`:

```bash
tar -xvf sdc-lite_<version>_<os>_<arch>.tar.gz
sudo mv sdc-lite /usr/local/bin/
```

### 3. Build from source
If you have Go installed:

```bash
git clone https://github.com/sdcio/sdc-lite.git
cd sdc-lite
go build -o sdc-lite main.go
```

### 4. Install with `go install`
If you just want the latest main branch build:

```bash
go install github.com/sdcio/sdc-lite@latest
```

### Enabling Shell Completions

`sdc-lite` provides tab-completion for commands, flags, and target names.

After installation, you can enable completions for your shell:

**Bash**
```
sdc-lite completion bash > ~/.bash_completion.d/sdc-lite
echo "source ~/.bash_completion.d/sdc-lite" >> ~/.bashrc
source ~/.bashrc
```

**Zsh**
```
mkdir -p ~/.zsh/completions
sdc-lite completion zsh > ~/.zsh/completions/_sdc-lite
echo "fpath=(~/.zsh/completions $fpath)" >> ~/.zshrc
autoload -Uz compinit && compinit
source ~/.zshrc
```

**Fish**
```
mkdir -p ~/.config/fish/completions
sdc-lite completion fish > ~/.config/fish/completions/sdc-lite.fish
```

Tip: If you use the provided `install.sh` script, completions are installed automatically for Bash, Zsh, and Fish.

---

## Examples

**Load a schema:**
```bash
sdc-lite schema load -t router1 -f https://raw.githubusercontent.com/sdcio/config-server/refs/heads/main/example/schemas/schema-nokia-srl-24.10.1.yaml
```
Creates a target by the name of router1, downloads the referenced schema data and assignes them to the target.

> **IMPORTANT:** The schema.yaml is a schema definition file used by sdc. The file format is described here [sdc schema doc](https://docs.sdcio.dev/user-guide/configuration/schemas/). Example schema definitions for different vendors can be found here as well.

**Load a baseline / running config**
```bash
sdc-lite config load -t router1 --file-format json --intent-name running --file https://raw.githubusercontent.com/sdcio/sdc-lite/refs/tags/v0.1.0/data/config/running/running_srl_01.json 
```
Output:
```
Target: router1
File: https://raw.githubusercontent.com/sdcio/sdc-lite/refs/tags/v0.1.0/data/config/running/running_srl_01.json - Name: running, Prio: 2147483547, Flag: update, Format: json - successfully loaded
```

**Load config snippet:**
```bash
sdc-lite config load -t router1 --file https://raw.githubusercontent.com/sdcio/sdc-lite/refs/tags/v0.1.0/data/config/additions/srl_01.json --file-format json --intent-name config1 --priority 50
```
Output:
```
Target: router1
File: data/config/additions/srl_01.json - Name: config1, Prio: 50, Flag: update, Format: json - successfully loaded
```

**Load sdc config intent**
```bash
sdc-lite config load -t router1 --file-format sdc  --file https://raw.githubusercontent.com/sdcio/sdc-lite/refs/tags/v0.1.0/data/config/additions/srl_01_sdc.yaml
```
Output:
```
Target: router1
File: https://raw.githubusercontent.com/sdcio/sdc-lite/refs/tags/v0.1.0/data/config/additions/srl_01_sdc.yaml - Name: test-orphan, Prio: 10, Flag: update, Format: json - successfully loaded
```

**Show Target details:**
```bash
sdc-lite target show -t router1 
```
Output:
```
Target: router1 (/home/mava/.cache/sdc-lite/targets/router1)
    Schema:
      Name: srl.nokia.sdcio.dev
      Version: 24.10.1
    Intent: config1
      Prio: 50
      Flag: update
      Format: json
    Intent: running
      Prio: 2147483547
      Flag: update
      Format: json
    Intent: test-orphan
      Prio: 10
      Flag: update
      Format: json
```

**Show current configuration:**
```bash
sdc-lite config show -t router1 -o json -a --path /interface[name="ethernet-1/1"]
```
Output formats can also be `json_ietf` or `xml`.
If you want to see only addtions on top of running, remove the `-a` option.

Output
```
Target: router1
{
 "admin-state": "enable",
 "name": "ethernet-1/1",
 "subinterface": [
  {
   "index": 2,
   "type": "bridged",
   "vlan": {
    "encap": {
     "single-tagged": {
      "vlan-id": 2
     }
    }
   }
  }
 ],
 "vlan-tagging": true
}
```

**Validate a config:**
```bash
sdc-lite config validate -t router1
```
Output:
```
Target: router1
Validations performed:
  leafref: 25
  length: 113
  mandatory: 5
  min/max: 4
  must-statement: 785
  pattern: 23
  range: 125
Successful Validated!
```

**Diff config changes:**
```bash
sdc-lite config diff -t router1 --type patch 
```
Output:
```
Target: router1
@@ -1720,5 +1720,19 @@
     {
       "admin-state": "enable",
-      "name": "ethernet-1/1"
+      "name": "ethernet-1/1",
+      "subinterface": [
+        {
+          "index": 2,
+          "type": "bridged",
+          "vlan": {
+            "encap": {
+              "single-tagged": {
+                "vlan-id": 2
+              }
+            }
+          }
+        }
+      ],
+      "vlan-tagging": true
     },
     {
---
@@ -1740,4 +1754,9 @@
         }
       ]
+    },
+    {
+      "admin-state": "enable",
+      "description": "k8s-system0-dummy",
+      "name": "system0"
     }
   ],
---
```


**Blame - show intent sources of configuration**

```bash
sdc-lite config blame -t router1 -p /interface
```

Output:
```
Target: router1
      -----    │     🎯 interface
      -----    │     ├── 📦 ethernet-1/1
    config1    │     │   ├── 🍃 admin-state -> enable
    config1    │     │   ├── 🍃 name -> ethernet-1/1
      -----    │     │   ├── 📦 subinterface
      -----    │     │   │   └── 📦 2
    config1    │     │   │       ├── 🍃 index -> 2
    config1    │     │   │       ├── 🍃 type -> bridged
      -----    │     │   │       └── 📦 vlan
      -----    │     │   │           └── 📦 encap
      -----    │     │   │               └── 📦 single-tagged
    config1    │     │   │                   └── 🍃 vlan-id -> 2
    config1    │     │   └── 🍃 vlan-tagging -> true
      -----    │     ├── 📦 mgmt0
    running    │     │   ├── 🍃 admin-state -> enable
    running    │     │   ├── 🍃 name -> mgmt0
      -----    │     │   └── 📦 subinterface
      -----    │     │       └── 📦 0
    running    │     │           ├── 🍃 admin-state -> enable
    running    │     │           ├── 🍃 index -> 0
    running    │     │           ├── 🍃 ip-mtu -> 1500
      -----    │     │           ├── 📦 ipv4
    running    │     │           │   ├── 🍃 admin-state -> enable
    running    │     │           │   └── 🍃 dhcp-client -> {}
      -----    │     │           └── 📦 ipv6
    running    │     │               ├── 🍃 admin-state -> enable
    running    │     │               └── 🍃 dhcp-client -> {}
      -----    │     └── 📦 system0
test-orphan    │         ├── 🍃 admin-state -> enable
test-orphan    │         ├── 🍃 description -> k8s-system0-dummy
test-orphan    │         └── 🍃 name -> system0
```

**Remove the target for cleanup**
```bash
sdc-lite target remove -t router1 
```
Output:
```
Target: router1
INFO[0000] target router1 - successfully removed        
```

## Usage

The general syntax is:

```bash
sdc-lite [command] [flags]
```

Use `--help` with any command to see its options:

```bash
sdc-lite schema load --help
```

---

## Command Reference

### **`config` — Config-based actions**
Manage and inspect device configurations.

#### Load a single config file
```bash
sdc-lite config load -t <target> --file <path|-> --file-format <format> [--priority 500] [--intent-name <name>]
```
Flags:
- `--file string` – Config file path or `-` for stdin
- `--file-format string` – One of `json`, `json-ietf`, `xml`, `sdc`, etc.
- `--priority int` – Config priority (default `500`)
- `--intent-name string` – Name of the configuration intent
- `--rpc` - Print the rpc definition for the actual command

#### Bulk load configs
```bash
sdc-lite config bulk -t <target> --files file1.yaml,file2.yaml
```
- `--files stringSlice` – List of files to load

#### Blame config changes
```bash
sdc-lite config blame -t <target> [--include-defaults]
```
- `--include-defaults` – Include schema defaults
- `--rpc` - Print the rpc definition for the actual command

#### Show configuration
```bash
sdc-lite config show -t <target> [-o json] [-a]
```
- `-o, --out-format string` – Output format (`json`, `xml`,`json_ietf`, etc.)
- `-a, --all` – Show entire config, not just updates
- `--rpc` - Print the rpc definition for the actual command

#### Diff config with running
```bash
sdc-lite config diff -t <target> [--type side-by-side-patch] [--context 2] [--no-color] [-o json]
```
- `--type string` – Diff type
- `--context int` – Context lines (default 2)
- `--no-color` – Disable colored output
- `-o, --out-format string` – Output format
- `--rpc` - Print the rpc definition for the actual command

#### Validate configuration
```bash
sdc-lite config validate -t <target>
```

---

### **`schema` — Schema-based actions**
Manage YANG schema versions and definitions.

#### List schemas
```bash
sdc-lite schema list
```

#### Load schema
```bash
sdc-lite schema load -t <target> -f schema.yaml [--cleanup]
```
- `-f, --schema-def string` – Schema definition file (**required**)
- `-c, --cleanup` – Cleanup schema directory after load (default `true`)
- `--rpc` - Print the rpc definition for the actual command

#### Remove schema
```bash
sdc-lite schema remove [-f schema.yaml] [--vendor <vendor>] [--version <version>]
```

---

### **`target` — Target-based actions**
Manage configured targets.

#### Show target details
```bash
sdc-lite target show -t <target>
```

#### Remove target
```bash
sdc-lite target remove -t <target>
```

### **`pipeline` — Pipeline-based actions**
Automate sequences of configuration operations.

#### Run a pipeline
```bash
sdc-lite pipeline run --pipeline-file <pipeline.json> [-t <target>] [--persist]
```
- `--pipeline-file string` – Path to the pipeline definition (NDJSON) file, `-` for stdin, or a named pipe (FIFO). Each line is a JSON-RPC 2.0 Request (`id` set) or Notification (`id` omitted).
- `-t, --target string` – Start from an existing target's stored schema and intents. The target must already exist with a schema (create it with `schema load -t <target>`).
- `--persist` – Write the resulting state (schema and all intents, including ones loaded by pipeline steps) back to the target. Requires `-t`. Nothing is written if a batch step fails.

By default a pipeline starts **empty and in-memory**: persisted targets are not loaded and nothing is written back. Steps that need data (`config-show`, `config-validate`, `config-diff`, `config-blame`) fail with a "no configuration loaded" error unless the pipeline first runs `schema-load` and `config-load` steps or is started with `-t <target>`.

Use `--rpc` on any `schema load` or `config` subcommand (no `-t` needed unless converting `--file-format sdc`) to print its JSON-RPC 2.0 step instead of executing it:
```bash
sdc-lite schema load -f schema.yaml --rpc > pipe.json
sdc-lite config load --file running.json --file-format json --intent-name running --rpc >> pipe.json
sdc-lite config show -a --path /interface -o yaml --rpc >> pipe.json
sdc-lite config validate --rpc >> pipe.json
```

Example `pipe.json`:
```json
{"method":"schema-load","jsonrpc":"2.0","id":1,"params":{"file":"schema.yaml"}}
{"method":"config-load","jsonrpc":"2.0","id":2,"params":{"name":"running","prio":500,"flags":{"new":false,"delete":false,"only-intended":false},"format":"json","base-path":"","file":"running.json"}}
{"method":"config-show","jsonrpc":"2.0","id":3,"params":{"all":true,"path":"/interface","format":"yaml"}}
{"method":"config-validate","jsonrpc":"2.0","id":4,"params":{}}
```

Every Request step (`id` present) produces a JSON-RPC 2.0 Response with a self-describing `result` (`{"method": "...", "format"?: "...", "data"?: ...}`) or structured `error` (`{"code": ..., "message": "...", "data"?: {"method": "..."}}`). Steps without `id` act as JSON-RPC 2.0 Notifications and emit no stdout response. See [`openrpc.json`](openrpc.json) for the full schema.

---

## Persistent Flags

Some commands share persistent flags:

- `-t, --target string` – The target to use (**required**)
- `-o, --out-format string` – Output format (`json`, `xml`, etc.)

## Join us

Have questions, ideas, bug reports or just want to chat? Come join [our discord server](https://discord.com/channels/1240272304294985800/1311031796372344894).

## License and Code of Conduct

Code is under the [Apache License 2.0](LICENSE), documentation is [CC BY 4.0](LICENSE-documentation).

The SDC project is following the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md). More information and links about the CNCF Code of Conduct are [here](https://www.cncf.io/conduct/).
