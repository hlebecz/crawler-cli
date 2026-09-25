package model

type Node struct {
	Resource string  `json:"resource"`
	Title    string  `json:"title"`
	Links    []*Node `json:"links"`
	Depth    uint    `json:"-"`
}

func NewNode(r, t string, d uint) Node {
	return Node{Resource: r, Title: t, Links: []*Node{}, Depth: d}
}
