// ABOUTME: GitHub CLI authentication operations for gh-context
// ABOUTME: Wraps gh auth commands for testing and switching accounts

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
)

// TestAuth checks if the given user is authenticated on the given host.
// Returns true if authentication is valid and ready to use.
func TestAuth(hostname, user string) (bool, error) {
	// Check if the user has authentication for this host
	stdout, _, err := gh.Exec("auth", "status", "--hostname", hostname)
	if err != nil {
		return false, nil // Not authenticated at all
	}

	output := stdout.String()
	expectedPattern := fmt.Sprintf("Logged in to %s account %s", hostname, user)
	if !strings.Contains(output, expectedPattern) {
		return false, nil // Different user or not logged in
	}

	// Try to switch to the user
	_, _, err = gh.Exec("auth", "switch", "--hostname", hostname, "--user", user)
	if err != nil {
		return false, nil // Switch failed
	}

	// Verify with a quick API call
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	currentUser, err := getCurrentUser(ctx, hostname)
	if err != nil {
		return false, nil
	}

	return currentUser == user, nil
}

// getCurrentUser fetches the current authenticated user via API.
func getCurrentUser(ctx context.Context, hostname string) (string, error) {
	opts := api.ClientOptions{
		Host: hostname,
	}
	client, err := api.NewRESTClient(opts)
	if err != nil {
		return "", err
	}

	var response struct {
		Login string `json:"login"`
	}

	err = client.Get("user", &response)
	if err != nil {
		return "", err
	}

	return response.Login, nil
}

// GetCurrentUserFromSession gets the current user from the active gh session.
func GetCurrentUserFromSession(hostname string) (string, error) {
	opts := api.ClientOptions{
		Host: hostname,
	}
	client, err := api.NewRESTClient(opts)
	if err != nil {
		return "", err
	}

	var response struct {
		Login string `json:"login"`
	}

	err = client.Get("user", &response)
	if err != nil {
		return "", err
	}

	return response.Login, nil
}

// SwitchUser switches the gh CLI to use a specific user on a host.
func SwitchUser(hostname, user string) error {
	_, _, err := gh.Exec("auth", "switch", "--hostname", hostname, "--user", user)
	return err
}

// HasToken checks if there's an auth token for the given host.
func HasToken(hostname string) bool {
	_, _, err := gh.Exec("auth", "token", "--hostname", hostname)
	return err == nil
}

// GetAuthStatus returns raw auth status output for a hostname.
func GetAuthStatus(hostname string) (string, error) {
	stdout, stderr, err := gh.Exec("auth", "status", "--hostname", hostname)
	if err != nil {
		// gh auth status returns non-zero if not logged in, but still outputs info
		return stderr.String(), nil
	}
	return stdout.String(), nil
}

// IsUserLoggedIn checks if a specific user is logged in on a host.
func IsUserLoggedIn(hostname, user string) bool {
	stdout, _, err := gh.Exec("auth", "status", "--hostname", hostname)
	if err != nil {
		return false
	}

	output := stdout.String()
	expectedPattern := fmt.Sprintf("Logged in to %s account %s", hostname, user)
	return strings.Contains(output, expectedPattern)
}

// VerifyConnectivity tests that we can reach the GitHub API on the given host.
func VerifyConnectivity(hostname string) error {
	opts := api.ClientOptions{
		Host: hostname,
	}
	client, err := api.NewRESTClient(opts)
	if err != nil {
		return err
	}

	var response json.RawMessage
	return client.Get("user", &response)
}
