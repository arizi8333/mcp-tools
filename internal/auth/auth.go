package auth

import (
	"context"
	"errors"
	"os"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

// Authenticator handles request authentication and role-based permissions.
type Authenticator struct {
	expectedKey string
	roleMap     map[string]Role
	permissions map[Role][]string
}

func NewAuthenticator() *Authenticator {
	// Simple mapping for demo purposes.
	// In production, this would be backed by a DB or Identity Provider.
	return &Authenticator{
		expectedKey: os.Getenv("MCP_API_KEY"),
		roleMap: map[string]Role{
			"admin-token":    RoleAdmin,
			"operator-token": RoleOperator,
			"viewer-token":   RoleViewer,
		},
		permissions: map[Role][]string{
			RoleAdmin:    {"*"}, // Wildcard for all tools
			RoleOperator: {"platform.health", "logs.analyze", "metrics.query", "db.inspect", "incident.analyze"},
			RoleViewer:   {"platform.health", "metrics.query"},
		},
	}
}

// Authenticate verifies the provided API key and returns the associated Role.
func (a *Authenticator) Authenticate(ctx context.Context, apiKey string) (Role, error) {
	if a.expectedKey == "" {
		return RoleAdmin, nil // Default to Admin if global auth is disabled
	}

	role, exists := a.roleMap[apiKey]
	if !exists {
		// Fallback to absolute match for MCP_API_KEY as Admin
		if apiKey == a.expectedKey {
			return RoleAdmin, nil
		}
		return "", errors.New("unauthorized: invalid API key")
	}
	return role, nil
}

// CheckPermission verifies if a role is authorized to execute a specific tool.
func (a *Authenticator) CheckPermission(role Role, toolName string) bool {
	perms, ok := a.permissions[role]
	if !ok {
		return false
	}

	for _, p := range perms {
		if p == "*" || p == toolName {
			return true
		}
	}
	return false
}
