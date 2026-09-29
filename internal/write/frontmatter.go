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
	doc, body, nl, ok, err := frontOf(src)
	if err != nil {
		return nil, err
	}
	if !ok {
		block, err := encode(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}, key, value)
		if err != nil {
			return nil, err
		}
		return put(nl, block, body), nil
	}
	block, err := encode(doc, key, value)
	if err != nil {
		return nil, err
	}
	return put(nl, block, body), nil
}

// RemoveField takes key out of the frontmatter. The other fields keep their
// order and the body is copied byte for byte. A key that is not there leaves
// the file exactly as it came in.
func RemoveField(src []byte, key string) ([]byte, error) {
	doc, body, nl, ok, err := frontOf(src)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("file has no frontmatter")
	}
	m := doc.Content[0]
	if m.Kind != yaml.MappingNode {
		return nil, errors.New("frontmatter is not a key: value list")
	}
	kept := make([]*yaml.Node, 0, len(m.Content))
	found := false
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			found = true
			continue
		}
		kept = append(kept, m.Content[i], m.Content[i+1])
	}
	if !found {
		return src, nil
	}
	m.Content = kept
	block, err := render(doc)
	if err != nil {
		return nil, err
	}
	return put(nl, block, body), nil
}

// hasField says whether the frontmatter already has key, so a date that is
// written once is not written again.
func hasField(src []byte, key string) bool {
	doc, _, _, ok, err := frontOf(src)
	if err != nil || !ok {
		return false
	}
	m := doc.Content[0]
	if m.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return true
		}
	}
	return false
}

// frontOf reads the frontmatter of src as a yaml document, and gives the
// body and the line ending the file uses. ok is false when the file has no
// frontmatter, so a caller can open one.
func frontOf(src []byte) (doc *yaml.Node, body, nl string, ok bool, err error) {
	text := string(src)
	nl = "\n"
	// The reader already accepts CRLF, so an edit must too: keep the
	// file's own line ending or a CRLF file gains a second block.
	if strings.HasPrefix(text, "---\r\n") {
		nl = "\r\n"
	} else if !strings.HasPrefix(text, "---\n") {
		return nil, text, nl, false, nil
	}
	// The closing fence is the first line after the opening one that is
	// "---" once a trailing \r is dropped, whatever mix of \n and \r\n
	// the file uses. A "---" line in the body is just body text.
	lines := strings.SplitAfter(text, "\n")
	close := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSuffix(strings.TrimSuffix(lines[i], "\n"), "\r") == "---" {
			close = i
			break
		}
	}
	if close < 0 {
		return nil, "", "", false, errors.New("frontmatter has no closing ---")
	}
	front := strings.TrimSuffix(strings.TrimSuffix(strings.Join(lines[1:close], ""), "\n"), "\r")
	body = strings.Join(lines[close+1:], "")

	var parsed yaml.Node
	if err := yaml.Unmarshal([]byte(front), &parsed); err != nil {
		return nil, "", "", false, err
	}
	if parsed.Kind == 0 {
		parsed = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}
	}
	return &parsed, body, nl, true, nil
}

// put puts a rewritten frontmatter block back between the fences, with the
// body behind it untouched.
func put(nl, block, body string) []byte {
	if nl == "\r\n" {
		block = strings.ReplaceAll(block, "\n", "\r\n")
	}
	return []byte("---" + nl + block + "---" + nl + body)
}

func render(doc *yaml.Node) (string, error) {
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
	return render(doc)
}
