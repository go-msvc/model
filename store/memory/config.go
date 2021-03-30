package memory

import "github.com/go-msvc/model"

type Config struct{}

func (c Config) ItemStore(item model.IItem) (model.IItemStore, error) {
	return New(item), nil
}

func (c Config) SubStore(sub model.ISub) (model.ISubStore, error) {
	panic("NYI")
	//return nil
	//return New(sub)
}
