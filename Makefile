.DEFAULT_GOAL := help

.PHONY: help
help: ## List targets in this Makefile
	@awk '\
		BEGIN { FS = ":$$|:[^#]+|:.*?## "; OFS="\t" }; \
		/^[0-9a-zA-Z_-]+?:/ { print $$1, $$2 } \
	' $(MAKEFILE_LIST) \
		| sort --dictionary-order \
		| column --separator $$'\t' --table --table-wrap 2 --output-separator '    '

SOURCE_FILES = go.mod go.sum $(wildcard *.go)
MAELSTROM_TEST = maelstrom test --bin sargasso

sargasso: $(SOURCE_FILES) ## Build the node binary
	go build -o sargasso

.PHONY: serve
serve:
	maelstrom serve

.PHONY: echo
echo: sargasso ## Challenge #1
	$(MAELSTROM_TEST) -w echo --node-count 1 --time-limit 10

.PHONY: unique-ids
unique-ids: sargasso ## Challenge #2
	$(MAELSTROM_TEST) -w unique-ids --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
