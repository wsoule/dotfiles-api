# 🚀 Dotfiles Manager - Developer Onboarding Toolkit

> The complete solution for new developer environment setup

A comprehensive command-line tool that manages your entire development environment: Homebrew packages, dotfiles with GNU Stow, GitHub SSH setup, and complete developer onboarding automation.

## Features

- 🎉 **Complete Developer Onboarding** - One command to set up everything for new developers
- 🔐 **GitHub SSH Integration** - Automated SSH key generation and GitHub configuration
- 📦 **Smart Package Management** - Curated essential packages for developers
- 🤝 **Configuration Sharing** - Share and discover configs via GitHub Gist
- 📚 **Template Library** - Pre-built configs for web-dev, data science, DevOps, and more
- 🌍 **Community Discovery** - Find and browse configurations from other developers
- 📋 **JSON Configuration** - Store your setup in simple, versionable JSON
- 🍺 **Brewfile Support** - Generate and import Brewfiles with full automation
- 🔗 **GNU Stow Integration** - Manage dotfiles with symbolic links
- 📊 **Status Checking** - Verify installation status across all components
- 💾 **Backup & Restore** - Save and restore complete configurations
- 🔍 **Multiple Output Formats** - JSON, count, and filtered views
- 📦 **Easy Distribution** - Install via Homebrew or direct from GitHub
- ✨ **Zero Configuration** - Works out of the box with sensible defaults
- 🚀 **Fast & Lightweight** - Single binary, comprehensive functionality

## 📦 Installation

### 🍺 Homebrew (Recommended)
```bash
brew tap wsoule/tap
brew install dotfiles
```

### 🚀 One-liner Install
```bash
curl -fsSL https://raw.githubusercontent.com/wsoule/go-dotfiles/main/install.sh | bash
```

### 📦 GitHub Releases
Download the latest binary from [releases](https://github.com/wsoule/go-dotfiles/releases/latest)

### 🔨 Build from Source
```bash
git clone https://github.com/wsoule/go-dotfiles.git
cd go-dotfiles
go build -o dotfiles
```

## 🚀 Quick Start

### 🆕 Setup from a Repository (Recommended)
```bash
# Setup dotfiles from your GitHub repository
dotfiles setup https://github.com/your-username/your-dotfiles.git

# This will:
# - Clone your repo to ~/.dotfiles/
# - Create a stow/ directory for your packages
# - Create a private/ directory for sensitive files
# - Set up proper directory structure
```

### 🎯 For New Developers (Complete Setup)
```bash
# Complete onboarding with everything included
dotfiles onboard

# Or with your GitHub email
dotfiles onboard --email your@email.com
```

This single command will:
- Initialize your dotfiles configuration
- Set up GitHub SSH keys
- Install essential development packages
- Guide you through next steps

### 🔧 Manual Setup (Step by Step)

1. **Initialize configuration:**
   ```bash
   dotfiles init
   ```

2. **Set up GitHub SSH:**
   ```bash
   dotfiles github setup --email your@email.com
   dotfiles github test  # Verify connection
   ```

3. **Add packages:**
   ```bash
   dotfiles add git lazygit curl
   dotfiles add --type=cask visual-studio-code
   dotfiles add --type=stow vim zsh tmux
   ```

4. **Install everything:**
   ```bash
   dotfiles install      # Installs Homebrew packages
   dotfiles stow vim zsh # Creates dotfile symlinks
   ```

## 📋 Available Commands

```
Usage:
  dotfiles [command]

Available Commands:
  setup       Fork and setup a dotfiles repository                 🆕 NEW!
  onboard     Complete developer onboarding and environment setup  🎯 NEW!
  github      Set up GitHub with SSH keys                          🔐 NEW!
  private     Add private files to stow packages                   🔒 NEW!
  share       Share your configuration with others                 🤝 NEW!
  clone       Clone a shared configuration                         📥 NEW!
  templates   Create, manage and discover templates                📚 ENHANCED!
  discover    Discover shared configurations from community        🔍 NEW!
  add         Add packages to your configuration
  backup      Backup your configuration to a file
  brewfile    Generate a Brewfile from your configuration
  import      Import packages from a Brewfile
  init        Initialize a new dotfiles configuration
  install     Generate Brewfile and install packages
  list        List all packages in your configuration
  remove      Remove packages from your configuration
  restore     Restore configuration from a backup file
  restow      Restow dotfile packages (unstow then stow)
  status      Check package installation status
  stow        Stow dotfile packages using GNU Stow
  unstow      Unstow dotfile packages using GNU Stow
```

## 🎯 Developer Onboarding

Perfect for new developers or setting up fresh machines:

### What the onboard command does:
1. **🔧 Initializes** your dotfiles configuration
2. **🔐 Creates** GitHub SSH keys and shows setup instructions
3. **📦 Installs** curated essential packages:
   - **Development tools**: git, curl, wget, tree, jq, gh
   - **Applications**: Visual Studio Code, iTerm2, Rectangle
   - **GNU Stow** for dotfiles management
4. **📋 Guides** you through next steps

### Essential packages included:
```bash
# Taps
homebrew/cask-fonts

# Brews (Command-line tools)
git, curl, wget, tree, jq, stow, gh

# Casks (GUI Applications)
visual-studio-code, ghostty, raycast
```

You can customize this list after onboarding with `dotfiles add` and `dotfiles remove`.

## 🤝 Configuration Sharing & Templates

Share your perfect development setup with others and discover configurations from the community!

### 🎯 Using Templates (Quick Start)

Choose from pre-built templates for different development workflows:

```bash
# Browse available templates
dotfiles templates list

# Preview a template
dotfiles templates show web-dev

# Apply a template
dotfiles clone template:web-dev

# Discover community templates 🆕
dotfiles templates discover --search "web-dev" --featured

# Browse all featured templates 🆕
dotfiles templates discover --featured
```

**Available Built-in Templates:**
- `web-dev` - Web development with Node.js, Python, Docker
- `mobile-dev` - iOS/Android with Flutter, React Native
- `data-science` - Python, R, Jupyter, analytics tools
- `devops` - Kubernetes, Terraform, cloud tools
- `minimal` - Essential tools only

### 📚 Create & Share Templates 🆕

Create reusable templates for the community:

```bash
# Create a new template with inheritance
dotfiles templates create my-web-setup --extends web-dev --add-only

# Create and immediately push to API
dotfiles templates create my-setup --push --public --featured

# Push existing template to API
dotfiles templates push my-template.json --public

# Validate template structure
dotfiles templates validate my-template.json
```

**🤔 Templates vs Configurations - What's the Difference?**

| **Configuration** | **Template** |
|------------------|-------------|
| Your current dotfiles setup | Reusable blueprint for others |
| Generated from `~/.dotfiles/config.json` | Created with inheritance & customization |
| Shared via GitHub Gist/files | Shared via community API |
| "Here's my current setup" | "Here's a starter kit you can build upon" |
| Use: `dotfiles share gist` | Use: `dotfiles templates create` |

Templates support inheritance (`--extends`), additive mode (`--add-only`), and community features like search, tags, and featured status.

### 📤 Share Your Configuration

Share your current setup with the community:

```bash
# Share via GitHub Gist (public)
dotfiles share gist --name="My Web Dev Setup" --description="Full-stack config" --author="YourName"

# Share + push as template to API 🆕
dotfiles share gist --name="My Config" --api --featured

# Share privately
dotfiles share gist --name="My Config" --private

# Export to file
dotfiles share file my-config.json --name="My Setup"
```

### 📥 Import Shared Configurations

Import configurations from others:

```bash
# From GitHub Gist URL
dotfiles clone https://gist.github.com/user/gist-id

# From local file
dotfiles clone my-config.json

# Preview before importing
dotfiles clone https://gist.github.com/user/gist-id --preview

# Merge with existing config
dotfiles clone https://gist.github.com/user/gist-id --merge
```

### 🔍 Discover Community Configs

Find configurations shared by other developers:

```bash
# Browse featured configs
dotfiles discover featured

# Search by topic
dotfiles discover search web-development
dotfiles discover search --tags=python,data-science

# View sharing statistics
dotfiles discover stats
```

## 💡 Usage Examples

### Adding different types of packages
```bash
# Add brew packages (default)
./dotfiles add git curl wget

# Add casks (GUI applications)
./dotfiles add --type=cask visual-studio-code firefox slack

# Add taps (additional repositories)
./dotfiles add --type=tap homebrew/cask-fonts

# Add Stow packages (dotfiles)
./dotfiles add --type=stow vim zsh tmux

# Add packages from file
./dotfiles add --file=packages.txt --type=brew
```

### Removing packages
```bash
# Remove brew packages
./dotfiles remove wget

# Remove casks
./dotfiles remove --type=cask firefox

# Remove taps
./dotfiles remove --type=tap homebrew/cask-fonts

# Remove Stow packages
./dotfiles remove --type=stow vim

# Bulk remove all of a type
./dotfiles remove --all-brews
./dotfiles remove --all-casks
./dotfiles remove --all-stow
```

### Working with Brewfiles
```bash
# Generate and install packages automatically
./dotfiles install

# Generate Brewfile in current directory
./dotfiles brewfile

# Generate Brewfile in specific location
./dotfiles brewfile --output ~/my-brewfile

# Import from existing Brewfile
./dotfiles import ~/existing-Brewfile

# Install packages from generated Brewfile
brew bundle --file=./Brewfile
```

### Managing Dotfiles with Stow
```bash
# Create symlinks for dotfiles
./dotfiles stow vim zsh tmux

# Remove symlinks
./dotfiles unstow vim

# Restow (remove and recreate symlinks)
./dotfiles restow vim

# Use custom directories
./dotfiles stow --dir=/path/to/dotfiles --target=~ vim

# Dry run to see what would happen
./dotfiles stow --dry-run --verbose vim
```

### Working with Private Files
```bash
# First, move your private file to the private directory
mv ~/.env-private.sh ~/.dotfiles/private/.env-private.sh

# Create a symlink in a stow package that points to the private file
./dotfiles private shell .env-private.sh

# Now stow the package to create the final symlink
./dotfiles stow shell

# This creates: ~/.env-private.sh -> ~/.dotfiles/stow/shell/.env-private.sh -> ~/.dotfiles/private/.env-private.sh
```

### Status and Backup Operations
```bash
# Check installation status of all packages
./dotfiles status

# List packages in different formats
./dotfiles list --json
./dotfiles list --count
./dotfiles list --type=stow

# Backup configuration
./dotfiles backup ~/my-backup.json

# Restore from backup
./dotfiles restore ~/my-backup.json
```

## 📁 Configuration

Your configuration is stored as simple JSON at `~/.dotfiles/config.json`:

```json
{
  "brews": [
    "git",
    "lazygit"
  ],
  "casks": [
    "visual-studio-code"
  ],
  "taps": [
    "homebrew/cask-fonts"
  ],
  "stow": [
    "vim",
    "zsh",
    "tmux"
  ]
}
```

This generates a Brewfile like:

```ruby
tap "homebrew/cask-fonts"

brew "git"
brew "lazygit"

cask "visual-studio-code"
```

### Directory Structure

The `dotfiles setup` command creates an organized directory structure at `~/.dotfiles/`:

```
~/.dotfiles/
├── config.json          # Your package configuration
├── .gitignore           # Excludes private/ directory
├── stow/                # Stow packages directory
│   ├── config/          # Auto-created .config package
│   │   └── .config/
│   ├── vim/
│   │   ├── .vimrc
│   │   └── .vim/
│   │       └── ... (vim config files)
│   ├── zsh/
│   │   ├── .zshrc
│   │   ├── .zprofile
│   │   └── .zsh/
│   │       └── ... (zsh config files)
│   └── tmux/
│       └── .tmux.conf
└── private/             # Private files (excluded from git)
    ├── .env.local       # Local environment variables
    ├── .gitconfig.local # Personal git config
    └── .ssh/            # SSH keys and config
```

### Stow Integration

When you run `dotfiles stow vim`, it will create symlinks from the `stow/` directory:
- `~/.vimrc` → `~/.dotfiles/stow/vim/.vimrc`
- `~/.vim/` → `~/.dotfiles/stow/vim/.vim/`

For `.config` files, use the auto-created `config` package:
- Put files in `~/.dotfiles/stow/config/.config/`
- Run `dotfiles stow config` to symlink them

## 📚 Command Reference

| Command | Description | Key Flags |
|---------|-------------|-----------|
| `dotfiles setup <repo-url>` 🆕 | Fork and setup dotfiles repository | `--force` |
| `dotfiles onboard` ⭐ | Complete developer setup | `--email=<email>`, `--skip-*` |
| `dotfiles github setup` 🔐 | Set up GitHub SSH keys | `--email=<email>`, `--key-type=<type>` |
| `dotfiles github test` 🧪 | Test GitHub SSH connection | None |
| `dotfiles private <package> <file>` 🔒 | Add private file to stow package | `--dry-run`, `--verbose` |
| `dotfiles templates list` 📚 | Browse configuration templates | None |
| `dotfiles templates show <name>` | Preview template details | None |
| `dotfiles templates create <name>` 🆕 | Create new template | `--extends=<template>`, `--add-only`, `--push`, `--public`, `--featured` |
| `dotfiles templates push <file>` 🆕 | Push template to API | `--public`, `--featured` |
| `dotfiles templates discover` 🆕 | Discover community templates | `--search=<term>`, `--tags=<tags>`, `--featured` |
| `dotfiles templates validate <file>` 🆕 | Validate template structure | None |
| `dotfiles share gist` 🤝 | Share config via GitHub Gist | `--name=<name>`, `--description`, `--private`, `--api`, `--featured` |
| `dotfiles share file <path>` | Export config to file | `--name=<name>`, `--description` |
| `dotfiles clone <source>` 📥 | Import shared configuration | `--merge`, `--preview` |
| `dotfiles discover search` 🔍 | Search community configs | `--tags=<tags>` |
| `dotfiles discover featured` | Browse featured configs | None |
| `dotfiles init` | Initialize new configuration | None |
| `dotfiles add <packages>` | Add packages to config | `--type=brew/cask/tap/stow`, `--file=<path>` |
| `dotfiles remove <packages>` | Remove packages from config | `--type=brew/cask/tap/stow`, `--all-*`, `--file=<path>` |
| `dotfiles list` | List configured packages | `--json`, `--count`, `--type=<type>` |
| `dotfiles status` | Check package installation status | None |
| `dotfiles install` | Generate Brewfile and install | `--dry-run` |
| `dotfiles brewfile` | Generate Brewfile | `--output=<path>` |
| `dotfiles import <brewfile>` | Import from Brewfile | `--replace` |
| `dotfiles backup <file>` | Backup configuration | None |
| `dotfiles restore <file>` | Restore from backup | `--no-backup` |
| `dotfiles stow <packages>` | Create symlinks with Stow | `--dir=<path>`, `--target=<path>`, `--dry-run`, `--verbose` |
| `dotfiles unstow <packages>` | Remove symlinks | `--dir=<path>`, `--target=<path>`, `--all`, `--keep-config` |
| `dotfiles restow <packages>` | Restow (unstow + stow) | `--dir=<path>`, `--target=<path>`, `--all` |

## 🛠 Development

### Prerequisites
- Go 1.25.1 or later
- Homebrew (for package management features)
- GNU Stow (for dotfiles symlinking features): `brew install stow`

### Building
```bash
go build -o dotfiles
```

### Project Structure
```
Go_Dotfiles/
├── cmd/                 # CLI commands
│   ├── root.go         # Root command
│   ├── init.go         # Initialize config
│   ├── add.go          # Add/remove packages
│   ├── list.go         # List packages
│   ├── status.go       # Status checking
│   ├── install.go      # Install packages
│   ├── brewfile.go     # Generate Brewfile
│   ├── import.go       # Import from Brewfile
│   ├── backup.go       # Backup/restore
│   └── stow.go         # GNU Stow integration
├── internal/config/    # Configuration management
│   └── config.go       # JSON config handling
└── main.go             # Entry point
```

## 📦 Publishing & Distribution

### Homebrew Tap
This package will be available via a custom Homebrew tap:

```bash
brew tap wsoule/tap
brew install dotfiles
```

### GitHub Releases
Automated releases are created for every tagged version with binaries for:
- macOS (Intel & Apple Silicon)
- Linux (x64 & ARM64)
- Windows (x64)

### Installation Script
The `install.sh` script automatically detects your platform and installs the appropriate binary.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 License

MIT License
