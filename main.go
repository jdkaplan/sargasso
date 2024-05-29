package main

import (
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	n.Handle("echo", func(m maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(m.Body, &body); err != nil {
			return fmt.Errorf("unpack message: %w", err)
		}

		body["type"] = "echo_ok"

		return n.Reply(m, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
