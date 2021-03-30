package model

type IModel interface {
	IStore
	UsedBy(IModel) IModel
	HasReferenceTo(IItemModel, ID) (IItemModel, ID, bool)
}
