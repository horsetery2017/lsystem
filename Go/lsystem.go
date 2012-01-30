package main

import (
    "bufio"
    "encoding/xml"
    "fmt"
    "io"
    "math/rand"
    "strconv"
    "strings"
    "vmath"
)

type CurvePoint struct {
    P   vmath.P3
    N   vmath.V3
}

type Curve []CurvePoint

// evaluates the rules in the given XML stream and returns a list of curves
func Evaluate(stream io.Reader) Curve {

    var curve Curve
    var lsys LSystem
    if err := xml.Unmarshal(stream, &lsys); err != nil {
        fmt.Println("Error parsing XML file:", err)
        return curve
    }

    // Parse the transform strings
    lsys.Matrices = make(MatrixCache)
    for _, rule := range lsys.Rules {
        for _, call := range rule.Calls {
            lsys.Matrices.ParseString(call.Transforms)
        }
        for _, inst := range rule.Instances {
            lsys.Matrices.ParseString(inst.Transforms)
        }
    }

    lsys.WeightSum = 0
    for _, rule := range lsys.Rules {
        if rule.Weight != 0 {
            lsys.WeightSum += rule.Weight
        } else {
            lsys.WeightSum++
        }
    }

    random := rand.New(rand.NewSource(42))
    start := StackNode{
        RuleIndex: lsys.PickRule("entry", random),
        Transform: vmath.M4Identity(),
    }

    lsys.ProcessRule(start, &curve, random)
    return curve
}

type LSystem struct {
    MaxDepth  int    `xml:"max_depth,attr"`
    Rules     []Rule `xml:"rule"`
    WeightSum int
    Matrices  MatrixCache
}

type Rule struct {
    Name      string     `xml:"name,attr"`
    Calls     []Call     `xml:"call"`
    Instances []Instance `xml:"instance"`
    MaxDepth  int        `xml:"max_depth,attr"`
    Successor string     `xml:"successor,attr"`
    Weight    int        `xml:"weight,attr"`
}

type Call struct {
    Transforms string `xml:"transforms,attr"`
    Rule       string `xml:"rule,attr"`
    Count      int    `xml:"count,attr"`
}

type Instance struct {
    Transforms string `xml:"transforms,attr"`
    Shape      string `xml:"string"`
}

func radians(degrees float32) float32 {
    return degrees * 3.1415926535 / 180.0
}

func (self *LSystem) ProcessRule(start StackNode, curve *Curve, random *rand.Rand) {

    stack := new(Stack)
    stack.Push(start)

    for stack.Len() > 0 {
        e := stack.Pop()

        localMax := self.MaxDepth
        rule := self.Rules[e.RuleIndex]

        if rule.MaxDepth > 0 {
            localMax = rule.MaxDepth
        }

        if stack.Len() >= self.MaxDepth {
            *curve = append(*curve, CurvePoint{})
            continue
        }

        matrix := e.Transform
        if e.Depth >= localMax {
            // Switch to a different rule is one is specified
            if rule.Successor != "" {
                next := StackNode{
                    RuleIndex: self.PickRule(rule.Successor, random),
                    Transform: matrix,
                }
                stack.Push(next)
            }
            *curve = append(*curve, CurvePoint{})
            continue
        }

        for _, call := range rule.Calls {
            t := self.Matrices[call.Transforms]
            count := call.Count
            if count == 0 {
                count = 1
            }
            for ; count != 0; count-- {
                matrix = matrix.MulM4(&t)
                newRule := self.PickRule(call.Rule, random)
                next := StackNode{
                    RuleIndex: newRule,
                    Depth:     e.Depth + 1,
                    Transform: matrix,
                }
                stack.Push(next)
            }
        }

        for _, instance := range rule.Instances {
            t := self.Matrices[instance.Transforms]
            matrix = matrix.MulM4(&t)
			p := vmath.P3FromV3(matrix.GetTranslation())
			n := vmath.V3New(0, 0, 1)
			n = matrix.GetUpperLeft().MulV3(n)
			c := CurvePoint{P: p, N:n}
            *curve = append(*curve, c)
			if len(*curve) % 10000 == 0 {
				fmt.Printf("Instanced %d nodes\n", len(*curve))
			}
        }
    }
}

func (self *LSystem) PickRule(name string, random *rand.Rand) int {
    n := random.Intn(self.WeightSum)
    for i, rule := range self.Rules {
        weight := rule.Weight
        if weight == 0 {
            weight = 1
        }
        if n < weight {
            return i
        }
        n -= weight
    }
    return -1
}

type MatrixCache map[string]vmath.M4

// Parse a string in the xform language and add the resulting matrix to the lookup table.
// Examples:
//   "rx -2 tx 0.1 sa 0.996"
//   "s 0.55 2.0 1.25"
func (cache *MatrixCache) ParseString(s string) {
    if len(s) == 0 {
        return
    }
    reader := bufio.NewReader(strings.NewReader(s))

    xform := vmath.M4Identity()

    readFloat := func() float32 {
        sx, _ := reader.ReadString(' ')
        fx, _ := strconv.ParseFloat(strings.TrimSpace(sx), 32)
        return float32(fx)
    }

    for {
        token, err := reader.ReadString(' ')
        token = strings.TrimSpace(token)
        switch token {
        case "s":
            x := readFloat()
            y := readFloat()
            z := readFloat()
            m := vmath.M4Scale(x, y, z)
            xform = xform.MulM4(m)
        case "t":
            x := readFloat()
            y := readFloat()
            z := readFloat()
            m := vmath.M4Translate(x, y, z)
            xform = xform.MulM4(m)
        case "sa":
            a := readFloat()
            m := vmath.M4Scale(a, a, a)
            xform = xform.MulM4(m)
        case "tx":
            x := readFloat()
            m := vmath.M4Translate(x, 0, 0)
            xform = xform.MulM4(m)
        case "ty":
            y := readFloat()
            m := vmath.M4Translate(0, y, 0)
            xform = xform.MulM4(m)
        case "tz":
            z := readFloat()
            m := vmath.M4Translate(0, 0, z)
            xform = xform.MulM4(m)
        case "rx":
            x := readFloat()
            m := vmath.M4RotateX(x)
            xform = xform.MulM4(m)
        case "ry":
            y := readFloat()
            m := vmath.M4RotateY(y)
            xform = xform.MulM4(m)
        case "rz":
            z := readFloat()
            m := vmath.M4RotateZ(z)
            xform = xform.MulM4(m)
        case "":
        default:
            fmt.Println("Unknown token: ", token)
        }
        if err != nil {
            break
        }
    }
    (*cache)[s] = *xform
    return
}
