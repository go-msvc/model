package model

import (
	"reflect"
)

//IStore interface stores items of the same data type
type IStore interface {
	Type() reflect.Type
	Count() int
}

//IItemStore is an IStore specifically for IItem
type IItemStore interface {
	IStore
	MustAdd(IItem) ID
	Add(IItem) (ID, error)
	Get(ID) IItem
	GetBy(key map[string]interface{}, limit int) []IItem
	Upd(IItem) error
	Del(ID) error
}

//ISubStore is an IStore specifically for ISub
type ISubStore interface {
	IStore
	//MustAdd(sub ISub)
	Add(sub ISub) error
	Get(parentID ID) []ISub
	//GetBy(key map[string]interface{}, limit int) []ISub
	Upd(sub ISub) error
	Del(parentID ID) error
}
