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
	"fmt"
	. "gopkg.in/check.v1"
	"reflect"
	"strconv"
)

type RbTreeSuite struct{}

var _ = Suite(&RbTreeSuite{})

var funcs map[string]reflect.Method

func init() {
	nilMethod := reflect.Method{}
	funcs = map[string]reflect.Method{
		"RotateRight":          nilMethod,
		"RotateLeft":           nilMethod,
		"Put":                  nilMethod,
		"Get":                  nilMethod,
		"Delete":               nilMethod,
		"InnerPut":             nilMethod,
		"RotateLeftTestAt":     nilMethod,
		"RotateRightTestAt":    nilMethod,
		"InnerPutTestMultiKey": nilMethod,
		"PutTestMultiKey":      nilMethod,
		"NOP":                  nilMethod,
	}

	t := reflect.TypeOf(NewTree(IntComparator))
	for fname, _ := range funcs {
		f, found := t.MethodByName(fname)
		if !found {
			panic(fmt.Sprintf("No method `%s` in Tree", fname))
		}
		funcs[fname] = f
	}
}

func WithArgs(a ...interface{}) []reflect.Value {
	in := make([]reflect.Value, len(a))
	it := make([]struct{}, len(a))
	for i, _ := range it {
		in[i] = reflect.ValueOf(a[i])
	}
	return in
}

func PrepandArg(a interface{}, old []reflect.Value) []reflect.Value {
	va := reflect.ValueOf(a)
	return append([]reflect.Value{va}, old...)
}

type testCase struct {
	ops      string
	args     []reflect.Value
	checker  CheckFun
	expected interface{}
}

type CheckFun func(c *C, a, b interface{})

func MiddleTravelEquals(c *C, a, b interface{}) {
	s := fmt.Sprint(a)
	s = "(.7.)"
	c.Assert(s, Equals, b)
}

func DLRTravelEquals(c *C, a, b interface{}) {
	tree := a.(*Tree)
	s := tree.PreOrderString(func(n *Node, level int) string {
		return strconv.Itoa(n.key.(int))
	})
	c.Assert(s, Equals, b)
}

func DLRTravelWithColorEquals(c *C, a, b interface{}) {
	tree := a.(*Tree)
	s := tree.PreOrderString(func(n *Node, level int) string {
		if n.color == RED {
			return strconv.Itoa(n.key.(int)) + "R"
		} else {
			return strconv.Itoa(n.key.(int))
		}
	})
	c.Assert(s, Equals, b)
}

func BlackDeepEquals(c *C, a, b interface{}) {
	tree := a.(*Tree)
	s := tree.BlackDeep()
	c.Assert(s, Equals, b)
}

var fixtureSmall = []testCase{
	{"Put", WithArgs(7, "data-7"), MiddleTravelEquals, "(.7.)"},
	{"Put", WithArgs(3, "data-3"), MiddleTravelEquals, "((.3.)7.)"},
	{"Put", WithArgs(8, "data-8"), MiddleTravelEquals, "((.3.)7(.8.))"},
	//	{"RotateLeft", WithArgs(8, "data-8"), "((.3.)7(.8.))"},
}

var innerPut = []testCase{
	{"InnerPut", WithArgs(7, "data-7"), DLRTravelEquals, "7"},
	{"InnerPut", WithArgs(3, "data-3"), DLRTravelEquals, "7(3)()"},
	{"InnerPut", WithArgs(8, "data-8"), DLRTravelEquals, "7(3)(8)"},
	{"InnerPut", WithArgs(18, "data-18"), DLRTravelEquals, "7(3)(8()(18))"},
	{"InnerPut", WithArgs(10, "data-10"), DLRTravelEquals, "7(3)(8()(18(10)()))"},
	{"InnerPut", WithArgs(8, "pay8"), DLRTravelEquals, "7(3)(8()(18(10)()))"},
	{"InnerPut", WithArgs(11, "data-11"), DLRTravelEquals, "7(3)(8()(18(10()(11))()))"},
	{"InnerPut", WithArgs(22, "data-22"), DLRTravelEquals, "7(3)(8()(18(10()(11))(22)))"},
	{"InnerPut", WithArgs(26, "data-26"), DLRTravelEquals, "7(3)(8()(18(10()(11))(22()(26))))"},
	{"InnerPut", WithArgs(30, "data-30"), DLRTravelEquals, "7(3)(8()(18(10()(11))(22()(26()(30)))))"},
	{"InnerPut", WithArgs(45, "data-45"), DLRTravelEquals, "7(3)(8()(18(10()(11))(22()(26()(30()(45))))))"},
	{"InnerPut", WithArgs(35, "data-35"), DLRTravelEquals, "7(3)(8()(18(10()(11))(22()(26()(30()(45(35)()))))))"},
	{"InnerPut", WithArgs(90, "data-90"), DLRTravelEquals, "7(3)(8()(18(10()(11))(22()(26()(30()(45(35)(90)))))))"},
	{"InnerPut", WithArgs(85, "data-85"), DLRTravelEquals, "7(3)(8()(18(10()(11))(22()(26()(30()(45(35)(90(85)())))))))"},
	{"InnerPut", WithArgs(83, "data-83"), DLRTravelEquals, "7(3)(8()(18(10()(11))(22()(26()(30()(45(35)(90(85(83)())())))))))"},
	{"InnerPut", WithArgs(100, "data-100"), DLRTravelEquals, "7(3)(8()(18(10()(11))(22()(26()(30()(45(35)(90(85(83)())(100))))))))"},
}

func (t *Tree) NOP() {}
func (t *Tree) RotateLeftTestAt(key int) {
	n := t.GetNode(key)
	if n == nil {
		n = t.root
	}
	t.RotateLeft(n)
}
func (t *Tree) RotateRightTestAt(key int) {
	n := t.GetNode(key)
	if n == nil {
		n = t.root
	}
	t.RotateRight(n)
}

func (t *Tree) InnerPutTestMultiKey(key ...int) {
	for _, k := range key {
		t.InnerPut(k, "")
	}
}

func (t *Tree) PutTestMultiKey(key ...int) {
	for _, k := range key {
		t.Put(k, "")
	}
}

var rotateTest = []testCase{
	{"InnerPutTestMultiKey",
		WithArgs(40, 30, 60, 35, 20, 55, 70),
		DLRTravelEquals, "40(30(20)(35))(60(55)(70))"},
	{"RotateLeftTestAt", WithArgs(30), DLRTravelEquals, "40(35(30(20)())())(60(55)(70))"},
	{"RotateRightTestAt", WithArgs(35), DLRTravelEquals, "40(30(20)(35))(60(55)(70))"},
	{"RotateRightTestAt", WithArgs(-1), DLRTravelEquals, "30(20)(40(35)(60(55)(70)))"},
	{"RotateLeftTestAt", WithArgs(-1), DLRTravelEquals, "40(30(20)(35))(60(55)(70))"},
}

var PutTest = []testCase{
	{"Put", WithArgs(7, "data-7"), DLRTravelEquals, "7"},
	{"NOP", WithArgs(), BlackDeepEquals, 1},
	{"Put", WithArgs(3, "data-3"), DLRTravelEquals, "7(3)()"},
	{"NOP", WithArgs(), BlackDeepEquals, 1},
	{"Put", WithArgs(8, "data-8"), DLRTravelEquals, "7(3)(8)"},
	{"NOP", WithArgs(), BlackDeepEquals, 1},
	{"Put", WithArgs(18, "data-18"), DLRTravelEquals, "7(3)(8()(18))"},
	{"NOP", WithArgs(), BlackDeepEquals, 2},
	{"Put", WithArgs(10, "data-10"), DLRTravelEquals, "7(3)(10(8)(18))"},
	{"NOP", WithArgs(), BlackDeepEquals, 2},
	{"Put", WithArgs(16, "data-16"), DLRTravelEquals, "7(3)(10(8)(18(16)()))"},
	{"NOP", WithArgs(), BlackDeepEquals, 2},
	{"Put", WithArgs(17, "data-17"), DLRTravelEquals, "7(3)(10(8)(17(16)(18)))"},
	{"Put", WithArgs(19, "data-19"), DLRTravelEquals, "10(7(3)(8))(17(16)(18()(19)))"},
	{"Put", WithArgs(20, "data-20"), DLRTravelWithColorEquals, "10(7R(3)(8))(17R(16)(19(18R)(20R)))"},
	{"Put", WithArgs(22, "data-22"), DLRTravelWithColorEquals, "10(7(3)(8))(17(16)(19R(18)(20()(22R))))"},
	{"NOP", WithArgs(), BlackDeepEquals, 3},
	//{"Put", WithArgs(8, "pay8"), DLRTravelEquals, "7(3)(8()(18(10)()))"},
	//{"Put", WithArgs(11, "data-11"), DLRTravelEquals, "7(3)(8()(18(10()(11))()))"},
}

var MultiPutTest = []testCase{
	{"PutTestMultiKey",
		WithArgs(7, 3, 18, 10, 8, 11, 22, 26, 30, 45, 35, 90, 85, 83, 100),
		DLRTravelWithColorEquals, "10(7(3)(8))(26R(18(11)(22))(35(30)(85R(45()(83R))(90()(100R)))))"},
	{"NOP", WithArgs(), BlackDeepEquals, 3},
}
var MultiPutTest2 = []testCase{
	{"PutTestMultiKey",
		WithArgs(12, 1, 9, 2, 0, 11, 7, 19, 4, 15, 18, 5, 14, 13, 10, 16, 6, 3, 8, 17),
		DLRTravelWithColorEquals, "9(4(1R(0)(2()(3R)))(6R(5)(7()(8R))))(14(12R(11(10R)())(13))(18R(16(15R)(17R))(19)))"},
	{"NOP", WithArgs(), BlackDeepEquals, 3},
}

func runCase(t []testCase, c *C) {
	t1 := NewTree(IntComparator)
	for _, per := range t {
		method := funcs[per.ops]
		args := PrepandArg(t1, per.args)
		method.Func.Call(args)
		per.checker(c, t1, per.expected)
	}
}

func (s *RbTreeSuite) TestAnyThing(c *C) {
	//runCase(fixtureSmall, c)
	runCase(innerPut, c)
	runCase(rotateTest, c)
	runCase(PutTest, c)
	runCase(MultiPutTest, c)
	runCase(MultiPutTest2, c)
}
