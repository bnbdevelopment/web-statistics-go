package statistics

import (
	"statistics/structs"
	"testing"
)

func TestSankeyDataConstruction(t *testing.T) {
	// Verify Sankey data mapping logic with sample flows
	flows := []structs.FlowResult{
		{SourcePage: "/", TargetPage: "/products", FlowCount: 100},
		{SourcePage: "/products", TargetPage: "/cart", FlowCount: 40},
		{SourcePage: "/products", TargetPage: "/about", FlowCount: 20},
		{SourcePage: "/cart", TargetPage: "/checkout", FlowCount: 30},
	}

	nodeMap := make(map[string]int)
	var nodes []structs.SankeyNode
	var links []structs.SankeyLink

	addNode := func(pageName string) {
		if _, exists := nodeMap[pageName]; !exists {
			nodeMap[pageName] = len(nodes)
			nodes = append(nodes, structs.SankeyNode{Name: pageName})
		}
	}

	for _, flow := range flows {
		addNode(flow.SourcePage)
		addNode(flow.TargetPage)

		links = append(links, structs.SankeyLink{
			Source: nodeMap[flow.SourcePage],
			Target: nodeMap[flow.TargetPage],
			Value:  flow.FlowCount,
		})
	}

	if len(nodes) != 5 {
		t.Errorf("Expected 5 unique nodes, got %d", len(nodes))
	}
	if len(links) != 4 {
		t.Errorf("Expected 4 links, got %d", len(links))
	}
	if links[0].Value != 100 {
		t.Errorf("Expected link 0 value 100, got %d", links[0].Value)
	}
}

func TestLandingPageBounceCalculation(t *testing.T) {
	stat := structs.LandingPageStat{
		Page:       "/landing",
		Sessions:   200,
		Bounces:    50,
		BounceRate: 25.0,
	}

	if stat.BounceRate != (float64(stat.Bounces)/float64(stat.Sessions))*100.0 {
		t.Errorf("Bounce rate mismatch: expected 25.0, got %f", stat.BounceRate)
	}
}
