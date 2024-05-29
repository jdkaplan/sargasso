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
}

func NewNode() *Node {
	n := &Node{
		node:   maelstrom.NewNode(),
		nextID: 0,
	}

	n.handle("echo", n.Echo)
	n.handle("generate", n.Generate)

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

func main() {
	n := NewNode()
	if err := n.run(); err != nil {
		log.Fatal(err)
	}
}
