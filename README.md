# gh-context

A kubectx-style context switcher for GitHub CLI - manage multiple GitHub accounts easily.

## What It Does

When you switch contexts, `gh-context`:
1. Sets the active context
2. **Updates `~/.ssh/config`** to use the correct SSH key for that account
3. Switches the `gh` CLI authentication to the correct user

This means both `git push` and `gh` commands will use the right credentials automatically.

## Installation

```bash
gh extension install pmorgan/gh-context
```

## Prerequisites: SSH Config Setup

Before using gh-context, set up your `~/.ssh/config` with your different SSH keys:

```
Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_work
    # IdentityFile ~/.ssh/id_personal
```

**Key points:**
- All IdentityFile lines for github.com should be in a single `Host github.com` block
- Comment out the keys you're not currently using with `#`
- `gh-context` will uncomment/comment these lines when switching contexts

## Quick Start

```bash
# Create a context from your current session (auto-detects active SSH key)
gh context new --from-current --name work

# Create a context with explicit SSH key
gh context new --from-current --name personal --ssh-key ~/.ssh/id_personal

# Switch contexts (updates SSH config + gh auth)
gh context use personal

# Verify everything is set up
gh context auth-status
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List all contexts with active indicator |
| `current` | Show active context and repo-bound context |
| `new` | Create a new context |
| `use <name>` | Switch to a context (updates SSH config + gh auth) |
| `delete <name>` | Remove a saved context |
| `bind <name>` | Bind current repository to a context |
| `unbind` | Remove repository binding |
| `apply` | Apply the repo's bound context |
| `shell-hook [shell]` | Print shell integration code |
| `auth-status` | Show authentication status for all contexts |

## Creating Contexts

### From Current Session
```bash
# Auto-detect user and SSH key from current state
gh context new --from-current --name work

# Override the SSH key
gh context new --from-current --name personal --ssh-key ~/.ssh/id_personal
```

### With Explicit Parameters
```bash
gh context new \
  --hostname github.com \
  --user myusername \
  --ssh-key ~/.ssh/id_mykey \
  --name mycontext
```

## How SSH Key Switching Works

When you run `gh context use personal`, the tool:

1. Finds the `Host github.com` block in `~/.ssh/config`
2. Comments out all `IdentityFile` lines
3. Uncomments the `IdentityFile` line matching your context's SSH key
4. Creates a backup at `~/.ssh/config.bak`

**Before:**
```
Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_work
    # IdentityFile ~/.ssh/id_personal
```

**After `gh context use personal`:**
```
Host github.com
    HostName github.com
    User git
    # IdentityFile ~/.ssh/id_work
    IdentityFile ~/.ssh/id_personal
```

## Repository Binding

Bind repositories to contexts for automatic switching:

```bash
# In your work repo
cd ~/work/project
gh context bind work

# In your personal repo
cd ~/personal/project
gh context bind personal
```

## Shell Integration

Add automatic context switching when entering repositories:

### Bash
```bash
gh context shell-hook bash >> ~/.bashrc
source ~/.bashrc
```

### Zsh
```bash
gh context shell-hook zsh >> ~/.zshrc
source ~/.zshrc
```

### PowerShell
```powershell
gh context shell-hook powershell >> $PROFILE
. $PROFILE
```

### Fish
```fish
gh context shell-hook fish >> ~/.config/fish/config.fish
source ~/.config/fish/config.fish
```

## Context File Format

Contexts are stored in `~/.config/gh/contexts/` (or `%APPDATA%\gh\contexts` on Windows):

```
HOSTNAME=github.com
USER=myuser
TRANSPORT=ssh
SSH_KEY=~/.ssh/id_personal
```

## Full Setup Example

```bash
# 1. Set up SSH keys (if not already done)
ssh-keygen -t ed25519 -f ~/.ssh/id_work -C "work@company.com"
ssh-keygen -t ed25519 -f ~/.ssh/id_personal -C "personal@gmail.com"

# 2. Add keys to GitHub accounts (via GitHub web UI)

# 3. Set up ~/.ssh/config
cat >> ~/.ssh/config << 'EOF'
Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_work
    # IdentityFile ~/.ssh/id_personal
EOF

# 4. Log in to both accounts with gh
gh auth login  # for work account
gh auth login  # for personal account (add second account)

# 5. Create contexts
gh context new --from-current --name work --ssh-key ~/.ssh/id_work
gh context new --hostname github.com --user personal-username --ssh-key ~/.ssh/id_personal --name personal

# 6. Switch and verify
gh context use personal
gh context auth-status
git push  # Now uses personal SSH key
```

## Troubleshooting

### "IdentityFile not found in Host block"
Make sure your `~/.ssh/config` has a `Host github.com` block with the IdentityFile lines:
```
Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_work
    # IdentityFile ~/.ssh/id_personal
```

### SSH key not switching
- Check `~/.ssh/config` was updated: `cat ~/.ssh/config`
- Verify backup exists: `ls -la ~/.ssh/config.bak`
- Run `gh context auth-status` to see current state

### Wrong account being used
- Run `gh context auth-status` to check both GH Auth and SSH Active status
- Make sure both show ✅ for the context you want to use

## Building from Source

```bash
go build -o gh-context .
```

## License

MIT
