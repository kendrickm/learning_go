package apt

//Abstract Picture Tree
import (
	"math"
	"math/rand"
	"reflect"
	"strconv"

	"github.com/kendrickm/learning_go/noise"
)

type Node interface {
	Eval(x, y float32) float32
	String() string
	AddRandom(node Node)
	SetParent(node Node)
	SetChildren([]Node)
	GetParent() Node
	GetChildren() []Node
	AddLeaf(leaf Node) bool
	NodeCount() int
}

type BaseNode struct {
	Parent   Node
	Children []Node
}

func (node *BaseNode) Eval(x, y float32) float32 {
	panic("tried to call eval on basenode")
}

func (node *BaseNode) String() string {
	panic("tried to call eval on basenode")
}

func (node *BaseNode) AddRandom(nodeToAdd Node) {
	addIndex := rand.Intn(len(node.Children))
	if node.Children[addIndex] == nil {
		nodeToAdd.SetParent(node)
		node.Children[addIndex] = nodeToAdd
	} else {
		node.Children[addIndex].AddRandom(nodeToAdd)
	}
}

func (node *BaseNode) SetParent(parent Node) {
	node.Parent = parent
}

func (node *BaseNode) GetParent() Node {
	return node.Parent
}

func (node *BaseNode) GetChildren() []Node {
	return node.Children
}

func (node *BaseNode) AddLeaf(leaf Node) bool {
	for i, child := range node.Children {
		if child == nil {
			leaf.SetParent(node)
			node.Children[i] = leaf
			return true
		} else if node.Children[i].AddLeaf(leaf) {
			return true
		}
	}
	return false
}

func (node *BaseNode) NodeCount() int {
	count := 1
	for _, child := range node.Children {
		if child == nil {
			panic("Found a nil child")
		}
		count += child.NodeCount()
	}
	return count
}

func CopyTree(node Node, parent Node) Node {
	copy := reflect.New(reflect.ValueOf(node).Elem().Type()).Interface().(Node)
	switch n := node.(type) {
	case *OpConstant:
		copy.(*OpConstant).value = n.value
	}
	copy.SetParent(parent)
	copyChildren := make([]Node, len(node.GetChildren()))
	copy.SetChildren(copyChildren)
	for i := range copyChildren {
		copyChildren[i] = CopyTree(node.GetChildren()[i], copy)
	}

	return copy
}

func ReplaceNode(old Node, new Node) {
	oldParent := old.GetParent()
	if oldParent != nil {
		for i, child := range oldParent.GetChildren() {
			if child == old {
				oldParent.GetChildren()[i] = new
			}
		}
	}
	new.SetParent(oldParent)
}

func GetNthNode(node Node, n, count int) (Node, int) {
	if n == count {
		return node, count
	}
	var result Node
	for _, child := range node.GetChildren() {
		count++
		result, count = GetNthNode(child, n, count)
		if result != nil {
			return result, count
		}
	}
	return nil, count
}

func (node *BaseNode) SetChildren(children []Node) {
	node.Children = children
}

func Mutate(node Node) Node {
	rand := rand.Intn(17)
	var mutatedNode Node
	if rand <= 13 {
		mutatedNode = GetRandomNode()
	} else {
		mutatedNode = GetRandomLeafNode()
	}

	// Fix up the child's parent pointer to point to our new node
	if node.GetParent() != nil {
		for i, parentChild := range node.GetParent().GetChildren() {
			if parentChild == node {
				node.GetParent().GetChildren()[i] = mutatedNode
			}
		}
	}

	// Add children from the old node to the new node
	for i, child := range node.GetChildren() {
		if i >= len(mutatedNode.GetChildren()) {
			break
		}
		mutatedNode.GetChildren()[i] = child
		child.SetParent(mutatedNode)
	}

	// Any nil children get filled in with random leaf nodes
	for i, child := range mutatedNode.GetChildren() {
		if child == nil {
			leaf := GetRandomLeafNode()
			mutatedNode.GetChildren()[i] = leaf
		}
	}

	mutatedNode.SetParent(node.GetParent())
	return mutatedNode
}

type OpLerp struct {
	BaseNode
}

func NewOpLerp() *OpLerp {
	return &OpLerp{BaseNode{nil, make([]Node, 3)}}
}

// func NewOp() *Op {
// 	return &Op{BaseNode(nil, make([]Node))}
// }

func (op *OpLerp) Eval(x, y float32) float32 {
	a := op.Children[0].Eval(x, y)
	b := op.Children[1].Eval(x, y)
	pct := op.Children[2].Eval(x, y)

	return a + pct*(b-a)
}

func (op *OpLerp) String() string {
	return "( Lerp " + op.Children[0].String() + " " + op.Children[1].String() + " " + op.Children[2].String() + " )"
}

type OpSin struct {
	BaseNode
}

func NewOpSin() *OpSin {
	return &OpSin{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpSin) Eval(x, y float32) float32 {
	return float32(math.Sin(float64(op.Children[0].Eval(x, y))))
}

func (op *OpSin) String() string {
	return "( Sin " + op.Children[0].String() + ")"
}

type OpCos struct {
	BaseNode
}

func NewOpCos() *OpCos {
	return &OpCos{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpCos) Eval(x, y float32) float32 {
	return float32(math.Cos(float64(op.Children[0].Eval(x, y))))
}

func (op *OpCos) String() string {
	return "( Cos " + op.Children[0].String() + ")"
}

type OpPlus struct {
	BaseNode
}

func NewOpPlus() *OpPlus {
	return &OpPlus{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpPlus) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) + op.Children[1].Eval(x, y)
}

func (op *OpPlus) String() string {
	return "( + " + op.Children[0].String() + "  " + op.Children[1].String() + " )"
}

type OpMinus struct {
	BaseNode
}

func NewOpMinus() *OpMinus {
	return &OpMinus{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpMinus) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) - op.Children[1].Eval(x, y)
}

func (op *OpMinus) String() string {
	return "( - " + op.Children[0].String() + "  " + op.Children[1].String() + " )"
}

type OpMult struct {
	BaseNode
}

func NewOpMult() *OpMult {
	return &OpMult{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpMult) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) * op.Children[1].Eval(x, y)
}

func (op *OpMult) String() string {
	return "( * " + op.Children[0].String() + "  " + op.Children[1].String() + " )"
}

type OpDiv struct {
	BaseNode
}

func NewOpDiv() *OpDiv {
	return &OpDiv{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpDiv) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) / op.Children[1].Eval(x, y)
}

func (op *OpDiv) String() string {
	return "( / " + op.Children[0].String() + "  " + op.Children[1].String() + " )"
}

type OpAtan2 struct {
	BaseNode
}

func NewOpAtan2() *OpAtan2 {
	return &OpAtan2{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpAtan2) Eval(x, y float32) float32 {
	return float32(math.Atan2(float64(op.Children[0].Eval(x, y)), float64(op.Children[1].Eval(x, y))))
}

func (op *OpAtan2) String() string {
	return "( Atan2 " + op.Children[0].String() + "  " + op.Children[1].String() + " )"
}

type OpAtan struct {
	BaseNode
}

func NewOpAtan() *OpAtan {
	return &OpAtan{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpAtan) Eval(x, y float32) float32 {
	return float32(math.Atan(float64(op.Children[0].Eval(x, y))))
}

func (op *OpAtan) String() string {
	return "( Atan " + op.Children[0].String() + " )"
}

type OpNoise struct {
	BaseNode
}

func NewOpNoise() *OpNoise {
	return &OpNoise{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpNoise) Eval(x, y float32) float32 {
	return 80*noise.Snoise2(op.Children[0].Eval(x, y), op.Children[1].Eval(x, y)) - 2.0
}

func (op *OpNoise) String() string {
	return "( SimplexNoise " + op.Children[0].String() + "  " + op.Children[1].String() + " )"
}

type OpSquare struct {
	BaseNode
}

func NewOpSquare() *OpSquare {
	return &OpSquare{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpSquare) Eval(x, y float32) float32 {
	return float32(math.Sqrt(float64(op.Children[0].Eval(x, y))))
}

func (op *OpSquare) String() string {
	return "( Sqrt " + op.Children[0].String() + ")"
}

type OpLog2 struct {
	BaseNode
}

func NewOpLog2() *OpLog2 {
	return &OpLog2{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpLog2) Eval(x, y float32) float32 {
	return float32(math.Log2(float64(op.Children[0].Eval(x, y))))
}

func (op *OpLog2) String() string {
	return "( Log2 " + op.Children[0].String() + ")"
}

type OpNegate struct {
	BaseNode
}

func NewOpNegate() *OpNegate {
	return &OpNegate{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpNegate) Eval(x, y float32) float32 {
	return -1 * op.Children[0].Eval(x, y)
}

func (op *OpNegate) String() string {
	return "( - " + op.Children[0].String() + ")"
}

type OpCeil struct {
	BaseNode
}

func NewOpCeil() *OpCeil {
	return &OpCeil{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpCeil) Eval(x, y float32) float32 {
	return float32(math.Ceil(float64(op.Children[0].Eval(x, y))))
}

func (op *OpCeil) String() string {
	return "( Ceiling " + op.Children[0].String() + ")"
}

type OpFloor struct {
	BaseNode
}

func NewOpFloor() *OpFloor {
	return &OpFloor{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpFloor) Eval(x, y float32) float32 {
	return float32(math.Floor(float64(op.Children[0].Eval(x, y))))
}

func (op *OpFloor) String() string {
	return "( Floor " + op.Children[0].String() + ")"
}

type OpX struct {
	BaseNode
}

func NewOpX() *OpX {
	return &OpX{BaseNode{nil, make([]Node, 0)}}
}

func (op *OpX) Eval(x, y float32) float32 {
	return x
}

func (op *OpX) String() string {
	return "X"
}

type OpY struct {
	BaseNode
}

func NewOpY() *OpY {
	return &OpY{BaseNode{nil, make([]Node, 0)}}
}

func (op *OpY) Eval(x, y float32) float32 {
	return y
}

func (op *OpY) String() string {
	return "Y"
}

type OpConstant struct {
	BaseNode
	value float32
}

func NewOpConstant() *OpConstant {
	x := &OpConstant{BaseNode{nil, make([]Node, 0)}, rand.Float32()*2 - 1}
	return x
}

func (op *OpConstant) Eval(x, y float32) float32 {
	return op.value
}

func (op *OpConstant) String() string {
	return "Constant: " + strconv.FormatFloat(float64(op.value), 'f', 9, 32)
}

type OpClip struct {
	BaseNode
}

func NewOpClip() *OpClip {
	return &OpClip{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpClip) Eval(x, y float32) float32 {
	value := op.Children[0].Eval(x, y)
	max := float32(math.Abs(float64(op.Children[1].Eval(x, y))))
	if value > max {
		return max
	} else if value < -max {
		return -max
	}
	return value
}

func GetRandomNode() Node {
	r := rand.Intn(14)
	switch r {
	case 0:
		return NewOpPlus()
	case 1:
		return NewOpMinus()
	case 2:
		return NewOpMult()
	case 3:
		return NewOpDiv()
	case 4:
		return NewOpAtan2()
	case 5:
		return NewOpAtan()
	case 6:
		return NewOpCos()
	case 7:
		return NewOpSin()
	case 8:
		return NewOpNoise()
	case 9:
		return NewOpSquare()
	case 10:
		return NewOpLog2()
	case 11:
		return NewOpNegate()
	case 12:
		return NewOpCeil()
	case 13:
		return NewOpFloor()
	case 14:
		return NewOpClip()
	}
	panic("Get Random Node failed")
}

func GetRandomLeafNode() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return NewOpX()
	case 1:
		return NewOpY()
	case 2:
		return NewOpConstant()
	}
	panic("Get Random Leaf Node failed")
}
