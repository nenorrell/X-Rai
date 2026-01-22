package schema

// Trigger represents a database trigger.
type Trigger struct {
	TriggerName       string   `json:"trigger_name"`
	Timing            string   `json:"timing"`
	Events            []string `json:"events"`
	FunctionOrProcedure string   `json:"function_or_procedure,omitempty"`
	Definition        *string  `json:"definition,omitempty"`
}

// TriggersOutput represents the table.triggers.json output file.
type TriggersOutput struct {
	Triggers []Trigger `json:"triggers"`
}
