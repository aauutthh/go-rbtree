/*
Copyright 2018 Daniel Li.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific
language governing permissions and limitations under the
License.
*/
package rbtree

import (
	"errors"
	_ "fmt"
	"strings"
	"sync/atomic"
)

// <0 : less
// ==0 : equals
// >0 : large
type Comparator func(a, b interface{}) int

var IntComparator Comparator = func(a, b interface{}) int { return a.(int) - b.(int) }

type Tree struct {
	root    *Node
	cmp     Comparator // required function to order keys
	counter int64
}

type Color int8
type Direction int8

const (
	RED   Color     = 0 // 红色不增加黑路径深度
	BLACK Color     = 1
	LEFT  Direction = 1
	RIGHT Direction = 2
	NODIR Direction = 0 // found
)

var (
	ErrNilKey     = errors.New("Key is nil")
	ErrUnvalidKey = errors.New("Key is not valid")
)

type Node struct {
	key, payload        interface{}
	color               Color
	left, right, parent *Node
}

type Visitor struct{}

func NewTree(c Comparator) *Tree {
	return &Tree{root: nil, cmp: c}
}

//TODO
func checkKey(key interface{}) error {
	return nil
}

func (t *Tree) Put(key interface{}, data interface{}) error {
	//fmt.Println("Put", key, data)
	if e := checkKey(key); e != nil {
		return e
	}
	n, fix := t.InnerPut(key, data)
	if fix {
		t.fix(n)
	}
	return nil
}

type Method int

const (
	DLR Method = 1 // 前序遍历 PreOrder
	LDR Method = 2 // 中序 InOrder
	LRD Method = 3 // 后序 PostOrder
)

type FormatToString func(n *Node, level int) string

func (t *Tree) PreOrderString(fmt FormatToString) string {
	s := t.PreOrderStringAt(t.root, fmt, 0)
	return strings.Join(s, "")
}

func (t *Tree) PreOrderStringAt(n *Node, fmt FormatToString, level int) []string {
	if n == nil {
		return nil
	}
	s := fmt(n, level)
	a := []string{s}
	b := t.PreOrderStringAt(n.left, fmt, level+1)
	c := t.PreOrderStringAt(n.right, fmt, level+1)
	if b != nil || c != nil {
		a = append(a, "(")
		if b != nil {
			a = append(a, b...)
		}
		a = append(a, ")(")
		if c != nil {
			a = append(a, c...)
		}
		a = append(a, ")")
	}
	return a
}

// type WalkFunc func(key, data interface{}, level int)
//
// func (t *Tree) Walk(method Method, funcs WalkFunc) {
// 	t.WalkAt(t.root, method, funcs, 0)
// }
//
// func (t *Tree) WalkAt(n *Node, method Method, funcs WalkFunc, level int) {
// 	if n == nil {
// 		return
// 	}
// 	switch method {
// 	case DLR:
// 		funcs(n.key, n.payload, level)
// 		t.WalkAt(n.left, funcs, level+1)
// 		t.WalkAt(n.right, funcs, level+1)
// 		break
// 	case LDR:
// 		t.WalkAt(n.left, funcs, level+1)
// 		funcs(n.key, n.payload, level)
// 		t.WalkAt(n.right, funcs, level+1)
// 		break
// 	case LRD:
// 		t.WalkAt(n.left, funcs, level+1)
// 		t.WalkAt(n.right, funcs, level+1)
// 		funcs(n.key, n.payload, level)
// 		break
// 	default:
// 		return
// 	}
// }

func (t *Tree) BlackDeep() int {
	cur := t.root
	rs := 0
	for cur != nil {
		rs += int(cur.color)
		cur = cur.left
	}
	return rs
}

func (t *Tree) InnerPut(key interface{}, data interface{}) (n *Node, needfix bool) {
	t.addCounter(1)

	if t.root == nil {
		t.root = &Node{key: key, color: BLACK, payload: data}
		return t.root, false
	}

	found, parent, Direction := t.innerSearch(nil, t.root, key, NODIR)
	if found {
		if Direction == LEFT {
			parent.left.payload = data
			return parent.left, false
		} else {
			parent.right.payload = data
			return parent.right, false
		}
	} else {
		inode := &Node{key: key, parent: parent, color: RED, payload: data}
		if Direction == LEFT {
			parent.left = inode
		} else {
			parent.right = inode
		}
		return inode, true
	}
}

//TODO TEST
func (t *Tree) fix(d *Node) {
	// 当前节点不为红不需要fix
	if d.color != RED {
		return
	}

	// 其父不为红，不需要fix
	if d.color != d.parent.color {
		return
	}
	//fmt.Println("fix", d.key, d.color)

	// 叔叔节点与父节点同为红色
	if d.parent.parent.left.GetColor() == RED && d.parent.parent.right.GetColor() == RED {
		d.parent.parent.left.color = BLACK
		d.parent.parent.right.color = BLACK
		if d.parent.parent != t.root {
			d.parent.parent.color = RED
			t.fix(d.parent.parent)
		}
		return
	}

	// 本节点为左节点， 而父为右节点，父节点右旋，使父子节点为同侧
	if d.parent.left == d && d.parent.parent.right == d.parent {
		// 旋转后父变子，子变父
		d = d.parent
		t.RotateRight(d)
	}
	// 本节点为右节点， 而父为左节点，父节点左旋，使父子节点为同侧
	if d.parent.right == d && d.parent.parent.left == d.parent {
		// 旋转后父变子，子变父
		d = d.parent
		t.RotateLeft(d)
	}
	// 修复rbtree , 将父节点改为黑，从父节点开始的子树是平衡的
	d.parent.color = BLACK
	//fmt.Println("parent info", d.parent.key, d.parent.color)

	// 从爷爷节点看，叔系黑点深度为n , 父系的黑点深度变成n+1
	// 将叔叔也染黑了，从爷爷节点开始的子树也平衡，就这么干
	// 但叔叔可能本来就是黑的
	if d.parent.parent.left.GetColor() == d.parent.parent.right.GetColor() {
		// 叔叔本来就是黑的
		// 那就把爷爷变红,旋转到叔叔那边，使父节点升级到爷爷的位置, 这样从父这个点开始的子树也平衡了，且深度也变回n
		d.parent.parent.color = RED
		if d.parent.parent.left == d.parent {
			// 父在左
			t.RotateRight(d.parent.parent) // 爷爷旋转去右边
		} else {
			t.RotateLeft(d.parent.parent)
		}
		// 修复完成
		return
	} else {
		// 叔叔是红
		// 把叔叔染黑，这样叔叔的黑点深度也变也n+1
		if d.parent.parent.left == d.parent {
			// 父在左
			d.parent.parent.right.color = BLACK
		} else {
			d.parent.parent.left.color = BLACK
		}
		// 爷爷是根的话，不可能再有冲突
		if d.parent.parent != t.root {
			// 爷爷本来是黑，从外面看爷爷的深度由 n+1 变成了n+2了，而叔公却还是n+1, 爷爷变红来救世界
			d.parent.parent.color = RED
			// 爷爷可能跟他自己爷爷冲突了，再修复一次
			t.fix(d.parent.parent)
		}
		return
	}
}

func (d *Node) Direction() Direction {
	if d.parent == nil {
		return NODIR
	}
	if d.parent.left == d {
		return LEFT
	} else {
		return RIGHT
	}
}

func (d *Node) GetColor() Color {
	if d == nil {
		return BLACK
	}
	return d.color
}

func (t *Tree) addCounter(delta int64) { atomic.AddInt64(&t.counter, delta) }

func (t *Tree) innerSearch(parent *Node, cur *Node, key interface{}, dir Direction) (bool, *Node, Direction) {
	if cur == nil {
		return false, parent, dir
	}
	c := t.cmp(key, cur.key)
	switch {
	case c > 0:
		return t.innerSearch(cur, cur.right, key, RIGHT)
	case c < 0:
		return t.innerSearch(cur, cur.left, key, LEFT)
	case c == 0:
		return true, parent, dir
	}
	return false, nil, NODIR
}

func (t *Tree) GetNode(key interface{}) *Node {
	found, parent, Direction := t.innerSearch(nil, t.root, key, NODIR)
	if found {
		if parent == nil {
			return t.root
		} else {
			if Direction == LEFT {
				return parent.left
			} else {
				return parent.right
			}
		}
	}
	return nil
}

func (t *Tree) Get(key interface{}) (bool, interface{}) {
	if e := checkKey(key); e != nil {
		return false, e
	}

	node := t.GetNode(key)
	if node != nil {
		return true, node.payload
	}
	return false, nil
}

func (t *Tree) GetParent(key interface{}) {} // (found bool, parent *Node, dir Direction) {}

func (t *Tree) RotateRight(y *Node) {
	if y == nil || y.left == nil {
		return
	}
	//   p?   p?
	//    \  /
	//      y    i'm left/right child? or root?
	//     / \
	//   x   n2
	//  / \
	// c0  z

	x := y.left
	y.left = x.right    // z
	x.parent = y.parent // p

	if x.right != nil {
		x.right.parent = y
	}

	if y.parent == nil {
		t.root = x
	} else {
		if y == y.parent.left {
			y.parent.left = x
		} else {
			y.parent.right = x
		}
	}
	y.parent = x
	x.right = y
}

func (t *Tree) RotateLeft(x *Node) {
	if x == nil {
		return
	}

	//   p?   p?
	//    \  /
	//      x    i'm left/right child? or root?
	//     / \
	//   n1   y
	//		   / \
	//		  z   c0

	y := x.right
	x.right = y.left    // z
	y.parent = x.parent // p

	if y.left != nil {
		y.left.parent = x
	}

	if x.parent == nil {
		t.root = y
	} else {
		if x == x.parent.left {
			x.parent.left = y
		} else {
			x.parent.right = y
		}
	}
	x.parent = y
	y.left = x
}

func (t *Tree) Size() int64 {
	return atomic.LoadInt64(&t.counter)
}

func (t *Tree) Has(key interface{}) bool {
	found, _, _ := t.innerSearch(nil, t.root, key, NODIR)
	return found
}

func (t *Tree) Delete(key interface{}) error {
	if e := checkKey(key); e != nil {
		return e
	}
	//TODO
	return nil
}

func (t *Tree) Walk(visitor Visitor) {}
