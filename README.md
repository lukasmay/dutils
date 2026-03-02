# dutils

Docker and Docker Compose workflow simplifier.

## Installation

### Local (Homebrew)

Homebrew requires formulas to be in a tap. You can create a local tap to test the installation:

```bash
# 1. Create a local tap
brew tap-new lukasmay/dutils

# 2. Link your local formula
cp Formula/dutils.rb $(brew --repository lukasmay/dutils)/Formula/dutils.rb

# 3. Install
brew install --build-from-source lukasmay/dutils/dutils
```

### Local (Makefile)

```bash
make
sudo make install
```

## Features

- **Project Management**: Switch between different Docker Compose projects.
- **Service Groups**: Define groups of services in `.dutils.yml` and manage them with `@group`.
- **Enhanced Commands**:
  - `dutils list`: Filters containers by project.
  - `dutils start`: Supports service groups and builds.
  - `dutils stop`: Supports service groups and `down`.
  - `dutils restart`: Rebuilds and force-recreates services.
- **Autocompletion**: Full tab-completion for containers, services, groups, and projects.

## Configuration

Add a `.dutils.yml` to your project root:

```yaml
project_name: my-app
groups:
  frontend:
    - web
    - nginx
  backend:
    - api
    - db
compose:
  files:
    - docker-compose.yml
    - docker-compose.override.yml
```
