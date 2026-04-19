---
title: gluon CLI
---

The root `gluon` command is the entry point for planning, inspection, and execution.

## Command map

| Command | Purpose |
| --- | --- |
| `gluon plan` | Compile intent and compositions into a deterministic execution plan |
| `gluon run` | Dry-run or execute a compiled plan |
| `gluon validate` | Validate intent and discovered components against schemas |
| `gluon debug` | Inspect intent processing and planning internals |
| `gluon compositions` | List or inspect available compositions |
| `gluon component` | List components or inspect a merged component view |
| `gluon completion` | Generate shell completion scripts |

## Global flags

| Flag | Meaning |
| --- | --- |
| `--config-dir`, `-c` | Path or glob used to load composition assets |
| `--version` | Print the CLI version |
| `--help` | Show command help |

`--config-dir` can also be set through `GLUON_CONFIG_DIR`.

## Typical flow

```bash
gluon validate --intent intent.yaml --config-dir assets/config/compositions
gluon plan --intent intent.yaml --config-dir assets/config/compositions --output plan.json
gluon run --plan plan.json
```

Read the command-specific pages next if you need examples and flag details.