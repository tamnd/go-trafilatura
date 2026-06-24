package trafilatura

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-shiori/dom"
)

// TestCloneNodeMatchesShiori verifies cloneNode produces a tree that renders
// byte-identically to go-shiori/dom.Clone across every mock HTML file, for both
// deep and shallow clones. This is the contract the slab clone must keep so the
// extractor output never changes.
func TestCloneNodeMatchesShiori(t *testing.T) {
	files, err := filepath.Glob("test-files/mock/*.html")
	if err != nil || len(files) == 0 {
		t.Skipf("no mock files: %v", err)
	}

	var deepChecked, shallowChecked int
	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("open %s: %v", path, err)
		}
		root, err := dom.Parse(f)
		f.Close()
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}

		// Deep clone of the whole document.
		want := dom.OuterHTML(dom.Clone(root, true))
		got := dom.OuterHTML(cloneNode(root, true))
		if want != got {
			t.Fatalf("deep clone mismatch for %s", filepath.Base(path))
		}
		deepChecked++

		// Shallow clone of every element node (attrs preserved, no children).
		for _, el := range dom.GetElementsByTagName(root, "*") {
			ws := dom.OuterHTML(dom.Clone(el, false))
			gs := dom.OuterHTML(cloneNode(el, false))
			if ws != gs {
				t.Fatalf("shallow clone mismatch in %s for <%s>", filepath.Base(path), dom.TagName(el))
			}
			shallowChecked++
		}
	}
	t.Logf("clone equivalence verified: %d deep, %d shallow, across %d files",
		deepChecked, shallowChecked, len(files))
}
