package main

import (
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Node struct {
	node *maelstrom.Node

	nextID int
	seen   map[int]bool

	neighbors map[string]bool
}

func NewNode() *Node {
	n := &Node{
		node:      maelstrom.NewNode(),
		nextID:    0,
		seen:      make(map[int]bool),
		neighbors: make(map[string]bool),
	}

	n.handle("echo", n.Echo)

	n.handle("generate", n.Generate)

	n.handle("broadcast", n.Broadcast)
	n.handle("read", n.Read)

	n.handle("topology", n.Topology)

	return n
}

func (n *Node) handle(name string, h maelstrom.HandlerFunc) {
	n.node.Handle(name, h)
}

func (n *Node) run() error {
	return n.node.Run()
}

func (n *Node) reply(m maelstrom.Message, body any) error {
	return n.node.Reply(m, body)
}

func (n *Node) genID() string {
	n.nextID++
	return fmt.Sprintf("%s-%d", n.node.ID(), n.nextID)
}

func (n *Node) Echo(m maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(m.Body, &body); err != nil {
		return fmt.Errorf("unpack message: %w", err)
	}

	body["type"] = "echo_ok"

	return n.reply(m, body)
}

func (n *Node) Generate(m maelstrom.Message) error {
	return n.reply(m, map[string]any{
		"type": "generate_ok",
		"id":   n.genID(),
	})
}

func (n *Node) Broadcast(m maelstrom.Message) error {
	var body struct {
		Message int `json:"message"`
	}
	if err := json.Unmarshal(m.Body, &body); err != nil {
		return fmt.Errorf("unpack message: %w", err)
	}

	n.seen[body.Message] = true

	return n.reply(m, map[string]any{
		"type": "broadcast_ok",
	})
}

func (n *Node) Read(m maelstrom.Message) error {
	return n.reply(m, map[string]any{
		"type":     "read_ok",
		"messages": keys(n.seen),
	})
}

func (n *Node) Topology(m maelstrom.Message) error {
	var body struct {
		Topology map[string][]string `json:"topology"`
	}
	if err := json.Unmarshal(m.Body, &body); err != nil {
		return fmt.Errorf("unpack message: %w", err)
	}

	for _, name := range body.Topology[n.node.ID()] {
		n.neighbors[name] = true
	}

	return n.reply(m, map[string]any{
		"type": "topology_ok",
	})
}

func keys[K comparable, V any](m map[K]V) []K {
	ks := make([]K, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func main() {
	n := NewNode()
	if err := n.run(); err != nil {
		log.Fatal(err)
	}
}
