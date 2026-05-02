package core

import (
	"sort"
	"strings"
	"text/template"
	"text/template/parse"

	"pilot/internal/core/domain"
)

// referencesAllServices is true when the template touches the root .Services
// (e.g. via {{ range .Services }}) rather than a specific service name.
type templateReferences struct {
	secrets               map[string]bool
	services              map[string]bool
	referencesAllServices bool
}

// ExtractTemplateVariables returns variable references grouped by type:
// "Secrets" keys are the full dotted path (e.g. "foo.bar"); "Services" keys
// are the leading service name only (e.g. "api" from .Services.api.path).
func ExtractTemplateVariables(text string) map[string][]string {
	references := extractReferences(text)
	result := make(map[string][]string)
	if len(references.secrets) > 0 {
		result["Secrets"] = sortedKeys(references.secrets)
	}
	if len(references.services) > 0 {
		result["Services"] = sortedKeys(references.services)
	}
	return result
}

// ExtractSecretKeys returns all unique secret keys referenced in a ConfigurationContext.
// It scans Scripts, Service.HelmArgs, and DockerImage.BuildArgs for template references.
func ExtractSecretKeys(ctx *domain.ConfigurationContext) []string {
	if ctx == nil {
		return nil
	}
	seen := make(map[string]bool)
	var keys []string

	collect := func(templates ...string) {
		for _, tmpl := range templates {
			for _, key := range ExtractTemplateVariables(tmpl)["Secrets"] {
				if !seen[key] {
					seen[key] = true
					keys = append(keys, key)
				}
			}
		}
	}

	for _, script := range ctx.Scripts {
		collect(script)
	}
	for _, svc := range ctx.Services {
		collect(svc.HelmArgs...)
		for _, img := range svc.DockerImages {
			collect(img.BuildArgs...)
		}
	}

	sort.Strings(keys)
	return keys
}

// ExtractServiceReferences returns service names referenced in the template,
// sorted alphabetically.
func ExtractServiceReferences(text string) []string {
	return ExtractTemplateVariables(text)["Services"]
}

// ReferencesAllServices reports whether a template touches the root .Services
// object (e.g. {{ range .Services }}) rather than a specific service; callers
// should over-fetch every configured service as a dependency when true.
func ReferencesAllServices(text string) bool {
	return extractReferences(text).referencesAllServices
}

// Parse failures (including templates using functions outside text/template
// builtins) yield zero-valued references; the renderer surfaces the error later.
func extractReferences(text string) *templateReferences {
	references := &templateReferences{
		secrets:  make(map[string]bool),
		services: make(map[string]bool),
	}
	if text == "" {
		return references
	}
	parsedTemplate, err := template.New("extract").Parse(text)
	if err != nil {
		return references
	}
	if parsedTemplate.Tree != nil {
		collectReferences(parsedTemplate.Root, references)
	}
	return references
}

func collectReferences(node parse.Node, references *templateReferences) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *parse.ListNode:
		// Typed-nil when an if/range/with has no else branch.
		if n == nil {
			return
		}
		for _, child := range n.Nodes {
			collectReferences(child, references)
		}
	case *parse.ActionNode:
		collectReferences(n.Pipe, references)
	case *parse.PipeNode:
		// Typed-nil when a TemplateNode has no pipeline argument.
		if n == nil {
			return
		}
		for _, command := range n.Cmds {
			collectReferences(command, references)
		}
	case *parse.CommandNode:
		if isIndexCall(n) {
			collectIndexReference(n, references)
			return
		}
		for _, arg := range n.Args {
			collectReferences(arg, references)
		}
	case *parse.IfNode:
		collectReferences(n.Pipe, references)
		collectReferences(n.List, references)
		collectReferences(n.ElseList, references)
	case *parse.RangeNode:
		collectReferences(n.Pipe, references)
		collectReferences(n.List, references)
		collectReferences(n.ElseList, references)
	case *parse.WithNode:
		collectReferences(n.Pipe, references)
		collectReferences(n.List, references)
		collectReferences(n.ElseList, references)
	case *parse.TemplateNode:
		collectReferences(n.Pipe, references)
	case *parse.FieldNode:
		collectFieldPath(n.Ident, references)
	case *parse.VariableNode:
		// $.Field accesses the root scope; skip the leading "$".
		if len(n.Ident) >= 2 && n.Ident[0] == "$" {
			collectFieldPath(n.Ident[1:], references)
		}
	case *parse.ChainNode:
		collectReferences(n.Node, references)
	}
}

func isIndexCall(command *parse.CommandNode) bool {
	if len(command.Args) < 3 {
		return false
	}
	identifier, ok := command.Args[0].(*parse.IdentifierNode)
	return ok && identifier.Ident == "index"
}

// collectIndexReference extracts the referenced secret or service name from an
// `(index .X "key" ...)` call. The collection argument is intentionally not
// re-walked so .Services or .Secrets is not misclassified as a bare-root
// reference; arguments past the key are still walked.
func collectIndexReference(command *parse.CommandNode, references *templateReferences) {
	collectionRoot := rootIdentifier(command.Args[1])
	keyNode, hasStringKey := command.Args[2].(*parse.StringNode)
	switch {
	case collectionRoot == "Secrets" && hasStringKey:
		references.secrets[keyNode.Text] = true
	case collectionRoot == "Services" && hasStringKey:
		references.services[keyNode.Text] = true
	case collectionRoot == "Services" && !hasStringKey:
		references.referencesAllServices = true
		collectReferences(command.Args[2], references)
	}
	for _, arg := range command.Args[3:] {
		collectReferences(arg, references)
	}
}

func rootIdentifier(node parse.Node) string {
	switch typed := node.(type) {
	case *parse.FieldNode:
		return typed.Ident[0]
	case *parse.VariableNode:
		// $.Field accesses the root scope; skip the leading "$".
		if len(typed.Ident) >= 2 && typed.Ident[0] == "$" {
			return typed.Ident[1]
		}
	}
	return ""
}

func collectFieldPath(path []string, references *templateReferences) {
	switch path[0] {
	case "Secrets":
		if len(path) >= 2 {
			references.secrets[strings.Join(path[1:], ".")] = true
		}
	case "Services":
		if len(path) >= 2 {
			references.services[path[1]] = true
		} else {
			references.referencesAllServices = true
		}
	}
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
