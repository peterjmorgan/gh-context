// ABOUTME: SSH config file parser and manipulator for gh-context
// ABOUTME: Handles reading, modifying, and writing ~/.ssh/config for key switching

package ssh

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DefaultConfigPath returns the default SSH config path.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "config")
}

// HostBlock represents a Host block in SSH config.
type HostBlock struct {
	StartLine    int      // Line number where "Host X" appears (0-indexed)
	EndLine      int      // Line number of last line in block (exclusive)
	Hostname     string   // The hostname pattern from "Host X"
	Lines        []string // All lines in the block including Host line
	IdentityFiles []IdentityFileLine
}

// IdentityFileLine represents an IdentityFile line (commented or not).
type IdentityFileLine struct {
	LineIndex  int    // Index within HostBlock.Lines
	Path       string // The path to the key (without ~ expansion)
	IsCommented bool
	FullLine   string // Original line content
}

// ConfigFile represents a parsed SSH config file.
type ConfigFile struct {
	Path   string
	Lines  []string
	Blocks []HostBlock
}

// ParseConfig reads and parses an SSH config file.
func ParseConfig(path string) (*ConfigFile, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigFile{Path: path, Lines: []string{}, Blocks: []HostBlock{}}, nil
		}
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	cfg := &ConfigFile{
		Path:  path,
		Lines: lines,
	}
	cfg.parseBlocks()
	return cfg, nil
}

// hostPattern matches "Host <pattern>" lines.
var hostPattern = regexp.MustCompile(`(?i)^\s*Host\s+(.+?)\s*$`)

// identityFilePattern matches "IdentityFile <path>" lines (commented or not).
var identityFilePattern = regexp.MustCompile(`(?i)^\s*(#\s*)?(IdentityFile)\s+(.+?)\s*$`)

func (c *ConfigFile) parseBlocks() {
	c.Blocks = nil

	var currentBlock *HostBlock

	for i, line := range c.Lines {
		if match := hostPattern.FindStringSubmatch(line); match != nil {
			// Save previous block
			if currentBlock != nil {
				currentBlock.EndLine = i
				c.Blocks = append(c.Blocks, *currentBlock)
			}
			// Start new block
			currentBlock = &HostBlock{
				StartLine: i,
				Hostname:  strings.TrimSpace(match[1]),
				Lines:     []string{line},
			}
		} else if currentBlock != nil {
			// Add line to current block
			currentBlock.Lines = append(currentBlock.Lines, line)

			// Check for IdentityFile
			if match := identityFilePattern.FindStringSubmatch(line); match != nil {
				ifl := IdentityFileLine{
					LineIndex:   len(currentBlock.Lines) - 1,
					IsCommented: match[1] != "",
					Path:        strings.TrimSpace(match[3]),
					FullLine:    line,
				}
				currentBlock.IdentityFiles = append(currentBlock.IdentityFiles, ifl)
			}
		}
	}

	// Save last block
	if currentBlock != nil {
		currentBlock.EndLine = len(c.Lines)
		c.Blocks = append(c.Blocks, *currentBlock)
	}
}

// FindHostBlock finds a Host block by hostname.
func (c *ConfigFile) FindHostBlock(hostname string) *HostBlock {
	for i := range c.Blocks {
		if c.Blocks[i].Hostname == hostname {
			return &c.Blocks[i]
		}
	}
	return nil
}

// GetActiveIdentityFile returns the currently active (uncommented) IdentityFile for a host.
func (c *ConfigFile) GetActiveIdentityFile(hostname string) string {
	block := c.FindHostBlock(hostname)
	if block == nil {
		return ""
	}

	for _, ifl := range block.IdentityFiles {
		if !ifl.IsCommented {
			return ifl.Path
		}
	}
	return ""
}

// ActivateKey activates a specific SSH key for a hostname by:
// - Uncommenting the IdentityFile line matching keyPath
// - Commenting out all other IdentityFile lines
// Returns error if the key is not found in the config.
func (c *ConfigFile) ActivateKey(hostname, keyPath string) error {
	block := c.FindHostBlock(hostname)
	if block == nil {
		return fmt.Errorf("no Host block found for '%s' in SSH config", hostname)
	}

	// Normalize the key path for comparison
	normalizedKeyPath := normalizePath(keyPath)

	// Check if the key exists in the block
	found := false
	for _, ifl := range block.IdentityFiles {
		if normalizePath(ifl.Path) == normalizedKeyPath {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("IdentityFile '%s' not found in Host %s block\nAdd it to your SSH config first", keyPath, hostname)
	}

	// Now modify the lines
	for _, ifl := range block.IdentityFiles {
		globalLineIdx := block.StartLine + ifl.LineIndex
		originalLine := c.Lines[globalLineIdx]

		if normalizePath(ifl.Path) == normalizedKeyPath {
			// This is the key we want active - uncomment it
			c.Lines[globalLineIdx] = uncommentIdentityFile(originalLine)
		} else {
			// This is a different key - comment it out
			c.Lines[globalLineIdx] = commentIdentityFile(originalLine)
		}
	}

	// Re-parse to update internal state
	c.parseBlocks()
	return nil
}

// AddIdentityFile adds a new IdentityFile line to a Host block.
// If the block doesn't exist, returns an error.
func (c *ConfigFile) AddIdentityFile(hostname, keyPath string, active bool) error {
	block := c.FindHostBlock(hostname)
	if block == nil {
		return fmt.Errorf("no Host block found for '%s' in SSH config", hostname)
	}

	// Check if it already exists
	normalizedKeyPath := normalizePath(keyPath)
	for _, ifl := range block.IdentityFiles {
		if normalizePath(ifl.Path) == normalizedKeyPath {
			return nil // Already exists
		}
	}

	// Create the new line
	indent := detectIndent(block.Lines)
	var newLine string
	if active {
		newLine = fmt.Sprintf("%sIdentityFile %s", indent, keyPath)
	} else {
		newLine = fmt.Sprintf("%s# IdentityFile %s", indent, keyPath)
	}

	// Find insertion point (after last IdentityFile, or after Host line)
	insertIdx := block.StartLine + 1 // Default: right after Host line
	if len(block.IdentityFiles) > 0 {
		lastIF := block.IdentityFiles[len(block.IdentityFiles)-1]
		insertIdx = block.StartLine + lastIF.LineIndex + 1
	}

	// Insert the line
	c.Lines = append(c.Lines[:insertIdx], append([]string{newLine}, c.Lines[insertIdx:]...)...)

	// Re-parse
	c.parseBlocks()
	return nil
}

// Save writes the config back to disk, creating a backup first.
func (c *ConfigFile) Save() error {
	// Create backup
	backupPath := c.Path + ".bak"
	if _, err := os.Stat(c.Path); err == nil {
		data, err := os.ReadFile(c.Path)
		if err != nil {
			return fmt.Errorf("failed to read config for backup: %w", err)
		}
		if err := os.WriteFile(backupPath, data, 0600); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Write new config
	content := strings.Join(c.Lines, "\n")
	if len(c.Lines) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if err := os.WriteFile(c.Path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write SSH config: %w", err)
	}

	return nil
}

// Helper functions

func normalizePath(p string) string {
	// Expand ~ to home directory for comparison
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			p = filepath.Join(home, p[2:])
		}
	}
	// Clean the path
	return filepath.Clean(p)
}

func uncommentIdentityFile(line string) string {
	// Remove leading # and normalize spacing
	match := identityFilePattern.FindStringSubmatch(line)
	if match == nil {
		return line
	}

	// Preserve original indentation
	indent := ""
	for _, ch := range line {
		if ch == ' ' || ch == '\t' {
			indent += string(ch)
		} else {
			break
		}
	}

	return fmt.Sprintf("%sIdentityFile %s", indent, strings.TrimSpace(match[3]))
}

func commentIdentityFile(line string) string {
	match := identityFilePattern.FindStringSubmatch(line)
	if match == nil {
		return line
	}

	// Already commented
	if match[1] != "" {
		return line
	}

	// Preserve original indentation
	indent := ""
	for _, ch := range line {
		if ch == ' ' || ch == '\t' {
			indent += string(ch)
		} else {
			break
		}
	}

	return fmt.Sprintf("%s# IdentityFile %s", indent, strings.TrimSpace(match[3]))
}

func detectIndent(lines []string) string {
	// Look at existing lines to detect indentation style
	for _, line := range lines[1:] { // Skip Host line
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed != "" && trimmed != line {
			indent := line[:len(line)-len(trimmed)]
			return indent
		}
	}
	return "    " // Default to 4 spaces
}

// ExpandPath expands ~ in a path to the home directory.
func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

// KeyExists checks if an SSH key file exists.
func KeyExists(keyPath string) bool {
	expanded := ExpandPath(keyPath)
	_, err := os.Stat(expanded)
	return err == nil
}
