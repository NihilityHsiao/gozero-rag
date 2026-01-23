package prompt

import "testing"

func TestPrompt(t *testing.T) {

	p, err := NewGraphPrompt("eino adk构建第一个AI智能体")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("prompt:%v", p)
}
