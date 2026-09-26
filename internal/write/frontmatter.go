// Package write changes planning files and commits the result.
package write

import (
	"bytes"
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

// SetField sets key to value in the frontmatter of src. Other fields keep
// their order and the body is copied byte for byte.
func SetField(src []byte, key, value string) ([]byte, error) {
	text := string(src)
	if !strings.HasPrefix(text, "---\n") {
		block, err := encode(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}, key, value)
		if err != nil {
			return nil, err
		}
		return []byte("---\n" + block + "---\n" + text), nil
	}
	if strings.HasSuffix(text, "\n---") {
		text += "\n"
	}
	rest := text[3:]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return nil, errors.New("frontmatter has no closing ---")
	}
	front, body := rest[1:end+1], rest[end+5:]

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(front), &doc); err != nil {
		return nil, err
	}
	if doc.Kind == 0 {
		doc = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}
	}
	block, err := encode(&doc, key, value)
	if err != nil {
		return nil, err
	}
	return []byte("---\n" + block + "---\n" + body), nil
}

func encode(doc *yaml.Node, key, value string) (string, error) {
	m := doc.Content[0]
	if m.Kind != yaml.MappingNode {
		return "", errors.New("frontmatter is not a key: value list")
	}
	val := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
	found := false
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = val
			found = true
		}
	}
	if !found {
		m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, val)
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
