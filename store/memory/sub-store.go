package memory

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/go-msvc/model"
)

//subType implements model.ISub, and embeds model.Sub as first field
func newSubStore(is *itemStore, subType reflect.Type) (*subStore, error) {
	return &subStore{
		itemStore: is,
		subType:   subType,
		subs:      map[model.ID][]model.ISub{},
	}, nil
}

//implements model.ISubStore in memory
type subStore struct {
	sync.Mutex
	itemStore *itemStore //parent of sub-store, not yet supporting sub of sub because sub has not simple unique id... need some thought
	subType   reflect.Type
	subs      map[model.ID][]model.ISub
}

func (ss *subStore) Add() error {
	return fmt.Errorf("NYI")
}
