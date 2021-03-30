package model

import (
	"fmt"
	"reflect"

	"github.com/go-msvc/msf/logger"
)

type IItemModel interface {
	IModel
	MustAdd(IItem) IItem
	Add(IItem) (IItem, error)
	Get(ID) IItem
	GetBy(key map[string]interface{}, limit int) []IItem
	Upd(IItem) error
	Del(ID) error
}

var log = logger.New("model").WithLevel(logger.LevelError)

var (
	itemInterfaceType = reflect.TypeOf(new(IItem)).Elem()
	subInterfaceType  = reflect.TypeOf(new(ISub)).Elem()
)

func MustNew(config IConfig, item IItem, refModels ...IItemModel) IItemModel {
	m, err := New(config, item, refModels...)
	if err != nil {
		panic(err)
	}
	return m
}

func New(config IConfig, item IItem, refModels ...IItemModel) (IItemModel, error) {
	itemType := reflect.TypeOf(item)
	if itemType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%v is not a struct", itemType)
	}
	if itemType.NumField() < 1 || !itemType.Field(0).Anonymous || itemType.Field(0).Type != reflect.TypeOf(Item{}) {
		return nil, fmt.Errorf("%v is not a struct starting with anonymous field model.Item", itemType)
	}
	return newModel(nil, config, itemType, refModels...)
}

//before calling this, check that item is struct starting with anonymous Item or Sub field.
func newModel(parent *itemModel, config IConfig, itemType reflect.Type, refModels ...IItemModel) (*itemModel, error) {
	m := &itemModel{
		config:    config,
		parent:    parent,
		itemType:  itemType,
		bareType:  reflect.StructOf(nil),
		values:    []ValueInfo{},
		refs:      []RefInfo{},
		subs:      []SubInfo{},
		bareStore: nil,
	}

	bareStructFields := []reflect.StructField{}
	for i := 0; i < m.itemType.NumField(); i++ {
		f := m.itemType.Field(i)
		switch f.Type.Kind() {
		case reflect.String, reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			m.values = append(m.values, ValueInfo{
				Name:           f.Name,
				Type:           f.Type,
				ItemFieldIndex: []int{i},
				BareFieldIndex: []int{len(bareStructFields)},
			})
			bareStructFields = append(bareStructFields, reflect.StructField{
				Name: f.Name,
				Type: f.Type,
			})
		case reflect.Struct:
			if m.parent == nil && i == 0 && f.Anonymous && f.Type == reflect.TypeOf(Item{}) {
				//this is the embedded Item field for ID, DateCreated and DateModified
				bareStructFields = append(bareStructFields, reflect.StructField{
					Anonymous: true,
					Name:      "Item",
					Type:      f.Type,
				})
				continue
			}
			if m.parent != nil && i == 0 && f.Anonymous && f.Type == reflect.TypeOf(Sub{}) {
				//this is the embedded Sub field for ParentID, DateCreated and DateModified
				bareStructFields = append(bareStructFields, reflect.StructField{
					Anonymous: true,
					Name:      "Sub",
					Type:      f.Type,
				})
				continue
			}
			if i > 0 && f.Type.Implements(itemInterfaceType) && f.Type != reflect.TypeOf(Item{}) {
				//this is a reference to an item in another item model which must already exist
				found := false
				var refModel IItemModel
				for _, refModel = range refModels {
					if refModel.Type() == f.Type {
						found = true
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("missing model for %v.%s = %v", m.itemType, f.Name, f.Type)
				}
				//describe the reference
				m.refs = append(m.refs, RefInfo{
					Name:           f.Name,
					Type:           f.Type,
					ItemFieldIndex: []int{i},
					BareFieldIndex: []int{len(bareStructFields)},
					RefModel:       refModel.WithReferenceFrom(m).(IItemModel),
				})
				//reference fields only stores the ID in the bare struct
				bareStructFields = append(bareStructFields, reflect.StructField{
					Anonymous: true,
					Name:      f.Name,
					Type:      reflect.TypeOf(ID(0)),
				})
				continue
			}
			return nil, fmt.Errorf("%v.field[%d]=%s type %v (anonymous=%v) is a %v which is not yet supported.", m.itemType, i, f.Name, f.Type, f.Anonymous, f.Type.Kind())

		case reflect.Slice:
			//this should be list of subs inside the item
			subType := f.Type.Elem()
			if !subType.Implements(subInterfaceType) {
				return nil, fmt.Errorf("%v.%s of type []%v does not implement model.ISub", m.itemType, f.Name, f.Type.Elem())
			}
			//confirm first field in sub is an anonymous model.Sub
			if subType.NumField() < 1 || !subType.Field(0).Anonymous || subType.Field(0).Type != reflect.TypeOf(Sub{}) {
				return nil, fmt.Errorf("%v.%s type []%v does not have anonymous model.Sub as first field", m.itemType, f.Name, f.Type.Elem())
			}
			//create a model to store the sub items
			subModel, err := newModel(m, m.config, subType, refModels...)
			if err != nil {
				return nil, fmt.Errorf("%v.%s type []%v failed to create sub model: %v", m.itemType, f.Name, subType, err)
			}

			//sub types are not added to the bare struct, they must be read from a sub-store
			//describe the sub
			m.subs = append(m.subs, SubInfo{
				Name:           f.Name,
				Type:           subType,
				ItemFieldIndex: []int{i},
				SubModel:       subModel,
			})
		default:
			return nil, fmt.Errorf("%v.%s is a %v which is not yet supported.", m.itemType, f.Name, f.Type.Kind())
		}
	}
	m.bareType = reflect.StructOf(bareStructFields)

	var err error
	if parent == nil {
		m.bareStore, err = m.config.ItemStore(reflect.New(m.bareType).Elem().Interface())
		if err != nil {
			return nil, fmt.Errorf("failed to create item store: %v", err)
		}
	} else {
		m.bareStore, err = m.config.SubStore(reflect.New(m.bareType).Elem().Interface())
		if err != nil {
			return nil, fmt.Errorf("failed to create sub store: %v", err)
		}
	}
	log.Debugf("model item=%+v bare=%+v", m.itemType, m.bareType)
	return m, nil
}

type itemModel struct {
	config    IConfig
	parent    *itemModel //only defined for sub models
	itemType  reflect.Type
	bareType  reflect.Type
	values    []ValueInfo
	refs      []RefInfo
	subs      []SubInfo
	bareStore IStore
	refFrom   []IModel
}

func (m itemModel) String() string {
	return fmt.Sprintf("ItemModel(%v)", m.itemType)
}

func (m *itemModel) WithReferenceFrom(user IModel) IModel {
	for _, u := range m.refFrom {
		if u == user {
			return m //already in the list
		}
	}
	m.refFrom = append(m.refFrom, user)
	return m
}

func (m itemModel) Type() reflect.Type { return m.itemType }

func (m itemModel) MustAdd(item IItem) IItem {
	//log.Debugf("model(%v).MustAdd(%T:%+v)", m.itemType, item, item)
	itemWithID, err := m.Add(item)
	if err != nil {
		panic(fmt.Errorf("failed to add: %v", err))
	}
	return itemWithID
}

//Add an item
func (m itemModel) Add(item IItem) (IItem, error) {
	//log.Debugf("model(%v).Add(%T:%+v)", m.itemType, item, item)
	if m.parent != nil {
		return nil, fmt.Errorf("cannot add item to sub model")
	}
	newItem, err := m.add(ID(0), item)
	if err != nil {
		return nil, fmt.Errorf("failed to add item: %v", err)
	}
	return newItem.(IItem), nil
}

//add new entry into model store
//param:
//	parentID - required for sub, must be omitted for item
//	item     - IItem or ISub, depending on the type of model
func (m itemModel) add(parentID ID, item interface{}) (interface{}, error) {
	log.Debugf("model(%v).add(%T:%+v)", m.itemType, item, item)
	if reflect.TypeOf(item) != m.itemType {
		return nil, fmt.Errorf("cannot add (%T), expecting %v", item, m.itemType)
	}

	bareItemPtrValue, err := m.bareFromItem(item)
	if err != nil {
		return nil, fmt.Errorf("failed to define bare item: %v", err)
	}

	//make copy of item so we can set the id assigned by the store
	itemWithIDPtrValue := reflect.New(m.itemType)
	itemWithIDPtrValue.Elem().Set(reflect.ValueOf(item))

	//add to the store
	var id ID
	if m.parent == nil {
		if parentID > 0 {
			return nil, fmt.Errorf("parentID specified for item")
		}
		id, err = m.bareStore.(IItemStore).Add(bareItemPtrValue.Elem().Interface().(IItem))
		if err != nil {
			return nil, fmt.Errorf("failed to store item: %v", err)
		}
		itemWithIDPtrValue.Elem().FieldByIndex([]int{0, 0}).Set(reflect.ValueOf(id))
	} else {
		if parentID <= 0 {
			return nil, fmt.Errorf("parentID not specified for sub")
		}
		bareItemPtrValue.Elem().FieldByIndex([]int{0, 0}).Set(reflect.ValueOf(parentID))
		//itemWithIDPtrValue.Elem().FieldByIndex([]int{0, 0}).Set(reflect.ValueOf(parentID))
		err := m.bareStore.(ISubStore).Add(bareItemPtrValue.Elem().Interface().(ISub))
		if err != nil {
			return nil, fmt.Errorf("failed to store sub: %v", err)
		}
		id = parentID
	}

	//store sub-items separately using the main item's id
	if m.parent != nil && len(m.subs) > 0 {
		return nil, fmt.Errorf("not yet storing sub-sub items... need complex id")
	}
	for _, sub := range m.subs {
		sliceValue := reflect.ValueOf(item).FieldByIndex(sub.ItemFieldIndex)
		nrSubs := sliceValue.Len()
		log.Debugf("Storing %d of %v ...", nrSubs, sub.Name)
		for i := 0; i < nrSubs; i++ {
			subValue := sliceValue.Index(i)
			log.Debugf("Saving %+v", subValue)
			if _, err := sub.SubModel.add(id, subValue.Interface().(ISub)); err != nil {
				return nil, fmt.Errorf("failed to add %v.%s[%d]: %v", m.Type(), sub.Name, i, err)
			}
		}
		// refID := fieldValue.FieldByIndex([]int{0, 0}).Interface().(ID) //-> model.Item{ItemID int}
		// refItem := r.model.Get(refID)
	}

	return itemWithIDPtrValue.Elem().Interface().(IItem), nil
} //itemModel.Add()

func (m itemModel) Upd(item IItem) error {
	if reflect.TypeOf(item) != m.itemType {
		return fmt.Errorf("cannot upd %T, expecting %v", item, m.itemType)
	}

	bareItemPtrValue, err := m.bareFromItem(item)
	if err != nil {
		return fmt.Errorf("failed to define bare item: %v", err)
	}
	if err = m.bareStore.(IItemStore).Upd(bareItemPtrValue.Elem().Interface()); err != nil {
		return fmt.Errorf("failed to update item: %v", err)
	}

	//log.Debugf("New Item With ID: (%T)%+v", bareItemPtrValue.Elem().Interface(), bareItemPtrValue.Elem().Interface())
	//and check that the reference items already exist
	//note: item ID is set after adding the the store because the store assigns the ID

	// //store sub-items separately using the main item's id
	// for _, s := range m.subs {
	// 	sliceValue := v.Field(s.fieldIndex)
	// 	nrSubs := sliceValue.Len()
	// 	log.Debugf("Storing %d subs %v...", nrSubs, s.fieldName)
	// 	for i := 0; i < nrSubs; i++ {
	// 		subValue := sliceValue.Index(i)
	// 		if _, err := s.model.Add(subValue.Interface().(ISub)); err != nil {
	// 			return nil, fmt.Errorf("failed to add %v.%s[%d]: %v", m.Type(), s.fieldName, i, err)
	// 		}
	// 	}

	// 	// refID := fieldValue.FieldByIndex([]int{0, 0}).Interface().(ID) //-> model.Item{ItemID int}
	// 	// refItem := r.model.Get(refID)

	// }

	return nil
} //itemMode.Upd()

func (m itemModel) Get(id ID) IItem {
	if m.parent != nil {
		panic(fmt.Errorf("model(%v) is sub, replace Get() with GetSubs()", m.itemType))
	}
	bareItem := m.bareStore.(IItemStore).Get(id)
	if bareItem == nil {
		return nil
	}

	//got item from store, read referenced items by ID and update in this struct
	return m.itemFromBareItem(bareItem)
}

func (m itemModel) GetSubs(parentID ID) []ISub {
	if m.parent == nil {
		panic(fmt.Errorf("model(%v) is item, replace GetSubs() with Get()", m.itemType))
	}
	bareItems := m.bareStore.(ISubStore).Get(parentID)
	if bareItems == nil {
		return nil
	}

	subs := []ISub{}
	for _, bareItem := range bareItems {
		subs = append(subs, m.itemFromBareItem(bareItem).(ISub))
	}
	//got item from store, read referenced items by ID and update in this struct
	return subs
}

func (m itemModel) GetBy(key map[string]interface{}, limit int) []IItem {
	bareItems := m.bareStore.(IItemStore).GetBy(key, limit)
	if len(bareItems) == 0 {
		return nil
	}

	//got item(s) from store, read referenced items by ID and update in this struct
	returnItems := []IItem{}
	for _, bareItem := range bareItems {
		returnItems = append(returnItems, m.itemFromBareItem(bareItem))
	}
	return returnItems
}

func (m *itemModel) Del(id ID) error {
	log.Debugf("%T(%v).Del(%v) refFrom=%+v...", m, m.itemType, id, m.refFrom)
	//before delete - check that this item is not used by other models
	for _, u := range m.refFrom {
		if userModel, userId, used := u.HasReferenceTo(m, id); used {
			return fmt.Errorf("cannot delete %v.id=%d because used by %s.%v.id=%d",
				m.itemType,
				id,
				userModel.Type().Name(),
				u.Type(),
				userId)
		}
	}
	return m.bareStore.(IItemStore).Del(id)
}

func (m *itemModel) HasReferenceTo(refModel IItemModel, refID ID) (IItemModel, ID, bool) {
	log.Debugf("%s.HasReference(%v,%v)...", m, refModel.Type(), refID)
	//see if any item in this model has a reference to the specified refID
	for _, refInfo := range m.refs {
		if refInfo.RefModel == refModel {
			//e.g. get membership where membership.person.id == id
			bareItems := m.bareStore.(IItemStore).GetBy(map[string]interface{}{refInfo.Name: refID}, 1)
			if len(bareItems) > 0 {
				log.Debugf("%s.HasReference(%v,%v)->true", m, refModel.Type(), refID)
				return m, ItemID(bareItems[0]), true
			}
		}
	}
	log.Debugf("%s.HasReference(%v,%v)->false", m, refModel.Type(), refID)
	return nil, 0, false
}

func (m itemModel) Count() int {
	return m.bareStore.Count()
}

type ValueInfo struct {
	Name           string
	Type           reflect.Type
	ItemFieldIndex []int
	BareFieldIndex []int
}

type RefInfo struct {
	Name           string
	Type           reflect.Type
	ItemFieldIndex []int
	BareFieldIndex []int
	RefModel       IItemModel
}

type SubInfo struct {
	Name           string
	Type           reflect.Type
	ItemFieldIndex []int
	SubModel       *itemModel
	//not present in bareType: BareFieldIndex []int
}

func (m itemModel) itemFromBareItem(bareItem IItem) IItem {
	id := ItemID(bareItem)
	bareItemValue := reflect.ValueOf(bareItem)
	//got bareItem from store: copy ID and values into item
	newItemPtrValue := reflect.New(m.itemType)
	newItemPtrValue.Elem().FieldByIndex([]int{0, 0}).Set(reflect.ValueOf(id))
	for _, valueInfo := range m.values {
		newItemPtrValue.Elem().FieldByIndex(valueInfo.BareFieldIndex).Set(bareItemValue.FieldByIndex(valueInfo.BareFieldIndex))
	}
	//read referenced items by ID and update in the item struct
	for _, refInfo := range m.refs {
		refID := bareItemValue.FieldByIndex(refInfo.BareFieldIndex).Interface().(ID)
		refItem := refInfo.RefModel.Get(refID)
		newItemPtrValue.Elem().FieldByIndex(refInfo.ItemFieldIndex).Set(reflect.ValueOf(refItem))
	}
	//read sub-items: TODO
	for _, subInfo := range m.subs {
		parentSubs := subInfo.SubModel.GetSubs(id)
		n := len(parentSubs)
		arrayType := reflect.ArrayOf(n, subInfo.Type)
		newSlicePtrValue := reflect.New(arrayType)
		for i, parentSub := range parentSubs {
			newSlicePtrValue.Elem().Index(i).Set(reflect.ValueOf(parentSub))
		}
		newItemPtrValue.Elem().FieldByIndex(subInfo.ItemFieldIndex).Set(newSlicePtrValue.Elem().Slice(0, n))
	}
	return newItemPtrValue.Elem().Interface().(IItem)
}

func (m itemModel) bareFromItem(item IItem) (bareItemPtrValue reflect.Value, err error) {
	//create new bareType struct
	bareItemPtrValue = reflect.New(m.bareType)
	//copy id
	bareItemPtrValue.Elem().FieldByIndex([]int{0, 0}).Set(reflect.ValueOf(ItemID(item)))
	//copy value fields from item to bare item
	itemValue := reflect.ValueOf(item)
	for _, valueInfo := range m.values {
		bareItemPtrValue.Elem().FieldByIndex(valueInfo.BareFieldIndex).Set(itemValue.FieldByIndex(valueInfo.ItemFieldIndex))
	}
	//copy ids of reference fields and check that they exist
	for _, refInfo := range m.refs {
		//refValue is the complete referenced item struct inside the item
		refValue := itemValue.FieldByIndex(refInfo.ItemFieldIndex)
		//get the Item.ID inside the refValue
		refID := refValue.FieldByIndex([]int{0, 0}).Interface().(ID)
		//read the current value of this refItem from its model
		refItem := refInfo.RefModel.Get(refID)
		if refItem == nil {
			err = fmt.Errorf("%v.%s.id=%v not found", m.itemType, refInfo.Name, refID)
			return
		}
		//compare specified value with what is in the store
		//todo: provide option to override this check and only check the ID exists
		//which is useful when one expect concurrent updates to the ref item that
		//are not significant to the caller of this function...
		if refItem != refValue.Interface() {
			err = fmt.Errorf("%v.%s.id=%v: the referenced value inside item: %+v, does not reflect what is stored for that id: %+v", m.itemType, refInfo.Name, refID, refValue.Interface(), refItem)
			return
		}
		bareItemPtrValue.Elem().FieldByIndex(refInfo.BareFieldIndex).Set(refValue.FieldByIndex([]int{0, 0}))
	}
	err = nil
	return
}

func ItemID(item IItem) ID {
	//log.Debugf("ItemID(%T:%+v)...", item, item)
	return reflect.ValueOf(item).FieldByIndex([]int{0, 0}).Interface().(ID)
}
