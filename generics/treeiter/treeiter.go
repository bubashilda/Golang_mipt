//go:build !solution

package treeiter

type node[T any] interface {
	Left() *T
	Right() *T
}

func DoInOrder[T node[T]](root *T, cb func(t *T)) {
	if root == nil {
		return
	}

	if (*root).Left() != nil {
		DoInOrder((*root).Left(), cb)
	}

	cb(root)

	if (*root).Right() != nil {
		DoInOrder((*root).Right(), cb)
	}
}
