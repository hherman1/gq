package gq

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNode(t *testing.T) {
	sample := `{
  "test": [1, 2, 3, {"1": "b"}]
}`
	var n Node
	err := json.Unmarshal([]byte(sample), &n)
	if err != nil {
		t.Fatalf("unmarshal sample: %v", err)
	}
	fmt.Println(n)
}

func TestNodeGet(t *testing.T) {
	sample := `{
  "test": [1, 2, 3, {"1": "b"}]
}`
	var n Node
	err := json.Unmarshal([]byte(sample), &n)
	if err != nil {
		t.Fatalf("unmarshal sample: %v", err)
	}
	fmt.Println(n.G("test"))
	fmt.Println(n.G("test").G("1"))
}
