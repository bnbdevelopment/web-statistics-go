package analysis

import (
	"testing"
	"time"
)

func TestClassifySession(t *testing.T) {
	now := time.Now()

	// 1. Frustrated: high loop score (repeated visits to same pages)
	frustrated := SessionFeature{
		SessionID:       "s-frustrated",
		Duration:        200,
		PageCount:       10,
		UniquePageCount: 3,
		LoopScore:       3.33,
		FirstEvent:      now,
		LastEvent:       now.Add(200 * time.Second),
		Pages:           []string{"/", "/search", "/", "/search", "/", "/search"},
	}
	if got := classifySession(frustrated); got != ArchetypeFrustrated {
		t.Errorf("Expected %s, got %s", ArchetypeFrustrated, got)
	}

	// 2. Targeted: fast, low page count (< 60s, <= 3 pages)
	targeted := SessionFeature{
		SessionID:       "s-targeted",
		Duration:        45,
		PageCount:       2,
		UniquePageCount: 2,
		LoopScore:       1.0,
		FirstEvent:      now,
		LastEvent:       now.Add(45 * time.Second),
		Pages:           []string{"/", "/contact"},
	}
	if got := classifySession(targeted); got != ArchetypeTargeted {
		t.Errorf("Expected %s, got %s", ArchetypeTargeted, got)
	}

	// 3. Engaged: long session (> 600s), >= 4 pages, low looping
	engaged := SessionFeature{
		SessionID:       "s-engaged",
		Duration:        800,
		PageCount:       7,
		UniquePageCount: 6,
		LoopScore:       1.16,
		FirstEvent:      now,
		LastEvent:       now.Add(800 * time.Second),
		Pages:           []string{"/", "/blog", "/post-1", "/post-2", "/about", "/pricing"},
	}
	if got := classifySession(engaged); got != ArchetypeEngaged {
		t.Errorf("Expected %s, got %s", ArchetypeEngaged, got)
	}

	// 4. Default / General visitor
	general := SessionFeature{
		SessionID:       "s-default",
		Duration:        180,
		PageCount:       3,
		UniquePageCount: 3,
		LoopScore:       1.0,
		FirstEvent:      now,
		LastEvent:       now.Add(180 * time.Second),
		Pages:           []string{"/", "/products", "/checkout"},
	}
	if got := classifySession(general); got != ArchetypeDefault {
		t.Errorf("Expected %s, got %s", ArchetypeDefault, got)
	}
}
