package vmath

import "fmt"

// https://bitbucket.org/prideout/pez-viewer/src/11899f6b6f02/vmath.h

type V3 struct {
    X, Y, Z float32
}

func V3New(x, y, z float32) V3 {
    v := new(V3)
    v.X = x
    v.Y = y
    v.Z = z
    return *v
}

func V3Dot(a V3, b V3) float32 {
    return (a.X * b.X) + (a.Y * b.Y) + (a.Z * b.Z)
}

func V3Cross(a V3, b V3) V3 {
    return V3New(
        (a.Y*b.Z)-(a.Z*b.Y),
        (a.Z*b.X)-(a.X*b.Z),
        (a.X*b.Y)-(a.Y*b.X))
}

func V3Add(a V3, b V3) V3 {
    return V3New(
        a.X+b.X,
        a.Y+b.Y,
        a.Z+b.Z)
}

func V3Sub(a V3, b V3) V3 {
    return V3New(
        a.X-b.X,
        a.Y-b.Y,
        a.Z-b.Z)
}

func (v V3) Clone() V3 {
    return V3New(v.X, v.Y, v.Z)
}

func (v V3) Length() float32 {
    return sqrt(V3Dot(v,v))
}

func (v V3) String() string {
    return fmt.Sprint(v.X, ", ", v.Y, ", ", v.Z)
}
