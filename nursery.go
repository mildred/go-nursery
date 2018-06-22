package nursery

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// Nursery if an object that reprensents the lifecycle of multiple goroutines.
// When the nursery is joined, its guaranteed that all its branches are
// finished.
type Nursery interface {

	// The nursery inherits from Context that gives a time frame for the main
	// branch of the nursery
	context.Context

	// Branch branches the nursery
	Branch() Branch

	// Cancel cancels the nursery context and all its branches
	Cancel()

	// Join makes sure that all branches have joined and returns any errors that
	// happened from any of them. Can be called more than once and will return the
	// same errors each time.
	Join() error
}

// Branch is an object that represents a goroutine issues by a Nursery. All
// branches needs to be joined before the Nursery terminates. The Nursery takes
// care of handling errors from all its branches.
type Branch interface {

	// The branch inherits from the nursery context.
	context.Context

	// Fail will terminate the current branch by panicking and will return the
	// error to the nursery
	Fail(error)

	// Join will make sure that the branch returns the branch status to the
	// nursery when completed. Must be used with defer.
	Join()
}

type nursery struct {
	context.Context
	cancel   context.CancelFunc
	results  chan error
	branches int
	errors   error
}

type branch struct {
	context.Context
	res chan<- error
}

// New creates a new nursery
func New(ctx0 context.Context) Nursery {
	ctx, cancel := context.WithCancel(ctx0)
	return &nursery{ctx, cancel, make(chan error), 0}
}

func (n *nursery) Branch() Branch {
	n.branches++
	return &branch{
		Context: n.Context,
		res:     n.results,
	}
}

func (n *nursery) Cancel() {
	n.cancel()
}

func (n *nursery) Join() error {
	for n.branches > 0 {
		e := <-n.results
		if e != nil {
			n.cancel()
			n.errors = multierror.Append(n.errors, e)
		}
	}
	return n.errors
}

func (b *branch) Fail(err error) {
	panic(err)
}

func (b *branch) Join() {
	switch err := recover().(type) {
	case error:
		b.res <- err
	default:
		b.res <- fmt.Errorf("%v", err)
	}
}
