package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DrSkyle/cloudslash/internal/graph"
)

type SlackMessage struct {
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Type   string     `json:"type"`
	Text   *TextObj   `json:"text,omitempty"`
	Fields []*TextObj `json:"fields,omitempty"`
}

type TextObj struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func SendSlackReport(webhookURL string, g *graph.Graph) error {
	if webhookURL == "" {
		return nil
	}

	totalWaste := 0
	totalCost := 0.0
	highRisk := 0

	g.Mu.RLock()
	var topItems []*graph.Node
	for _, node := range g.Nodes {
		if node.IsWaste {
			totalWaste++
			totalCost += node.Cost
			if node.RiskScore >= 80 {
				highRisk++
			}
			topItems = append(topItems, node)
		}
	}
	g.Mu.RUnlock()

	// Title Section
	limit := 5
	if len(topItems) < limit {
		limit = len(topItems)
	}

	// Format Money
	costStr := fmt.Sprintf("$%.2f", totalCost)

	msg := SlackMessage{
		Blocks: []Block{
			{
				Type: "header",
				Text: &TextObj{Type: "plain_text", Text: "CloudSlash Audit Report 🗡️☁️"},
			},
			{
				Type: "section",
				Fields: []*TextObj{
					{Type: "mrkdwn", Text: fmt.Sprintf("*Total Waste Cost:*\n%s/mo", costStr)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Resources Flagged:*\n%d", totalWaste)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*High Risk Items:*\n%d", highRisk)},
					{Type: "mrkdwn", Text: fmt.Sprintf("*Scan Time:*\n%s", time.Now().Format("2006-01-02 15:04"))},
				},
			},
			{
				Type: "divider",
			},
			{
				Type: "section",
				Text: &TextObj{Type: "mrkdwn", Text: "*Top Waste Items (Preview):*"},
			},
		},
	}

	// Add top items
	for i := 0; i < limit; i++ {
		node := topItems[i]
		idParts := strings.Split(node.ID, "/")
		shortID := node.ID
		if len(idParts) > 1 {
			shortID = idParts[len(idParts)-1]
		}

		emoji := "⚠️"
		if node.RiskScore >= 80 {
			emoji = "🚨"
		}

		reason, _ := node.Properties["Reason"].(string)

		itemText := fmt.Sprintf("%s *%s* (Risk: %d)\n> %s", emoji, shortID, node.RiskScore, reason)
		if node.Cost > 0 {
			itemText += fmt.Sprintf("\n> *Cost: $%.2f/mo*", node.Cost)
		}

		msg.Blocks = append(msg.Blocks, Block{
			Type: "section",
			Text: &TextObj{Type: "mrkdwn", Text: itemText},
		})
	}

	// Warning footer if trial
	/*
		msg.Blocks = append(msg.Blocks, Block{
			Type: "context",
			Elements: ...
		})
	*/

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("slack api returned %d", resp.StatusCode)
	}

	return nil
}
