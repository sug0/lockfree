// Copyright 2020 The golang.design Initiative authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.

package lockfree

import (
	"sync/atomic"
	"unsafe"
)

type directItem[T any] struct {
	next unsafe.Pointer
	v    *T
}

func loaditem[T any](p *unsafe.Pointer) *directItem[T] {
	return (*directItem[T])(atomic.LoadPointer(p))
}

func casitem[T any](p *unsafe.Pointer, old, new *directItem[T]) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
