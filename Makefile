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

sargasso: $(SOURCE_FILES) ## Build the node binary
	go build -o sargasso

.PHONY: serve
serve:
	maelstrom serve

.PHONY: echo
echo: sargasso ## Challenge #1
	maelstrom test -w echo --bin sargasso --node-count 1 --time-limit 10
