package model

import "sync"

type Node struct {
	Resource string  `json:"resource"`
	Title    string  `json:"title"`
	Links    []*Node `json:"links"`
	Depth    uint    `json:"-"`
	mx       *sync.RWMutex
}

func (n *Node) AddLink(link *Node) {
	n.mx.RLock()
	defer n.mx.RUnlock()
	n.Links = append(n.Links, link)
}

func NewNode(r, t string, d uint) Node {
	return Node{Resource: r, Title: t, Links: []*Node{}, Depth: d, mx: &sync.RWMutex{}}
}
