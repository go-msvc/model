package model

//ISubStore is an IStore specifically for ISub
type ISubStore interface {
	IStore
	MustAdd(sub ISub) ISub
	Add(sub ISub) (storedSub ISub, err error)
	Get(ParentID ID) (storedSub ISub)
	GetBy(key map[string]interface{}, limit int) []ISub
	Upd(sub ISub) error
	Del(ID) error
}
