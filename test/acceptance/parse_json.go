package acceptance

import (
	"encoding/json"
	"io"

	"github.com/hlebecz/crawler-cli/internal/model"
)

func CountNodes(r io.Reader) (nodes, maxDepth int, err error) {
	var roots []model.Node
	if err := json.NewDecoder(r).Decode(&roots); err != nil {
		return 0, 0, err
	}

	for _, n := range roots {
		c, d := walk(n, 1)
		nodes += c
		if d > maxDepth {
			maxDepth = d
		}
	}
	return nodes, maxDepth, nil
}

func walk(n model.Node, depth int) (count, maxDepth int) {
	count = 1
	maxDepth = depth
	for _, child := range n.Links {
		c, d := walk(*child, depth+1)
		count += c
		if d > maxDepth {
			maxDepth = d
		}
	}
	return count, maxDepth
}
