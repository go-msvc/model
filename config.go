package model

type IConfig interface {
	ItemStore(item IItem) (IItemStore, error)
	SubStore(sub ISub) (ISubStore, error)
}
