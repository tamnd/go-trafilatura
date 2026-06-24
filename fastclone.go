// This file is part of go-trafilatura, Go package for extracting readable
// content, comments and metadata from a web page. Source available in
// <https://github.com/tamnd/go-trafilatura>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0

package trafilatura

import "golang.org/x/net/html"

// cloneNode is a drop-in replacement for go-shiori/dom.Clone. The shiori
// version allocates one *html.Node and one attribute slice per source node,
// which makes deep clones of large documents one of the heaviest allocators in
// the extractor. cloneNode instead counts the subtree once and allocates every
// node in a single backing slab, turning N small heap objects into one. The
// produced tree is identical to dom.Clone's: same Type, DataAtom, Data and
// attributes, Namespace left zero (dom.Clone does not copy it either), and the
// same parent/sibling wiring.
//
// Slab nodes cannot be freed individually, so a node that is later moved into
// the result keeps its slab alive until the result itself is collected. The
// clones here are short-lived backups (fallback comparison, baseline and
// wild-text rescue), so that costs nothing in practice.
func cloneNode(src *html.Node, deep bool) *html.Node {
	if src == nil {
		return nil
	}

	if !deep {
		dst := &html.Node{
			Type:     src.Type,
			DataAtom: src.DataAtom,
			Data:     src.Data,
		}
		if len(src.Attr) > 0 {
			dst.Attr = append([]html.Attribute(nil), src.Attr...)
		}
		return dst
	}

	slab := make([]html.Node, countNodes(src))
	idx := 0

	var build func(s *html.Node) *html.Node
	build = func(s *html.Node) *html.Node {
		n := &slab[idx]
		idx++
		n.Type = s.Type
		n.DataAtom = s.DataAtom
		n.Data = s.Data
		if len(s.Attr) > 0 {
			n.Attr = append([]html.Attribute(nil), s.Attr...)
		}
		for c := s.FirstChild; c != nil; c = c.NextSibling {
			child := build(c)
			child.Parent = n
			if n.FirstChild == nil {
				n.FirstChild = child
			} else {
				n.LastChild.NextSibling = child
				child.PrevSibling = n.LastChild
			}
			n.LastChild = child
		}
		return n
	}

	return build(src)
}

// countNodes returns the number of nodes in the subtree rooted at n, so the
// whole clone can be allocated in one slab.
func countNodes(n *html.Node) int {
	count := 1
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		count += countNodes(c)
	}
	return count
}

// hasElementByTag reports whether the subtree rooted at node has a descendant
// element with the given tag name. It matches the traversal of
// dom.GetElementsByTagName(node, tag) but stops at the first hit and allocates
// nothing, so it is a drop-in for callers that only test existence rather than
// len(...) > 0 or len(...) == 0 on the full slice.
func hasElementByTag(node *html.Node, tag string) bool {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tag {
			return true
		}
		if hasElementByTag(c, tag) {
			return true
		}
	}
	return false
}
