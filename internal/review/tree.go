package review

// FlowTree represents a complete review flow with multiple root trees.
type FlowTree struct {
	Mode        string      `json:"mode"`                  // "diff" or "codebase"
	Title       string      `json:"title"`                 // review title
	Description string      `json:"description,omitempty"` // AI-generated overview
	Roots       []*FlowNode `json:"roots"`                 // multiple root nodes
}

// FlowNode represents a single node in the flow tree.
type FlowNode struct {
	ID          string      `json:"id"`
	Label       string      `json:"label"`                 // display name (e.g. "LoadProject")
	Description string      `json:"description,omitempty"` // AI-generated commentary
	Package     string      `json:"package"`               // package path
	File        string      `json:"file"`                  // file path
	LineStart   int         `json:"lineStart"`             // code start line
	LineEnd     int         `json:"lineEnd"`               // code end line
	Code        string      `json:"code"`                  // source code (current version)
	Diff        string      `json:"diff,omitempty"`        // raw diff content (diff mode only)
	NodeType    string      `json:"nodeType"`              // "function" / "method" / "rpc" / "mq" / "file"
	IsNew       bool        `json:"isNew"`                 // diff mode: new file/function
	Children    []*FlowNode `json:"children,omitempty"`
}
