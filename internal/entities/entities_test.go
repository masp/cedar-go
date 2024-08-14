package entities

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntities(t *testing.T) {
	t.Parallel()
	t.Run("Clone", func(t *testing.T) {
		t.Parallel()
		e := Entities{
			types.EntityUID{Type: "A", ID: "A"}: {},
			types.EntityUID{Type: "A", ID: "B"}: {},
			types.EntityUID{Type: "B", ID: "A"}: {},
			types.EntityUID{Type: "B", ID: "B"}: {},
		}
		clone := e.Clone()
		testutil.Equals(t, clone, e)
	})

}

func TestEntitiesJSON(t *testing.T) {
	t.Parallel()
	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()
		e := Entities{}
		ent := Entity{
			UID:        types.NewEntityUID("Type", "id"),
			Parents:    []types.EntityUID{},
			Attributes: types.Record{"key": types.Long(42)},
		}
		e[ent.UID] = ent
		b, err := e.MarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(b), `[{"uid":{"type":"Type","id":"id"},"attrs":{"key":42}}]`)
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()
		b := []byte(`[{"uid":{"type":"Type","id":"id"},"parents":[],"attrs":{"key":42}}]`)
		var e Entities
		err := json.Unmarshal(b, &e)
		testutil.OK(t, err)
		want := Entities{}
		ent := Entity{
			UID:        types.NewEntityUID("Type", "id"),
			Parents:    []types.EntityUID{},
			Attributes: types.Record{"key": types.Long(42)},
		}
		want[ent.UID] = ent
		testutil.Equals(t, e, want)
	})

	t.Run("UnmarshalErr", func(t *testing.T) {
		t.Parallel()
		var e Entities
		err := e.UnmarshalJSON([]byte(`!@#$`))
		testutil.Error(t, err)
	})
}
