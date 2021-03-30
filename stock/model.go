package stock

import (
	"github.com/go-msvc/model"
	"github.com/go-msvc/model/store/memory"
	"github.com/go-msvc/msf/logger"
)

var log = logger.New("model").New("stock")

type Location struct {
	model.Item
	Name string `sql:"name,VARCHAR(64)"`
	Info string `sql:"info,VARCHAR(265)"`
}

func (l Location) WithID(id model.ID) model.IItem {
	l.ItemID = id
	return l
}

type Company struct {
	model.Item
	Name string `sql:"name,VARCHAR(64)"`
	Info string `sql:"info,VARCHAR(265)"`
}

func (c Company) WithID(id model.ID) model.IItem {
	c.ItemID = id
	return c
}

type Owner struct {
	model.Item
	Company           Company `sql:"company_id,INTEGER(11),ForeignKey(company.company_id)"`
	MerchantReference string  `sql:"merchant_reference,VARCHAR(64)"`
}

func (o Owner) WithID(id model.ID) model.IItem {
	o.ItemID = id
	return o
}

type StockStatus struct {
	model.Item
	Status      string `sql:"status,VARCHAR(64)"`
	Description string `sql:"description,VARCHAR(256)"`
}

func (ss StockStatus) WithID(id model.ID) model.IItem {
	ss.ItemID = id
	return ss
}

type Inventory struct {
	model.Item
	ProductReference string `sql:"product_reference"`
	Owner            Owner  `sql:"owner"`
}

func (i Inventory) WithID(id model.ID) model.IItem {
	i.ItemID = id
	return i
}

type Stock struct {
	model.Item
	Inventory Inventory
	Location  Location
	Levels    []StockLevel //sub-item
}

func (s Stock) WithID(id model.ID) model.IItem {
	s.ItemID = id
	return s
}

type StockLevel struct {
	model.Sub
	StockStatus
	Quantity int
}

type Model struct {
	Location    model.IItemModel
	Company     model.IItemModel
	Owner       model.IItemModel
	StockStatus model.IItemModel
	Inventory   model.IItemModel
	Stock       model.IItemModel
}

func New() Model {
	m := Model{}
	c := memory.Config{}
	m.Location = model.MustNew(c, Location{})
	m.Company = model.MustNew(c, Company{})
	m.Owner = model.MustNew(c, Owner{}, m.Company)
	m.StockStatus = model.MustNew(c, StockStatus{})
	m.Inventory = model.MustNew(c, Inventory{}, m.Owner)
	m.Stock = model.MustNew(c, Stock{}, m.Inventory, m.Location, m.StockStatus)

	// locJhb := m.Location.MustAdd(Location{Name: "Jhb"}).(Location)
	// locCpt := m.Location.MustAdd(Location{Name: "Cpt"}).(Location)
	// talCompany := m.Company.MustAdd(Company{Name: "Takealot"}).(Company)

	// talOwner := m.Owner.MustAdd(Owner{Company: talCompany, MerchantReference: "Takealot"}).(Owner)

	// i1 := m.Inventory.MustAdd(Inventory{Owner: talOwner, ProductReference: "1"}).(Inventory)
	// s1jhb := m.Stock.MustAdd(Stock{Inventory: i1, Location: locJhb})
	// s1cpt := m.Stock.MustAdd(Stock{Inventory: i1, Location: locCpt})
	// log.Debugf("s1: %v %v", s1jhb, s1cpt)

	return m
}

// func (m Model) GetLevelsByLocationId(productReference string, companyId model.ID, locationId model.ID) []interface{} {
// 	return m.client.GetSessionContext("stock.leader").
// 		Query(StockLevel{}, StockStatus{}).
// 		Join(Stock{}).
// 		Join(Inventory{}).
// 		Join(Owner{}).
// 		Join(StockStatus{}).
// 		Filter("Owner.CompanyId == companyId",
// 				"Inventory.ProductReference == productReference",
// 				"StockLevel.StockStatusId == StockStatus.stockStatusId",
// 				"Stock.LocationId == locationId").
// 		Merged(). //put stock level and stock_status fields in one item
// 		All()
// }
