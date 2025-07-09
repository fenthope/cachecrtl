package cachectrl

import (
	"unsafe"
)

func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func bytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

// node 是我们树的核心节点.
type node struct {
	path     string
	indices  string
	children []*node
	value    any
}

// AddRule 向树中添加一条规则.
// path 是要匹配的前缀, value 是与该前缀关联的数据.
// 非并发安全.
func (n *node) AddRule(path string, value any) {
	if len(path) == 0 {
		return // 不允许空路径
	}

walk:
	for {
		i := longestCommonPrefix(path, n.path)

		if i < len(n.path) {
			child := &node{
				path:     n.path[i:],
				indices:  n.indices,
				children: n.children,
				value:    n.value,
			}
			n.children = []*node{child}
			n.indices = string(n.path[i])
			n.path = path[:i]
			n.value = nil
		}

		if i < len(path) {
			path = path[i:]
			c := path[0]

			for j := 0; j < len(n.indices); j++ {
				if c == n.indices[j] {
					n = n.children[j]
					continue walk
				}
			}

			n.indices += string(c)
			child := &node{
				path:  path,
				value: value,
			}
			n.children = append(n.children, child)
			return
		}

		if n.value != nil {
			panic("a value is already registered for path '" + path + "'")
		}
		n.value = value
		return
	}
}

// GetValue 查找与给定路径最匹配的值.
// 它会返回匹配到的最长前缀所关联的值.
func (n *node) GetValue(path string) (value any) {
	var bestMatch any

walk:
	for {
		// 如果当前节点有值, 它可能是一个潜在的最佳匹配.
		if n.value != nil {
			bestMatch = n.value
		}

		i := longestCommonPrefix(path, n.path)

		if i < len(n.path) {
			// 路径不匹配当前节点, 查找结束.
			return bestMatch
		}

		if i < len(path) {
			path = path[i:]
			c := path[0]
			for j := 0; j < len(n.indices); j++ {
				if c == n.indices[j] {
					n = n.children[j]
					continue walk
				}
			}
		}

		// 没有更多的子路径可以匹配了.
		return bestMatch
	}
}
