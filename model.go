package model

type IModel interface {
	IStore
	WithReferenceFrom(IModel) IModel
	HasReferenceTo(IItemModel, ID) (IItemModel, ID, bool)
}
