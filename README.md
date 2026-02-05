# gf - GitFlic CLI

`gf` is a command-line tool for working with [GitFlic](https://gitflic.ru) — a Russian code hosting platform.

## Installation

### From releases

Download the latest binary from [Releases](https://github.com/josinSbazin/gf/releases).

### From source

```bash
go install github.com/josinSbazin/gf@latest
```

## Quick Start

```bash
# Authenticate
gf auth login

# List merge requests
gf mr list

# View a merge request
gf mr view 12

# Create a merge request
gf mr create --title "Add feature" --target main

# Watch a pipeline
gf pipeline watch 45
```

## Commands

### Authentication

```bash
gf auth login              # Login to GitFlic
gf auth login -h company   # Login to self-hosted instance
gf auth status             # Check authentication status
```

### Merge Requests

```bash
gf mr list                 # List open merge requests
gf mr list --state all     # List all merge requests
gf mr view <id>            # View merge request details
gf mr create               # Create a merge request (interactive)
gf mr merge <id>           # Merge a merge request
```

### Pipelines

```bash
gf pipeline list           # List pipelines
gf pipeline view <id>      # View pipeline and jobs
gf pipeline watch <id>     # Watch pipeline in real-time
```

### Repository

```bash
gf repo view               # View current repository info
gf repo view owner/name    # View specific repository
```

## Configuration

Configuration is stored in `~/.gf/config.json`.

```json
{
  "version": 1,
  "active_host": "gitflic.ru",
  "hosts": {
    "gitflic.ru": {
      "token": "your-token-here",
      "user": "username"
    }
  }
}
```

## Environment Variables

- `GF_REPO` - Override repository detection (format: `owner/repo`)
- `GF_TOKEN` - Override authentication token

## License

MIT License. See [LICENSE](LICENSE) for details.
