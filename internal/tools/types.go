package tools

// ServiceInfo represents information about a protobuf service
type ServiceInfo struct {
	Name        string       `json:"name"`
	FullName    string       `json:"full_name"`
	Methods     []MethodInfo `json:"methods"`
	File        string       `json:"file"`
	Package     string       `json:"package"`
	Description string       `json:"description"`
	Options     []OptionInfo `json:"options,omitempty"`
}

// MethodInfo represents information about a protobuf service method
type MethodInfo struct {
	Name            string       `json:"name"`
	InputType       string       `json:"input_type"`
	OutputType      string       `json:"output_type"`
	ClientStreaming bool         `json:"client_streaming"`
	ServerStreaming bool         `json:"server_streaming"`
	Description     string       `json:"description"`
	Options         []OptionInfo `json:"options,omitempty"`
}

// OptionInfo represents information about protobuf options
type OptionInfo struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// Helper functions for creating protobuf values
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func boolPtr(b bool) *bool {
	return &b
}
