package nursery

import (
	"fmt"
)

func NurseryExample() {
	fmt.Println(runNursery())
}

func runNursery() error {
	n := New()
	defer n.Join()

	go taskA(n.Branch())
	go taskB(n.Branch())

	return n.Join()
}

func taskA(b Branch) {
	// do something in parallel
	defer b.Join()
}

func taskA(b Branch) {
	// do something in parallel
	b.Fail(fmt.Errorf("an error"))

	defer b.Join()
}
