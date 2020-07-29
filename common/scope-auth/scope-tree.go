package scopeauth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cobinhood/mochi/common/utils"
	"github.com/cobinhood/mochi/types"
)

// alias
type aliasMap = map[string][]types.Scope

type scopeData struct {
	Method string
	Path   string
	Scopes []types.Scope
}

func (s *scopeData) toMap() (m aliasMap) {
	m = make(aliasMap)
	m[s.Method] = s.Scopes
	return
}

func insertScope(node *utils.Trie, data interface{}) (err error) {
	if data == nil {
		err = errors.New("nil data")
		return
	}

	var ok bool

	sData, ok := data.(*scopeData)
	if !ok {
		err = fmt.Errorf("not scopeData struct: %v", data)
		return
	}

	// node create
	if node.Data == nil {
		node.Data = sData.toMap()
		return
	}

	var m aliasMap

	// node update
	m, ok = node.Data.(aliasMap)
	if !ok {
		err = fmt.Errorf("not expected data format on existing node: %v", data)
		return
	}

	_, ok = m[sData.Method]
	if ok {
		err = fmt.Errorf("scopes already existing: %v", data)
		return
	}

	m[sData.Method] = sData.Scopes
	node.Data = m
	return
}

// ScopeTree is a tree of ScopeNode, use `NewTree()` to create a new instance.
type ScopeTree struct {
	*utils.Trie

	name string
}

// NewTree returns a new ScopeTree instance with `name`.
func NewTree(name string) *ScopeTree {
	return &ScopeTree{
		Trie: &utils.Trie{
			Children: make(map[string]*utils.Trie),
			Meta: &utils.TrieMeta{
				KeyFormatter:    utils.URLGinParamKeyFormatter,
				Tokenizer:       utils.URLTokenizer,
				TokenJoiner:     utils.URLTokenJoiner,
				DataInsertionFn: insertScope,
				ArbitraryKey:    utils.URLGinArbitraryRepl,
			},
		},
		name: name,
	}
}

// InsertWithPath inserts a leaf (a node with `scopeMap`) to the tree, calling func
// `Insert(elements []string, method string, scopes []types.Scope) error` under
// the hood.
func (t *ScopeTree) InsertWithPath(path, method string, scopes []types.Scope) error {
	return t.Insert(
		path,
		&scopeData{
			Method: strings.ToUpper(method),
			Path:   path,
			Scopes: scopes,
		},
	)
}

// GetScopesWithPath returns the scopes of `path` and `method`, calling `t.GetScopes(...)`
// under the hood.
func (t *ScopeTree) GetScopesWithPath(path, method string) ([]types.Scope, error) {
	trie, ok := t.Get(path)
	if !ok {
		return nil, fmt.Errorf("node(%v) not existing", path)
	}

	if trie.Data == nil {
		return nil, fmt.Errorf("node(%v) without data", path)
	}

	m, ok := trie.Data.(aliasMap)
	if !ok {
		return nil, fmt.Errorf(
			"node(%v) data assert error: %v",
			path, trie.Data,
		)
	}

	upper := strings.ToUpper(method)
	scopes, ok := m[upper]
	if !ok {
		return nil, fmt.Errorf("node(%v) without method(%v) scopes", path, upper)
	}

	return scopes, nil
}

func (t *ScopeTree) String() (result string) {
	if len(t.Children) == 0 {
		return
	}

	result = "\n"
	i := 1
	for _, v := range t.Children {
		result += t.stringNode(v, "", 1, []bool{i == len(t.Children)}) + "\n"
		i++
	}
	return
}

func (t *ScopeTree) stringNode(
	node *utils.Trie, prefix string, depth int, isLastNode []bool) (str string) {
	prefix = t.Meta.TokenJoiner(prefix, node.Key)
	str = prefix
	if node.Data != nil {
		str += fmt.Sprintf("  %s", node.Data)
	} else {
		depth++
	}

	count, i := len(node.Children), 1
	for _, child := range node.Children {
		str += "\n"

		for j := 0; j < len(isLastNode); j++ {
			if isLastNode[j] {
				str += "    "
			} else {
				str += "│   "
			}
		}

		if i < count {
			str += "├──"
		} else {
			str += "└──"
		}

		nextLast := append(isLastNode, i == count)
		str += t.stringNode(child, prefix, depth+1, nextLast)
		i++
	}
	return str
}
