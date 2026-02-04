package structs

// ArchetypeCharacteristic defines a single behavioral trait of an archetype.
type ArchetypeCharacteristic struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Archetype represents a user persona identified by the analysis.
type Archetype struct {
	Name             string                    `json:"name"`
	Percentage       float64                   `json:"percentage"`
	Characteristics  []ArchetypeCharacteristic `json:"characteristics"`
	ExampleSessionID string                    `json:"example_session_id"` // To potentially link to a session explorer later
}
