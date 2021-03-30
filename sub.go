package model

type ISub interface {
	ParentID() ID
}

type Sub struct {
	ParentItemID ID `sql:"id,INTEGER(11),PRIMARY_KEY"`
}

func (i Sub) ParentID() ID { return i.ParentItemID }
