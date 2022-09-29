// Package lagoon implements high-level functions for interacting with the
// Lagoon API.
package lagoon

import (
	"context"

	"github.com/uselagoon/deploy2lights/internal/schema"
)

// Deploy interface contains methods for deploying branches and environments in lagoon.
type Deploy interface {
	DeployEnvironmentLatest(ctx context.Context, deploy *schema.DeployEnvironmentLatestInput, result *schema.DeployEnvironmentLatest) error
	DeploymentsByBulkID(ctx context.Context, bulkID string, deployments *[]schema.Deployment) error
}

// DeployLatest deploys the latest environment.
func DeployLatest(ctx context.Context, deploy *schema.DeployEnvironmentLatestInput, m Deploy) (*schema.DeployEnvironmentLatest, error) {
	result := schema.DeployEnvironmentLatest{}
	return &result, m.DeployEnvironmentLatest(ctx, deploy, &result)
}

// GetDeploymentsByBulkID gets info of projects in lagoon that have matching metadata.
func GetDeploymentsByBulkID(ctx context.Context, bulkID string, d Deploy) (*[]schema.Deployment, error) {
	deployments := []schema.Deployment{}
	return &deployments, d.DeploymentsByBulkID(ctx, bulkID, &deployments)
}
