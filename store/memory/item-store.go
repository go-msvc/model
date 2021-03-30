package memory

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/go-msvc/model"
	"github.com/go-msvc/msf/logger"
)

var log = logger.New("model").New("store").New("memory").WithLevel(logger.LevelError)

func New(item model.IItem) model.IItemStore {
	s := &itemStore{
		itemType: reflect.TypeOf(item),
		nextID:   1,
		items:    map[model.ID]model.IItem{},
	}
	return s
}

//implements model.IItemStore in memory
type itemStore struct {
	sync.Mutex
	itemType reflect.Type
	nextID   model.ID
	items    map[model.ID]model.IItem
}

func (s *itemStore) Type() reflect.Type {
	return s.itemType
}

func (s *itemStore) MustAdd(item model.IItem) model.ID {
	id, err := s.Add(item)
	if err != nil {
		panic(err)
	}
	return id
}

func (s *itemStore) Add(item model.IItem) (model.ID, error) {
	if item == nil {
		return model.ID(0), fmt.Errorf("cannot add nil")
	}
	if reflect.TypeOf(item) != s.Type() {
		return model.ID(0), fmt.Errorf("cannot add %v != %v", reflect.TypeOf(item), s.Type())
	}
	s.Lock()
	defer s.Unlock()
	id := s.nextID
	s.nextID++

	itemWithIDPtrValue := reflect.New(s.itemType)
	itemWithIDPtrValue.Elem().Set(reflect.ValueOf(item))
	itemWithIDPtrValue.Elem().FieldByIndex([]int{0, 0}).Set(reflect.ValueOf(id))
	s.items[id] = itemWithIDPtrValue.Elem().Interface().(model.IItem)
	return id, nil
}

func (s *itemStore) Get(id model.ID) model.IItem {
	s.Lock()
	defer s.Unlock()
	if item, ok := s.items[id]; ok {
		return item
	}
	return nil
}

var itemInterfaceType = reflect.TypeOf(new(model.IItem)).Elem()

func (s *itemStore) GetBy(key map[string]interface{}, limit int) []model.IItem {
	// log.Debugf("%T.GetBy(%+v) (limit=%d)...", s, key, limit)
	//iterate over all items to find:
	items := []model.IItem{}
	for _, item := range s.items {
		itemValue := reflect.ValueOf(item)
		// log.Debugf("checking item=%+v", itemValue)
		match := true
		for keyName, keyValue := range key {
			fieldValue := itemValue.FieldByName(keyName)
			// log.Debugf("field(%s)=%+v (exp:%+v)", keyName, fieldValue, keyValue)
			if fieldValue.Interface() != reflect.ValueOf(keyValue).Interface() {
				match = false
				break
			}
		}
		if match {
			items = append(items, item)
		}
	}
	return items
}

func (s *itemStore) Upd(item model.IItem) error {
	if item == nil {
		return fmt.Errorf("cannot upd nil")
	}
	s.Lock()
	defer s.Unlock()
	id := model.ItemID(item)
	if _, ok := s.items[id]; !ok {
		return fmt.Errorf("item.id=%d not found", id)
	}
	s.items[id] = item
	return nil
}

func (s *itemStore) Del(id model.ID) error {
	s.Lock()
	defer s.Unlock()
	delete(s.items, id)
	return nil
}

func (s *itemStore) Count() int {
	return len(s.items)
}

//todo: add indexes - but memory is not intended for high volume, mainly for testing, so not yet needed
