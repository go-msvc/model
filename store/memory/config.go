package memory

import "github.com/go-msvc/model"

type Config struct{}

func (c Config) ItemStore(item model.IItem) (model.IItemStore, error) {
	return New(item), nil
}

func (c Config) SubStore(sub model.ISub) (model.ISubStore, error) {
	return newSubStore(sub)
}
