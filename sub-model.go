package model

// func MustNewSub(store ISubStore, referenceModels ...IItemModel) ISubModel {
// 	m, err := NewSub(store, referenceModels...)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return m
// }

// func NewSub(store ISubStore, referenceModels ...IItemModel) (ISubModel, error) {
// 	// m := &subModel{
// 	// 	store:      store,
// 	// 	references: []reference{},
// 	// 	subs:       []sub{},
// 	// }
// 	// //look for references to other models
// 	// structType := store.Type()
// 	// for i := 0; i < structType.NumField(); i++ {
// 	// 	f := structType.Field(i)
// 	// 	if f.Type.Implements(itemInterfaceType) {
// 	// 		//this is a reference to an item in another model which must already exist
// 	// 		found := false
// 	// 		for _, referenceModel := range referenceModels {
// 	// 			if referenceModel.Type() == f.Type {
// 	// 				m.references = append(m.references, reference{
// 	// 					fieldIndex: i,
// 	// 					fieldName:  f.Name,
// 	// 					model:      referenceModel.UsedBy(m).(IItemModel),
// 	// 				})
// 	// 				found = true
// 	// 				break
// 	// 			}
// 	// 		}
// 	// 		if !found {
// 	// 			return nil, fmt.Errorf("missing model for %v.%s = %v", store.Type(), f.Name, f.Type)
// 	// 		}
// 	// 	}

// 	// 	if f.Type.Kind() == reflect.Slice {
// 	// 		//lists must be stored as sub-items
// 	// 		if !f.Type.Elem().Implements(subInterfaceType) {
// 	// 			return nil, fmt.Errorf("%v.%s type %v does not implement model.Sub", store.Type(), f.Name, f.Type.Elem())
// 	// 		}
// 	// 		//this is a sub item stored in a list
// 	// 		//todo: get min/max limits
// 	// 		found := false
// 	// 		for _, subModel := range referenceModels {
// 	// 			if subModel.Type() == f.Type.Elem() {
// 	// 				m.subs = append(m.subs, sub{
// 	// 					fieldIndex: i,
// 	// 					fieldName:  f.Name,
// 	// 					model:      subModel.UsedBy(m).(ISubModel),
// 	// 				})
// 	// 				found = true
// 	// 				break
// 	// 			}
// 	// 		}
// 	// 		if !found {
// 	// 			return nil, fmt.Errorf("missing model for %v.%s = %v", store.Type(), f.Name, f.Type)
// 	// 		}
// 	// 		log.Debugf("%v.%f is sub", store.Type(), f.Name)
// 	// 	}
// 	// }
// 	// return m, nil
// 	return nil, fmt.Errorf("NYI")
// }

// type ISubModel interface {
// 	//IModel
// 	//ISubStore
// }

// type subModel struct {
// 	parentItemModel IItemModel
// 	store           ISubStore
// 	// references      []reference
// 	// subs            []sub
// 	usedBy []IModel
// }

// func (m *subModel) UsedBy(user IModel) IModel {
// 	for _, u := range m.usedBy {
// 		if u == user {
// 			return m //already in the list
// 		}
// 	}
// 	m.usedBy = append(m.usedBy, user)
// 	return m
// }

// func (m subModel) Type() reflect.Type { return m.store.Type() }

// func (m subModel) MustAdd(item ISub) ISub {
// 	itemWithID, err := m.Add(item)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return itemWithID
// }

// func (m subModel) Add(item ISub) (ISub, error) {
// 	if reflect.TypeOf(item) != m.store.Type() {
// 		return nil, fmt.Errorf("cannot add %T, expecting %v", item, m.store.Type())
// 	}

// 	//check referenced items must exist as specified
// 	v := reflect.ValueOf(item)
// 	for _, r := range m.references {
// 		fieldValue := v.Field(r.fieldIndex)
// 		refID := fieldValue.FieldByIndex([]int{0, 0}).Interface().(ID) //-> model.Item{ItemID int}
// 		refItem := r.model.Get(refID)
// 		if refItem == nil {
// 			return nil, fmt.Errorf("%v.%s.id=%v not found", m.store.Type(), r.fieldName, refID)
// 		}
// 		if refItem != fieldValue.Interface() {
// 			return nil, fmt.Errorf("%v.%s.id=%v=(%+v) != stored(%+v)", m.store.Type(), r.fieldName, refID, fieldValue.Interface(), refItem)
// 		}
// 	}

// 	//store the main item
// 	storedItem, err := m.store.Add(item)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to store item: %v", err)
// 	}

// 	//store sub-items separately using the main item's id
// 	for _, s := range m.subs {
// 		sliceValue := v.Field(s.fieldIndex)
// 		nrSubs := sliceValue.Len()
// 		log.Debugf("Storing %d subs %v...", nrSubs, s.fieldName)
// 		for i := 0; i < nrSubs; i++ {
// 			subValue := sliceValue.Index(i)
// 			if _, err := s.model.Add(subValue.Interface().(ISub)); err != nil {
// 				return nil, fmt.Errorf("failed to add %v.%s[%d]: %v", m.Type(), s.fieldName, i, err)
// 			}
// 		}

// 		// refID := fieldValue.FieldByIndex([]int{0, 0}).Interface().(ID) //-> model.Item{ItemID int}
// 		// refItem := r.model.Get(refID)

// 	}

// 	return storedItem, nil
// }

// func (m subModel) Get(id ID) ISub {
// 	item := m.store.Get(id)
// 	if item == nil {
// 		return nil
// 	}

// 	//got item from store, read referenced items by ID and update in this struct
// 	readItemValue := reflect.ValueOf(item)
// 	newItemPtrValue := reflect.New(m.store.Type())
// 	newItemPtrValue.Elem().Set(readItemValue)
// 	for _, r := range m.references {
// 		fieldValue := readItemValue.Field(r.fieldIndex)
// 		refID := fieldValue.FieldByIndex([]int{0, 0}).Interface().(ID) //-> model.Item{ItemID int}
// 		refItem := r.model.Get(refID)
// 		newItemPtrValue.Elem().Field(r.fieldIndex).Set(reflect.ValueOf(refItem))
// 	}
// 	return newItemPtrValue.Elem().Interface().(ISub)
// }

// func (m subModel) GetBy(key map[string]interface{}, limit int) []ISub {
// 	items := m.store.GetBy(key, limit)
// 	if len(items) == 0 {
// 		return nil
// 	}

// 	//got item(s) from store, read referenced items by ID and update in this struct
// 	returnItems := []ISub{}
// 	for _, item := range items {
// 		readItemValue := reflect.ValueOf(item)
// 		newItemPtrValue := reflect.New(m.store.Type())
// 		newItemPtrValue.Elem().Set(readItemValue)
// 		for _, r := range m.references {
// 			fieldValue := readItemValue.Field(r.fieldIndex)
// 			refID := fieldValue.FieldByIndex([]int{0, 0}).Interface().(ID) //-> model.Item{ItemID int}
// 			refItem := r.model.Get(refID)
// 			newItemPtrValue.Elem().Field(r.fieldIndex).Set(reflect.ValueOf(refItem))
// 		}
// 		returnItems = append(returnItems, newItemPtrValue.Elem().Interface().(ISub))
// 	}
// 	return returnItems
// }

// func (m subModel) Upd(item ISub) error {
// 	if reflect.TypeOf(item) != m.store.Type() {
// 		return fmt.Errorf("cannot upd %T, expecting %v", item, m.store.Type())
// 	}

// 	//check referenced items must exist as specified
// 	v := reflect.ValueOf(item)
// 	for _, r := range m.references {
// 		fieldValue := v.Field(r.fieldIndex)
// 		refID := fieldValue.FieldByIndex([]int{0, 0}).Interface().(ID) //-> model.Item{ItemID int}
// 		refItem := r.model.Get(refID)
// 		if refItem == nil {
// 			return fmt.Errorf("%v.%s.id=%v not found", m.store.Type(), r.fieldName, refID)
// 		}
// 		if refItem != fieldValue.Interface() {
// 			return fmt.Errorf("%v.%s.id=%v=(%+v) != stored(%+v)", m.store.Type(), r.fieldName, refID, fieldValue.Interface(), refItem)
// 		}
// 	}
// 	return m.store.Upd(item)
// }

// func (m *subModel) Del(id ID) error {
// 	// //before delete - check that this item is not used by other models
// 	// for _, u := range m.usedBy {
// 	// 	if userId, used := u.HasReferenceTo(m, id); used {
// 	// 		return fmt.Errorf("cannot delete %v.id=%d because used by %v.id=%d",
// 	// 			m.store.Type(),
// 	// 			id,
// 	// 			u.Type(),
// 	// 			userId)
// 	// 	}
// 	// }
// 	return m.store.Del(id)
// }

// func (m subModel) HasReferenceTo(usedModel IItemModel, id ID) (IItemModel, ID, bool) {
// 	//see if any sub has a reference to specified item id
// 	for _, r := range m.references {
// 		//e.g. get membership where membership.person.id == id
// 		//as far as the store is concerned, field "person" is a reference and it should only store the id
// 		//in the case of store/memory, it stores the whole struct, but when key on person, it must realise
// 		//it is a IItem reference and compare only the item id...
// 		items := m.store.GetBy(map[string]interface{}{r.fieldName: id}, 1)
// 		if len(items) > 0 {
// 			//return the parent model and item id that owns this sub, i.e. that has the reference
// 			return m.parentItemModel, items[0].ParentID(), true
// 		}
// 	}
// 	return nil, 0, false
// }

// func (m subModel) Count() int {
// 	return 0 //m.store.Count()
// }
