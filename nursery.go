package nursery

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

type Nursery interface {
	Branch() Branch
	Cancel()
	Join() error
}

type Branch interface {
	context.Context
	Fail(error)
	Join()
}

type nursery struct {
	context.Context
	cancel   context.CancelFunc
	results  chan error
	branches int
}

type branch struct {
	context.Context
	res chan<- error
}

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
	var err error
	for n.branches > 0 {
		e := <-n.results
		if e != nil {
			n.cancel()
			err = multierror.Append(err, e)
		}
	}
	return err
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
