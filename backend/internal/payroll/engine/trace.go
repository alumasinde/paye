package engine

// Trace is intentionally independent from HTTP and persistence. It can be
// enabled by verification and future admin/support tooling without changing
// calculation semantics.
type Trace struct { Steps []TraceStep `json:"steps"` }
type TraceStep struct {
    Name string `json:"name"`
    Input string `json:"input,omitempty"`
    Output string `json:"output,omitempty"`
    Detail map[string]string `json:"detail,omitempty"`
}
