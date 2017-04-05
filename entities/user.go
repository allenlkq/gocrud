package entities

type User struct {
Id       string `json:"id,omitempty"`
Device   []Device `json:"device,omitempty"`
Role     []Role `json:"role,omitempty"`
Tool     []Tool `json:"tool,omitempty"`
Name     string `json:"name,omitempty"`
Password string `json:"password,omitempty"`
}