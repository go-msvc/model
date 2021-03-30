package model

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
