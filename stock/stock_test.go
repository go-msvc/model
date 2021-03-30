package stock_test

import (
	"testing"

	"github.com/go-msvc/model/stock"
)

func TestModel(t *testing.T) {
	m := stock.New()

	status1 := m.StockStatus.MustAdd(stock.StockStatus{Status: "S1", Description: "Status1"}).(stock.StockStatus)
	status2 := m.StockStatus.MustAdd(stock.StockStatus{Status: "S2", Description: "Status2"}).(stock.StockStatus)

	talCompany := m.Company.MustAdd(stock.Company{Name: "Takealot", Info: "Takealot is hot"}).(stock.Company)
	t.Logf("talCompany=%+v", talCompany)

	locJhb := m.Location.MustAdd(stock.Location{Name: "Jhb", Info: "Johannesburg DC"}).(stock.Location)
	t.Logf("locJhb=%+v", locJhb)

	locCpt := m.Location.MustAdd(stock.Location{Name: "Cpt", Info: "Cape Town DC"}).(stock.Location)
	t.Logf("locCpt=%+v", locCpt)

	//todo: validate to ensure all required fields are specified
	talOwner := m.Owner.MustAdd(stock.Owner{Company: talCompany, MerchantReference: "Takealot"}).(stock.Owner)
	t.Logf("owner=%+v", talOwner)

	//create inventories for a few product references owned by tal
	inv1 := m.Inventory.MustAdd(stock.Inventory{Owner: talOwner, ProductReference: "11111"}).(stock.Inventory)
	t.Logf("inventory=%+v", inv1)
	inv2 := m.Inventory.MustAdd(stock.Inventory{Owner: talOwner, ProductReference: "22222"}).(stock.Inventory)
	t.Logf("inventory=%+v", inv2)
	inv3 := m.Inventory.MustAdd(stock.Inventory{Owner: talOwner, ProductReference: "33333"}).(stock.Inventory)
	t.Logf("inventory=%+v", inv3)

	//create stock at both locations for the products
	s1 := m.Stock.MustAdd(stock.Stock{
		Inventory: inv1,
		Location:  locJhb,
		Levels: []stock.StockLevel{
			{StockStatus: status1, Quantity: 5},
			{StockStatus: status2, Quantity: 9},
		},
	}).(stock.Stock)
	t.Logf("s1=%+v", s1)
	s2 := m.Stock.MustAdd(stock.Stock{Inventory: inv1, Location: locCpt}).(stock.Stock)
	t.Logf("s2=%+v", s2)

	//todo... set levels in sub-items
	//todo... set levels in sub-items

	if m.Stock.Count() != 2 {
		t.Fatalf("%d instead of 2 stock items", m.Stock.Count())
	}

	s11 := m.Stock.Get(s1.ID()).(stock.Stock)
	t.Logf("got s1: %+v", s11)
	if len(s11.Levels) != 2 {
		t.Fatalf("s1 retrieved with %d instead of 2 levels", len(s11.Levels))
	}

	s22 := m.Stock.Get(s2.ID())
	if s22 == nil {
		t.Fatalf("did not get s2 back")
	}
}
