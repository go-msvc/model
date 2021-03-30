package model

type ISub interface {
	//ParentID() ID
}

type Sub struct {
	ParentItemID ID   `sql:"parent_id,INTEGER(11),PRIMARY_KEY"`
	DateCreated  Date `sql:"date_created,DATETIME"`
	DateModified Date `sql:"date_modified,DATETIME"`
}

func (i Sub) ParentID() ID { return i.ParentItemID }
