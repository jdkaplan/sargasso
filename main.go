package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"

	gsync "github.com/jdkaplan/sargasso/gsync"
)

type Node struct {
	node *maelstrom.Node

	nextID atomic.Uint64
	seen   *gsync.Set[int]

	neighbors []string
	pending   *gsync.Map[string, *gsync.Set[int]]
}

func NewNode() *Node {
	n := &Node{
		node:      maelstrom.NewNode(),
		nextID:    atomic.Uint64{},
		seen:      gsync.NewSet[int](),
		neighbors: nil,
		pending:   gsync.NewMap[string, *gsync.Set[int]](),
	}

	n.handle("echo", n.Echo)

	n.handle("generate", n.Generate)

	n.handle("broadcast", n.Broadcast)
	n.handle("read", n.Read)

	n.handle("gossip", n.Gossip)
	n.handle("gossip_ok", n.GossipOK)

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

//nolint:unused
func (n *Node) send(dest string, body any) error {
	return n.node.Send(dest, body)
}

func (n *Node) rpc(dest string, body any, h maelstrom.HandlerFunc) error {
	return n.node.RPC(dest, body, h)
}

func (n *Node) genID() string {
	count := n.nextID.Add(1)
	return fmt.Sprintf("%s-%d", n.node.ID(), count)
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
	msg := body.Message

	n.gossip(msg)

	n.seen.Add(msg)

	return n.reply(m, map[string]any{
		"type": "broadcast_ok",
	})
}

func (n *Node) gossip(msg int) {
	if n.seen.Has(msg) {
		return
	}

	body := map[string]any{
		"type":    "gossip",
		"message": msg,
	}

	for _, name := range n.neighbors {
		pending, _ := n.pending.LoadOrStore(name, gsync.NewSet[int]())
		pending.Add(msg)

		go func() {
			tick := time.NewTicker(5 * time.Millisecond)
			defer tick.Stop()

			for pending.Has(msg) {
				<-tick.C

				if err := n.rpc(name, body, n.GossipOK); err != nil {
					panic(err)
				}
			}
		}()
	}
}

func (n *Node) Gossip(m maelstrom.Message) error {
	var body struct {
		Message int `json:"message"`
	}
	if err := json.Unmarshal(m.Body, &body); err != nil {
		return fmt.Errorf("unpack message: %w", err)
	}
	msg := body.Message

	n.gossip(msg)

	n.seen.Add(msg)

	return n.reply(m, map[string]any{
		"type":    "gossip",
		"message": msg,
	})
}

func (n *Node) GossipOK(m maelstrom.Message) error {
	var body struct {
		Message int `json:"message"`
	}
	if err := json.Unmarshal(m.Body, &body); err != nil {
		return fmt.Errorf("unpack message: %w", err)
	}
	msg := body.Message

	if pending, ok := n.pending.Load(m.Src); ok {
		pending.Del(msg)
	}

	return nil
}

func (n *Node) Read(m maelstrom.Message) error {
	return n.reply(m, map[string]any{
		"type":     "read_ok",
		"messages": n.seen.Values(),
	})
}

func (n *Node) Topology(m maelstrom.Message) error {
	var body struct {
		Topology map[string][]string `json:"topology"`
	}
	if err := json.Unmarshal(m.Body, &body); err != nil {
		return fmt.Errorf("unpack message: %w", err)
	}

	n.neighbors = body.Topology[n.node.ID()]

	return n.reply(m, map[string]any{
		"type": "topology_ok",
	})
}

func main() {
	n := NewNode()
	if err := n.run(); err != nil {
		log.Fatal(err)
	}
}
