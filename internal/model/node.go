package model

import "sync"

type Node struct {
	Resource string `json:"resource"`
	Title    string `json:"title"`
	Links    []Node `json:"links"`
	Depth    int
	mx       *sync.Mutex
}

func (n *Node) AddLink(link Node) {
	n.mx.Lock()
	defer n.mx.Unlock()
	n.Links = append(n.Links, link)
}

func NewNode(r, t string, d int) Node {
	return Node{Resource: r, Title: t, Links: []Node{}, Depth: d, mx: &sync.Mutex{}}
}
