package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TrieSuite struct {
	suite.Suite
}

func (s *TrieSuite) debugPrint(node *Trie, prefix string) {
	fmt.Printf(
		"node: (%p)\n\tkey: %v\tdata: %v\tchildren: %v\n",
		node, node.Key, node.Data, node.Children,
	)

	if len(node.Children) > 0 {
		if len(prefix) > 0 {
			fmt.Printf("--iter children of %v/%v --\n", prefix, node.Key)
			prefix = prefix + "/" + node.Key
		} else {
			fmt.Printf("--iter children of %v --\n", node.Key)
			prefix = node.Key
		}

		for _, v := range node.Children {
			s.debugPrint(v, prefix)
		}
		fmt.Printf("-- end of children %v --\n", prefix)
	}
}

func (s *TrieSuite) TestGinURLLeaf() {
	tree := &Trie{
		Children: make(map[string]*Trie),
		Meta: &TrieMeta{
			KeyFormatter: URLGinParamKeyFormatter,
			Tokenizer:    URLTokenizer,
			TokenJoiner:  URLTokenJoiner,
			DataInsertionFn: func(t *Trie, data interface{}) (err error) {
				t.Data = data
				return
			},
			DataDeletionFn: func(t *Trie, data interface{}) (err error) {
				if t.Data == data {
					t.Data = nil
				}
				return
			},
		},
	}

	tests := []struct {
		Path  string
		Value string
	}{
		// same prefix group
		{"/a/b/c", "abc"},
		{"/a/b/d", "abd"},
		{"/a/b", "ab"},
		{"/a/c", "ac"},
		// another prefix
		{"/b/c", "bc"},
		{"/c", "c"},
		// params
		{"/d/:ddd", "1 param"},
		{"/e/:e1/:e2/:e3", "3 params"},
		// dummy long test
		{"/i/j/m/n/t/x/y/z", "long one"},
	}

	var err error
	for i, t := range tests {
		err = tree.Insert(t.Path, t.Value)
		s.Require().Nil(err, "index: %v, error: %v", i, err)
	}

	s.debugPrint(tree, "")

	// test prefix
	collections := tree.Find("a")
	s.Require().Len(collections, 4)

	collections = tree.Find("a/b")
	s.Require().Len(collections, 3)

	collections = tree.Find(TrieRootIdentifier)
	s.Require().Len(collections, 9)

	// test Get
	var (
		node *Trie
		ok   bool
	)
	for i, t := range tests {
		node, ok = tree.Get(t.Path)
		s.Require().True(ok, "index: %v, can't find path: %v", i, t.Path)
		s.Require().NotNil(node)
		s.Require().NotNil(node.Data)
		s.Require().Equal(
			t.Value, node.Data.(string),
			"index: %v, node: %v, data: %v != %v",
			i, node, t.Value, node.Data.(string),
		)
	}

	// test Get with params
	node, ok = tree.Get("/d/:abc")
	s.Require().True(ok, "can't find path: /d/:abc")
	s.Require().NotNil(node)
	s.Require().NotNil(node.Data)
	s.Require().Equal("1 param", node.Data.(string))

	node, ok = tree.Get("e/:v1/:v2/:3")
	s.Require().True(ok, "can't find path: e/:v1/:v2/:v3")
	s.Require().NotNil(node)
	s.Require().NotNil(node.Data)
	s.Require().Equal("3 params", node.Data.(string))

	// make tree2 copy
	tree2 := tree.Copy()
	fmt.Printf("\n----tree2-----\n")
	s.debugPrint(tree2, "")
	fmt.Printf("\n----tree2-----\n")

	// test delete
	for i, t := range tests {
		err = tree.Delete(t.Path)
		s.Require().Nil(err, "index: %v, error: %v", i, err)

		fmt.Printf("\n\n--- after delete index: %v ---\n", i)
		s.debugPrint(tree, "")
	}
	s.Require().Len(tree.Children, 0)

	// test tree2 Get
	for i, t := range tests {
		node, ok = tree2.Get(t.Path)
		s.Require().True(ok, "tree2 index: %v, can't find path: %v", i, t.Path)
		s.Require().NotNil(node)
		s.Require().NotNil(node.Data)
		s.Require().Equal(
			t.Value, node.Data.(string),
			"tree2 index: %v, node: %v, data: %v != %v",
			i, node, t.Value, node.Data.(string),
		)
	}
}

func (s *TrieSuite) TestGinURLPath() {
	tree := &Trie{
		Children: make(map[string]*Trie),
		Meta: &TrieMeta{
			KeyFormatter: URLGinParamKeyFormatter,
			Tokenizer:    URLTokenizer,
			TokenJoiner:  URLTokenJoiner,
			DataInsertionFn: func(t *Trie, data interface{}) (err error) {
				t.Data = data
				return
			},
			DataDeletionFn: func(t *Trie, data interface{}) (err error) {
				if t.Data == data {
					t.Data = nil
				}
				return
			},

			DefautlDataMode: DataPath,
		},
	}

	tests := []struct {
		Path  string
		Value string
	}{
		{"/a/b/c", "abc"},
		{"/a/b/d", "abd"},
		{"/b/c", "bc"},
		{"/c", "c"},
		{"/d/:ccc", "1"},
	}

	var err error
	for i, t := range tests {
		err = tree.Insert(t.Path, t.Value)
		s.Require().Nil(err, "index: %v, error: %v", i, err)
	}

	s.debugPrint(tree, "")

	verifies := []struct {
		Path  string
		Value string
	}{
		{"/a/b/c", "abc"},
		{"/a/b/d", "abd"},
		{"/a/b", "abd"},
		{"/a", "abd"},
	}
	var (
		node *Trie
		ok   bool
	)
	for i, t := range verifies {
		node, ok = tree.Get(t.Path)
		s.Require().True(ok, "index: %v, can't find path: %v", i, t.Path)
		s.Require().NotNil(node, "index: %v, null", i)
		s.Require().NotNil(node.Data, "index: %v, null data", i)
		s.Require().Equal(
			t.Value, node.Data.(string),
			"index: %v, node: %v, data: %v != %v",
			i, node, t.Value, node.Data.(string),
		)
	}

	for i, t := range tests {
		err = tree.Delete(t.Path)
		s.Require().Nil(err, "index: %v, error: %v", i, err)

		fmt.Printf("--- after delete index: %v ---\n", i)
		s.debugPrint(tree, "")
	}

	s.Require().Len(tree.Children, 0)
}

func TestTrie(t *testing.T) {
	suite.Run(t, new(TrieSuite))
}
