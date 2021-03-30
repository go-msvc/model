package model

import (
	"reflect"
)

//IStore interface stores items of the same data type
type IStore interface {
	Type() reflect.Type
	Count() int
}
