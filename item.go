package model

type IItem interface {
	//ID() ID
	//	WithID(ID) IItem //must be implemented in user struct
}

type Item struct {
	ItemID       ID   `sql:"id,INTEGER(11),PRIMARY_KEY"`
	DateCreated  Date `sql:"date_created,DATETIME"`
	DateModified Date `sql:"date_modified,DATETIME"`
}

func (i Item) ID() ID { return i.ItemID }
