package client

import (
	"context"

	"github.com/uselagoon/deploy2lights/internal/schema"
)

func (c *Client) DeploymentsByBulkID(
	ctx context.Context, bulkID string, deployments *[]schema.Deployment) error {

	req, err := c.newRequest("_lgraphql/getDeploymentsByBulkID.graphql",
		map[string]interface{}{
			"bulkId": bulkID,
		})
	if err != nil {
		return err
	}

	return c.client.Run(ctx, req, &struct {
		Response *[]schema.Deployment `json:"deploymentsByBulkId"`
	}{
		Response: deployments,
	})
}
