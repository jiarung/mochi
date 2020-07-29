package utils

// TrieTokenizer defines tokenizer helper.
type TrieTokenizer func(string) []string

// TrieTokenJoiner defines token join helper.
type TrieTokenJoiner func(...string) string

// TrieKeyFormatter defines function to process given key/token.
type TrieKeyFormatter func(string) string

// TrieDataHandler defines data handler function.
// First param is target node, second param is data.
type TrieDataHandler func(*Trie, interface{}) error

// TrieDataMode defines the policy to insert/delete data.
type TrieDataMode int

// DataMode enum.
const (
	// DataLeaf affects data only on leaf.
	DataLeaf TrieDataMode = iota
	// DataPath affects data on all nodes of walking path and leaf.
	DataPath TrieDataMode = iota
)

const (
	// TrieRootIdentifier is key to dump all paths with Find.
	TrieRootIdentifier = "\t\n\t\n\t\n"
)

// TrieMeta collects shared functions and params if any.
type TrieMeta struct {
	KeyFormatter TrieKeyFormatter `json:"-"`

	Tokenizer   TrieTokenizer   `json:"-"`
	TokenJoiner TrieTokenJoiner `json:"-"`

	DataInsertionFn TrieDataHandler `json:"-"`
	DataDeletionFn  TrieDataHandler `json:"-"`
	DefautlDataMode TrieDataMode

	ArbitraryKey string
}

// Copy returns a new meta struct.
func (m *TrieMeta) Copy() (meta *TrieMeta) {
	meta = &TrieMeta{
		KeyFormatter:    m.KeyFormatter,
		Tokenizer:       m.Tokenizer,
		TokenJoiner:     m.TokenJoiner,
		DataInsertionFn: m.DataInsertionFn,
		DataDeletionFn:  m.DataDeletionFn,
		DefautlDataMode: m.DefautlDataMode,
		ArbitraryKey:    m.ArbitraryKey,
	}
	return
}

// Trie node struct.
type Trie struct {
	Key string

	Children map[string]*Trie

	Data interface{}

	Meta *TrieMeta
}

// Insert given values into trie tree.
func (t *Trie) Insert(
	value string, data interface{}, mode ...TrieDataMode) (err error) {
	err = t.insert(t.Meta.Tokenizer(value), data, mode...)
	return
}

func (t *Trie) insert(
	tokens []string, data interface{}, mode ...TrieDataMode) (err error) {
	if len(tokens) == 0 {
		return
	}

	token := t.Meta.KeyFormatter(tokens[0])

	next, ok := t.Children[token]
	if !ok {
		next = &Trie{
			Key:      token,
			Children: make(map[string]*Trie),
			Meta:     t.Meta,
		}
	}
	t.Children[token] = next

	dataMode := t.Meta.DefautlDataMode
	if len(mode) > 0 {
		dataMode = mode[0]
	}

	if dataMode == DataPath {
		err = t.Meta.DataInsertionFn(t, data)
		if err != nil {
			return
		}
	}

	if len(tokens) == 1 {
		err = next.Meta.DataInsertionFn(next, data)
		return
	}

	err = next.insert(tokens[1:], data, mode...)
	return
}

// Get given values in trie tree or not.
func (t *Trie) Get(value string) (found *Trie, ok bool) {
	tokens := t.Meta.Tokenizer(value)
	if len(tokens) == 0 {
		return
	}

	found, ok = t.walk(tokens)
	return
}

func (t *Trie) walk(tokens []string) (found *Trie, ok bool) {
	if len(tokens) == 0 {
		return
	}

	token := t.Meta.KeyFormatter(tokens[0])

	found, ok = t.Children[token]
	if !ok {
		arbitraryNode, exists := t.Children[t.Meta.ArbitraryKey]
		if !exists {
			return
		}

		if len(tokens) == 1 {
			found = arbitraryNode
			ok = true
			return
		}

		found, ok = arbitraryNode.walk(tokens[1:])
		return
	}

	if len(tokens) == 1 {
		return
	}

	found, ok = found.walk(tokens[1:])
	return
}

// Find possible result with given prefix.
func (t *Trie) Find(prefix string) (results map[string]*Trie) {
	results = make(map[string]*Trie)

	if prefix == TrieRootIdentifier {
		results = t.collect("")
		return
	}

	tokens := t.Meta.Tokenizer(prefix)
	if len(tokens) == 0 {
		return
	}

	found, ok := t.walk(tokens)
	if !ok {
		return
	}

	results = found.collect(prefix)
	return
}

func (t *Trie) collect(prefix string) (results map[string]*Trie) {
	results = make(map[string]*Trie)
	if t.Data != nil {
		results[prefix] = t
	}

	for key, child := range t.Children {
		updates := child.collect(t.Meta.TokenJoiner(prefix, key))
		t.merge(results, updates)
	}
	return
}

func (t *Trie) merge(sources, updates map[string]*Trie) {
	for k, v := range updates {
		sources[k] = v
	}
}

// Delete given value.
func (t *Trie) Delete(value string, mode ...TrieDataMode) (err error) {
	_, err = t.delete(t.Meta.Tokenizer(value), mode...)
	return

}

// recursive call child's delete and return leaf data for updating.
func (t *Trie) delete(
	tokens []string, mode ...TrieDataMode) (data interface{}, err error) {
	if len(tokens) == 0 {
		return
	}

	token := t.Meta.KeyFormatter(tokens[0])

	child, ok := t.Children[token]
	if !ok {
		// ignore not-existing node deletion error
		return
	}

	dataMode := t.Meta.DefautlDataMode
	if len(mode) > 0 {
		dataMode = mode[0]
	}

	data, err = child.delete(tokens[1:], mode...)
	if err != nil {
		return
	}

	if dataMode == DataPath {
		err = t.Meta.DataInsertionFn(t, data)
		if err != nil {
			return
		}
	}

	if len(child.Children) == 0 {
		delete(t.Children, token)
	}
	return
}

// Copy return a new Trie.
func (t *Trie) Copy() (trie *Trie) {
	trie = t.copy(t.Meta.Copy())
	return
}

func (t *Trie) copy(meta *TrieMeta) (trie *Trie) {
	trie = &Trie{
		Key:      t.Key,
		Children: make(map[string]*Trie),
		Meta:     meta,
		Data:     t.Data,
	}

	for token, oldTrie := range t.Children {
		trie.Children[token] = oldTrie.copy(meta)
	}
	return
}
