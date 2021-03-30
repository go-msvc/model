package memory_test

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
	s := memory.New(Person{})
	pA := Person{Name: "A"}
	pA.ItemID = s.MustAdd(pA)
	if s.Count() != 1 {
		t.Fatalf("add -> count != 1")
	}
	t.Logf("idA = %v", pA.ItemID)

	pB := Person{Name: "B"}
	pB.ItemID = s.MustAdd(pB)
	if s.Count() != 2 {
		t.Fatalf("add -> count != 2")
	}
	t.Logf("idB = %v", pB.ItemID)
	if pA.ItemID == pB.ItemID {
		t.Fatalf("add A and B got same ID")
	}

	pC := Person{Name: "C"}
	pC.ItemID = s.MustAdd(pC)
	if s.Count() != 3 {
		t.Fatalf("add -> count != 3")
	}
	t.Logf("idC = %v", pC.ItemID)

	pAA := s.Get(pA.ItemID).(Person)
	if pAA.Name != "A" || pAA.ItemID != pA.ItemID {
		t.Fatalf("did not get A back: (%T)%+v", pAA, pAA)
	}
	t.Logf("got A back: (%T)%+v", pAA, pAA)

	pBB := s.Get(pB.ItemID).(Person)
	if pBB.Name != "B" || pBB.ItemID != pB.ItemID {
		t.Fatalf("did not get B back: (%T)%+v", pBB, pBB)
	}
	t.Logf("got B back: (%T)%+v", pBB, pBB)

	pCC := s.Get(pC.ItemID).(Person)
	if pCC.Name != "C" || pCC.ItemID != pC.ItemID {
		t.Fatalf("did not get C back: (%T)%+v", pCC, pCC)
	}
	t.Logf("got C back: (%T)%+v", pCC, pCC)

	pBB.Name = "BBB"
	if err := s.Upd(pBB); err != nil {
		t.Fatalf("failed to update: %v", err)
	}

	pBBB := s.Get(pBB.ItemID).(Person)
	if pBBB.Name != "BBB" {
		t.Fatalf("updated but not retrieved")
	}
	t.Logf("got updated (%T)%+v", pBBB, pBBB)
}
