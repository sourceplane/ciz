---
title: gluon compositions
---

`gluon compositions` lists or inspects the composition types available under the configured compositions directory.

## Usage

```bash
gluon compositions --config-dir assets/config/compositions
```

The command also accepts a composition name directly:

```bash
gluon compositions helm --config-dir assets/config/compositions
```

The alias `composition` is also supported.

## Subcommand form

For detailed output, use the explicit `list` subcommand:

```bash
gluon compositions list helm \
  --config-dir assets/config/compositions \
  --long \
  --expand-jobs
```

## Flags

| Flag | Meaning |
| --- | --- |
| `--expand-jobs`, `-e` | Expand job details in the output |
| `--long`, `-l` | Detailed listing mode on `gluon compositions list` |
| `--config-dir`, `-c` | Global flag used to find compositions |

Use this command to confirm which types are available before validating or planning against them.