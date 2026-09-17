package config

import (
	_ "embed"
	"encoding/json"
	"log"
	"os"
	"sync"
)

//go:embed funnel_steps.json
var embeddedFunnelSteps []byte

// FunnelStep represents a single configurable funnel milestone.
type FunnelStep struct {
	Label    string   `json:"label"`
	Keywords []string `json:"keywords"`
}

var (
	funnelSteps map[string][]FunnelStep
	once        sync.Once
)

func loadFunnelConfig() {
	data := embeddedFunnelSteps

	if customPath := os.Getenv("FUNNEL_CONFIG_PATH"); customPath != "" {
		if b, err := os.ReadFile(customPath); err == nil {
			data = b
		} else {
			log.Printf("funnel config: unable to read %s, falling back to embedded defaults: %v", customPath, err)
		}
	} else if _, err := os.Stat("config/funnel_steps.json"); err == nil {
		if b, err := os.ReadFile("config/funnel_steps.json"); err == nil {
			data = b
		} else {
			log.Printf("funnel config: unable to read config/funnel_steps.json, falling back to embedded defaults: %v", err)
		}
	}

	if err := json.Unmarshal(data, &funnelSteps); err != nil {
		log.Printf("funnel config: failed to parse config, using baked-in defaults: %v", err)
		funnelSteps = make(map[string][]FunnelStep)
	}

	if len(funnelSteps["default"]) == 0 {
		funnelSteps["default"] = []FunnelStep{
			{Label: "Kampány landing", Keywords: []string{"landing", "kampany", "promo"}},
			{Label: "Regisztrációs űrlap", Keywords: []string{"register", "signup", "regisztracio"}},
			{Label: "Próbaanyag indítása", Keywords: []string{"trial", "demo", "ingyenes"}},
			{Label: "Fizetés / beiratkozás", Keywords: []string{"checkout", "payment", "befizetes"}},
		}
	}
}

// GetFunnelSteps returns the configured steps for a site or the default definition.
func GetFunnelSteps(site string) []FunnelStep {
	once.Do(loadFunnelConfig)
	if steps, ok := funnelSteps[site]; ok && len(steps) > 0 {
		return steps
	}
	return funnelSteps["default"]
}
