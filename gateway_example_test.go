package gateway_test

import (
	"github.com/apex/go-apex"
	"github.com/tj/go-gateway"
)

type Math struct{}

type AddInput struct {
	A int `json:"a"`
	B int `json:"b"`
}

func (m *Math) Add(in *AddInput) (int, error) {
	return in.A + in.B, nil
}

func (m *Math) Sub(in *AddInput) (int, error) {
	return in.A - in.B, nil
}

func Example() {
	apex.Handle(gateway.New(&Math{}))
}
