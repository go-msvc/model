package memory

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/go-msvc/model"
)

//subType implements model.ISub, and embeds model.Sub as first field
func newSubStore(sub model.ISub) (model.ISubStore, error) {
	return &subStore{
		subType: reflect.TypeOf(sub),
		subs:    map[model.ID][]model.ISub{},
	}, nil
}

//implements model.ISubStore in memory
type subStore struct {
	sync.Mutex
	subType reflect.Type
	subs    map[model.ID][]model.ISub
}

func (ss *subStore) Type() reflect.Type {
	return ss.subType
}

func (ss *subStore) Count() int {
	count := 0
	for _, parentSubs := range ss.subs {
		count += len(parentSubs)
	}
	return count
}

func (ss *subStore) Add(sub model.ISub) error {
	if sub == nil {
		return fmt.Errorf("cannot add nil")
	}
	if reflect.TypeOf(sub) != ss.subType {
		return fmt.Errorf("cannot add %v != %v", reflect.TypeOf(sub), ss.subType)
	}
	ss.Lock()
	defer ss.Unlock()
	copyOfSubPtrValue := reflect.New(ss.subType)
	copyOfSubPtrValue.Elem().Set(reflect.ValueOf(sub))
	copyOfSub := copyOfSubPtrValue.Elem().Interface().(model.ISub)

	parentID := reflect.ValueOf(sub).FieldByIndex([]int{0, 0}).Interface().(model.ID)
	if ss.subs[parentID] == nil {
		ss.subs[parentID] = []model.ISub{copyOfSub}
	} else {
		ss.subs[parentID] = append(ss.subs[parentID], copyOfSub)
	}
	return nil
}

func (ss *subStore) Get(parentID model.ID) []model.ISub {
	if parentSubs, ok := ss.subs[parentID]; ok {
		return parentSubs
	}
	return nil
}

//func (ss subStore) GetBy(key map[string]interface{}, limit int) []ISub
func (ss *subStore) Upd(sub model.ISub) error {
	return fmt.Errorf("NYI")
}

func (ss *subStore) Del(parentID model.ID) error {
	delete(ss.subs, parentID)
	return nil
}
