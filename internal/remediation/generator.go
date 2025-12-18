package remediation

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DrSkyle/cloudslash/internal/graph"
)

// Generator handles creates remediation scripts.
type Generator struct {
	Graph *graph.Graph
}

// NewGenerator creates a new remediation generator.
func NewGenerator(g *graph.Graph) *Generator {
	return &Generator{Graph: g}
}

// GenerateSafeDeleteScript creates a shell script for safe cleanup.
func (g *Generator) GenerateSafeDeleteScript(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	g.Graph.Mu.RLock()
	defer g.Graph.Mu.RUnlock()

	fmt.Fprintf(f, "#!/bin/bash\n")
	fmt.Fprintf(f, "# CloudSlash Safe Remediation Script\n")
	fmt.Fprintf(f, "# Generated: %s\n\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "set -e\n\n") // Exit on error

	wasteCount := 0

	for _, node := range g.Graph.Nodes {
		if !node.IsWaste {
			continue
		}

		// Resource ID extraction helpers
		// Typically ID is ARN or ID.
		// For commands we usually need the ID.
		resourceID := extractResourceID(node.ID)

		switch node.Type {
		case "AWS::EC2::Volume":
			fmt.Fprintf(f, "echo \"Processing Volume: %s\"\n", resourceID)
			// Safety Snapshot
			desc := fmt.Sprintf("CloudSlash-Archive-%s", resourceID)
			fmt.Fprintf(f, "aws ec2 create-snapshot --volume-id %s --description \"%s\" --tag-specifications 'ResourceType=snapshot,Tags=[{Key=CloudSlash,Value=Archive}]'\n", resourceID, desc)
			// Delete
			fmt.Fprintf(f, "aws ec2 delete-volume --volume-id %s\n\n", resourceID)
			wasteCount++

		case "AWS::RDS::DBInstance":
			fmt.Fprintf(f, "echo \"Processing RDS: %s\"\n", resourceID)
			// Safety Snapshot
			snapID := fmt.Sprintf("cloudslash-snap-%s-%d", resourceID, time.Now().Unix())
			fmt.Fprintf(f, "aws rds create-db-snapshot --db-instance-identifier %s --db-snapshot-identifier %s\n", resourceID, snapID)
			// Delete (Skip final snapshot since we just took one, or force skip)
			fmt.Fprintf(f, "aws rds delete-db-instance --db-instance-identifier %s --skip-final-snapshot\n\n", resourceID)
			wasteCount++

		case "AWS::EC2::NatGateway":
			fmt.Fprintf(f, "echo \"Processing NAT Gateway: %s\"\n", resourceID)
			// NAT Gateways don't have snapshots, but maybe log it?
			// Just delete.
			fmt.Fprintf(f, "aws ec2 delete-nat-gateway --nat-gateway-id %s\n\n", resourceID)
			wasteCount++

		case "AWS::EC2::EIP":
			fmt.Fprintf(f, "echo \"Processing EIP: %s\"\n", resourceID)
			// Release
			fmt.Fprintf(f, "aws ec2 release-address --allocation-id %s\n\n", resourceID)
			wasteCount++
		}
	}

	if wasteCount == 0 {
		fmt.Fprintf(f, "echo \"No waste found to remediate.\"\n")
	} else {
		fmt.Fprintf(f, "echo \"Safe Remediation Complete. %d resources processed.\"\n", wasteCount)
	}

	return nil
}

func extractResourceID(id string) string {
	// Simple ARN parser: arn:aws:service:region:account:type/id
	parts := strings.Split(id, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	// Some ARNs use : separator for ID (e.g. SNS)
	parts = strings.Split(id, ":")
	if len(parts) > 6 {
		return parts[6] // arn:aws:sns:us-east-1:123456:topic-name (index 5 or 6 depending on split)
	}
	// Fallback: return as is assuming it's already an ID if not ARN
	return id
}
