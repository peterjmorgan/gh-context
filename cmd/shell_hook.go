// ABOUTME: Shell-hook command for gh-context - generates shell integration code
// ABOUTME: Supports bash, zsh, PowerShell, and fish shells for auto-apply on cd

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var shellHookCmd = &cobra.Command{
	Use:   "shell-hook [shell]",
	Short: "Print shell snippet for auto-apply on cd",
	Long: `Print shell integration code that automatically applies context when entering a repo with .ghcontext.

Supported shells: bash, zsh, powershell, pwsh, fish

Examples:
  gh context shell-hook bash >> ~/.bashrc
  gh context shell-hook zsh >> ~/.zshrc
  gh context shell-hook powershell >> $PROFILE
  gh context shell-hook fish >> ~/.config/fish/config.fish

If no shell is specified, outputs bash/zsh compatible code.`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"bash", "zsh", "powershell", "pwsh", "fish"},
	RunE:      runShellHook,
}

func runShellHook(cmd *cobra.Command, args []string) error {
	shell := "bash" // Default
	if len(args) > 0 {
		shell = args[0]
	}

	var hook string
	switch shell {
	case "bash":
		hook = bashHook()
	case "zsh":
		hook = zshHook()
	case "powershell", "pwsh":
		hook = powershellHook()
	case "fish":
		hook = fishHook()
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, powershell, pwsh, fish)", shell)
	}

	fmt.Print(hook)
	return nil
}

func bashHook() string {
	return `# gh-context: Auto-apply context when entering a repo with .ghcontext
# Add this to your ~/.bashrc

__gh_context_auto_apply() {
  local root
  root="$(git rev-parse --show-toplevel 2>/dev/null)" || return 0

  if [[ -f "$root/.ghcontext" ]]; then
    local name current
    name="$(cat "$root/.ghcontext")"
    current=""
    [[ -f "${XDG_CONFIG_HOME:-$HOME/.config}/gh/contexts/active" ]] && \
      current="$(cat "${XDG_CONFIG_HOME:-$HOME/.config}/gh/contexts/active")"

    if [[ "$current" != "$name" ]]; then
      echo "• Auto-applying gh context: $name"
      gh context use "$name" 2>/dev/null || true
    fi
  fi
}

PROMPT_COMMAND="__gh_context_auto_apply${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
`
}

func zshHook() string {
	return `# gh-context: Auto-apply context when entering a repo with .ghcontext
# Add this to your ~/.zshrc

__gh_context_auto_apply() {
  local root
  root="$(git rev-parse --show-toplevel 2>/dev/null)" || return 0

  if [[ -f "$root/.ghcontext" ]]; then
    local name current
    name="$(cat "$root/.ghcontext")"
    current=""
    [[ -f "${XDG_CONFIG_HOME:-$HOME/.config}/gh/contexts/active" ]] && \
      current="$(cat "${XDG_CONFIG_HOME:-$HOME/.config}/gh/contexts/active")"

    if [[ "$current" != "$name" ]]; then
      echo "• Auto-applying gh context: $name"
      gh context use "$name" 2>/dev/null || true
    fi
  fi
}

autoload -U add-zsh-hook
add-zsh-hook precmd __gh_context_auto_apply
`
}

func powershellHook() string {
	return `# gh-context: Auto-apply context when entering a repo with .ghcontext
# Add this to your PowerShell profile ($PROFILE)

function Invoke-GhContextAutoApply {
    $root = git rev-parse --show-toplevel 2>$null
    if (-not $root) { return }

    $ghContextFile = Join-Path $root ".ghcontext"
    if (Test-Path $ghContextFile) {
        $name = (Get-Content $ghContextFile -Raw).Trim()

        # Get current active context
        $configDir = if ($env:XDG_CONFIG_HOME) { $env:XDG_CONFIG_HOME } else { "$env:APPDATA" }
        $activeFile = Join-Path $configDir "gh\contexts\active"
        $current = ""
        if (Test-Path $activeFile) {
            $current = (Get-Content $activeFile -Raw).Trim()
        }

        if ($current -ne $name) {
            Write-Host "• Auto-applying gh context: $name"
            gh context use $name 2>$null
        }
    }
}

# Hook into prompt
$__ghContextOriginalPrompt = $function:prompt
function prompt {
    Invoke-GhContextAutoApply
    & $__ghContextOriginalPrompt
}
`
}

func fishHook() string {
	return `# gh-context: Auto-apply context when entering a repo with .ghcontext
# Add this to your ~/.config/fish/config.fish

function __gh_context_auto_apply --on-variable PWD
    set -l root (git rev-parse --show-toplevel 2>/dev/null)
    if test -z "$root"
        return
    end

    set -l ghcontext_file "$root/.ghcontext"
    if test -f $ghcontext_file
        set -l name (cat $ghcontext_file | string trim)

        # Get current active context
        set -l config_dir
        if test -n "$XDG_CONFIG_HOME"
            set config_dir $XDG_CONFIG_HOME
        else
            set config_dir ~/.config
        end

        set -l active_file "$config_dir/gh/contexts/active"
        set -l current ""
        if test -f $active_file
            set current (cat $active_file | string trim)
        end

        if test "$current" != "$name"
            echo "• Auto-applying gh context: $name"
            gh context use $name 2>/dev/null
        end
    end
end
`
}
