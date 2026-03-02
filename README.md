# dutils

Docker and Docker Compose workflow simplifier.

## Installation

### Homebrew (macOS / Linux)

The easiest way to install `dutils` is via Homebrew:

```bash
brew tap lukasmay/dutils
brew install dutils
```

*Note: Homebrew automatically installs the required autocomplete scripts for Zsh, Bash, and Fish.*

### Local Development (From Source)

If you want to build and test features locally:

```bash
# Builds and installs the binary to ~/go/bin/dutils
make dev
```
*(Make sure `~/go/bin` is in your `$PATH`!)*

## Features

## Features

- **Project Management**: Switch between different Docker Compose projects globally.
  - `dutils project init`: Initialize a new project and add it to your global registry.
  - `dutils project list`: List all registered projects.
  - `dutils project status`: See which project is currently active globally.
  - `dutils project switch <name>`: Set your active project so `dutils` commands run against it from anywhere.
  - `dutils project clear`: Unset the active project (default to current directory).
- **Service Groups**: Define logical groups of services in `.dutils.yml` (e.g., `@frontend`) and manage them simultaneously.
- **Enhanced Commands**:
  - `dutils start`: Supports starting individual services, `@groups`, and building (`-b`).
  - `dutils stop`: Supports stopping service groups and taking down whole environments (`-d`).
  - `dutils restart`: Rebuilds and force-recreates services or groups.
- **Autocompletion**: Robust tab-completion dynamically suggests active containers, valid compose services, config groups, and registered projects.

## Configuration

Add a `.dutils.yml` to your project root:

```yaml
# Used to group your containers visually
project_name: my-app

# Define logical groups to control multiple services at once via @groupname
groups:
  frontend:
    - web
    - nginx
  backend:
    - api
    - db

# Explicitly declare which compose files make up your project
compose:
  files:
    - docker-compose.yml
    - docker-compose.override.yml

# Customize default behaviors for various commands
defaults:
  dlist:
    scope: all # Example: limit list commands to specific scopes
```
