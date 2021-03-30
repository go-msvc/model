package model_test

import (
	"testing"

	"github.com/go-msvc/model"
	"github.com/go-msvc/model/store/memory"
)

type Person struct {
	model.Item
	Name string
}

func (p Person) WithID(id model.ID) model.IItem { p.ItemID = id; return p }

func TestPerson(t *testing.T) {
	//basic same as store/memory_test - just wrapped by model...
	//with a simple item struct
	//personStore := memory.New(Person{})
	personModel, _ := model.New(memory.Config{}, Person{})
	//one can add simple items
	pA := personModel.MustAdd(Person{Name: "A"}).(Person)
	pB := personModel.MustAdd(Person{Name: "B"}).(Person)
	pC := personModel.MustAdd(Person{Name: "C"}).(Person)
	//then expect them to exist with count
	if personModel.Count() != 3 {
		t.Fatalf("not 3")
	}
	//and retriev them
	pAA := personModel.Get(pA.ItemID).(Person)
	if pAA.ItemID != pA.ItemID || pAA.Name != "A" {
		t.Fatalf("incorrect name")
	}
	pBB := personModel.Get(pB.ItemID).(Person)
	if pBB.ItemID != pB.ItemID || pBB.Name != "B" {
		t.Fatalf("incorrect name")
	}
	pCC := personModel.Get(pC.ItemID).(Person)
	if pCC.ItemID != pC.ItemID || pCC.Name != "C" {
		t.Fatalf("incorrect name")
	}
	//and update them
	pCC.Name = "DDD"
	if err := personModel.Upd(pCC); err != nil {
		t.Fatalf("upd failed: %v", err)
	}
	//and delete some
	personModel.Del(pA.ItemID)
	if personModel.Count() != 2 {
		t.Fatalf("after del not 2 left")
	}
	pCCC := personModel.Get(pC.ItemID).(Person)
	if pCCC.Name != "DDD" {
		t.Fatalf("after update name not correct: %+v", pCCC)
	}
}

type Membership struct {
	model.Item
	Person
	Number int
	Active bool
}

func (m Membership) WithID(id model.ID) model.IItem { m.ItemID = id; return m }

func TestReference(t *testing.T) {
	//when one type refer to another, e.g. Membership refers to a person
	//the store can access items without respecting relationships
	//the model respects relationships - so use model when there are relationships
	personModel := model.MustNew(memory.Config{}, Person{})
	membershipModel := model.MustNew(memory.Config{}, Membership{}, personModel)

	//with some persons...
	pA := personModel.MustAdd(Person{Name: "A"}).(Person)
	pB := personModel.MustAdd(Person{Name: "B"}).(Person)
	pC := personModel.MustAdd(Person{Name: "C"}).(Person)
	t.Logf("Added persons %v %v %v", pA, pB, pC)

	//one can create a membership on person A:
	mA := membershipModel.MustAdd(Membership{Person: pA, Number: 111, Active: true}).(Membership)

	//one cannot create a membership on a person that does not exist
	if _, err := membershipModel.Add(Membership{Person: Person{Name: "D"}, Number: 111, Active: true}); err == nil {
		t.Fatalf("created membership with non-existing person")
	}

	//the membership store does not store the whole person, just the ID!
	//the store/memory use the struct with everything in it, but it can only rely on the ID
	//so if we now update the person A to have a different name, the change
	//only reflects in the person store, while the membership store still has the
	//correct id but the old name. So when we retrieve the membership through the
	//model, it must take the person id and read the rest of the data from the person store
	//the fact that it stores the whole struct is just for simplicity
	//  the alternative is to create a new reflect struct, replacing the person field with an ID-only field.
	pA.Name = "AAA"
	personModel.Upd(pA)

	//now confirm old name still in membership store
	// mA = membershipStore.Get(mA.ItemID).(Membership)
	// t.Logf("membership store -> (%v=%v,%v,%v)", mA.Person.ItemID, mA.Person.Name, mA.Number, mA.Name)
	// if mA.Person.ItemID != pA.ItemID || mA.Person.Name != "A" {
	// 	t.Fatalf("membership store was updated unexpectedly: %+v", mA)
	// }

	//but read through the model to get the updated person data
	mA = membershipModel.Get(mA.ItemID).(Membership)
	t.Logf("membership model -> (%v=%v,%v,%v)", mA.Person.ItemID, mA.Person.Name, mA.Number, mA.Name)
	if mA.Person.ItemID != pA.ItemID || mA.Person.Name != "AAA" {
		t.Fatalf("membership model was NOT updated: %+v", mA)
	}

	//cannot delete person A while referenced from membership
	if err := personModel.Del(pA.ItemID); err == nil {
		t.Fatalf("deleted person A while in use by membership")
	}

	//after deleting membership
	if err := membershipModel.Del(mA.ItemID); err != nil {
		t.Fatalf("cannot delete membership")
	}

	//we can also delete the person
	if err := personModel.Del(pA.ItemID); err != nil {
		t.Fatalf("cannot delete person A after membership was deleted")
	}
}

type Car struct {
	model.Item
	Make   string
	Model  string
	Wheels []Wheel
}

func (c Car) WithID(id model.ID) model.IItem {
	c.Item.ItemID = id
	return c
}

type Wheel struct {
	model.Sub
	Name string
}

func TestSubs(t *testing.T) {
	carModel, err := model.New(memory.Config{}, Car{})
	if err != nil {
		t.Fatalf("failed to create model: %v", err)
	}

	//should not be possible to add a person to a car store
	if _, err := carModel.Add(Person{Name: "A"}); err == nil {
		t.Fatalf("successfully added person to car store!!!")
	}

	//add cat without subs
	cA := carModel.MustAdd(Car{Make: "A", Model: "1"}).(Car)
	if carModel.Count() != 1 {
		t.Fatalf("add -> count != 1")
	}
	t.Logf("idA = %v", cA.ItemID)

	cAA := carModel.Get(cA.ItemID).(Car)
	t.Logf("got cA back: %+v", cAA)
	if cAA.Make != cA.Make || cAA.Model != cA.Model {
		t.Fatalf("got back not same: %+v != %+v", cA, cAA)
	}

	//add car with wheels (subs)
	cB := carModel.MustAdd(Car{Make: "B", Model: "2", Wheels: []Wheel{{Name: "w1"}, {Name: "w2"}}}).(Car)
	t.Logf("Added cB: %+v", cB)
	cBB := carModel.Get(cB.ItemID)
	if cBB == nil {
		t.Fatalf("Did not get car back")
	}
	cBBB := cBB.(Car)
	t.Logf("got cB back: %+v", cBBB)
	if cBBB.Make != cB.Make || cBBB.Model != cB.Model {
		t.Fatalf("got back not same: %+v != %+v", cB, cBB)
	}
	if len(cBBB.Wheels) != 2 {
		t.Fatalf("retrieved cB has %d instead of 2 wheels", len(cBBB.Wheels))
	}
}
