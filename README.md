# dutils

![Tests](https://github.com/lukasmay/dutils/actions/workflows/test.yml/badge.svg)

Docker and Docker Compose workflow simplifier.

## Installation

### macOS (Homebrew)

```bash
brew tap lukasmay/dutils
brew install dutils
```

Then enable shell completions:

```bash
# Zsh
echo 'eval "$(dutils completion zsh)"' >> ~/.zshrc

# Bash
echo 'eval "$(dutils completion bash)"' >> ~/.bashrc

# Fish
dutils completion fish > ~/.config/fish/completions/dutils.fish
```

### Linux (Debian / Ubuntu)

Download the `.deb` from the [latest release](https://github.com/lukasmay/dutils/releases/latest):

```bash
wget https://github.com/lukasmay/dutils/releases/latest/download/dutils_Linux_x86_64.deb
sudo dpkg -i dutils_Linux_x86_64.deb
```

For ARM64: replace `x86_64` with `arm64`.

### Windows

Download the `.zip` from the [latest release](https://github.com/lukasmay/dutils/releases/latest), extract `dutils.exe`, and place it somewhere on your `PATH`.

Or with PowerShell:

```powershell
$release = Invoke-RestMethod "https://api.github.com/repos/lukasmay/dutils/releases/latest"
$asset = $release.assets | Where-Object { $_.name -like "*Windows_x86_64*" }
Invoke-WebRequest $asset.browser_download_url -OutFile "$env:TEMP\dutils.zip"
Expand-Archive "$env:TEMP\dutils.zip" -DestinationPath "$HOME\bin" -Force
```

### From Source

```bash
# Builds and installs the binary to ~/go/bin/dutils
make dev
```

Make sure `~/go/bin` is in your `$PATH`.

## Features

- **Container List**: `dutils ps` shows running containers in a clean table. Use `-a` to include stopped containers.
- **Project Management**: Register Docker Compose projects globally and switch between them from anywhere.
  - `dutils project init`: Create a `.dutils.yml` in the current directory and register it as the active project.
  - `dutils project add [path]`: Register an existing project directory (must already have a `.dutils.yml`).
  - `dutils project list`: List all registered projects, with the active one marked with `*`.
  - `dutils project switch <name>`: Set the active project so all `dutils` commands target it from any directory.
  - `dutils project status`: Show which project is currently active.
  - `dutils project clear`: Clear the active project override (falls back to current directory).
- **Service Groups**: Define logical groups of services in `.dutils.yml` (e.g., `@frontend`) and manage them as one.
- **Enhanced Commands**:
  - `dutils start [services|@groups]`: Start services or groups. Use `-b` to rebuild with `--no-cache` first.
  - `dutils stop [services|@groups]`: Stop services or groups. Use `-d` to remove containers instead of just stopping.
  - `dutils restart [services|@groups]`: Rebuild with `--no-cache` and force-recreate services.
- **Autocompletion**: Tab-completion suggests running containers, compose services, config groups, and registered project names.

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
  ps:
    scope: all # Example: limit list commands to specific scopes
```
