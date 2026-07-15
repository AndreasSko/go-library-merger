package model

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"text/tabwriter"

	"github.com/stretchr/testify/assert"
)

func TestInputField_SetID(t *testing.T) {
	m1 := &InputField{
		LocationID: 1,
		TextTag:    "2",
		Value:      "3",
		pseudoID:   4,
	}
	assert.Equal(t, 4, m1.ID())
	m1.SetID(12345)
	assert.Equal(t, 4, m1.ID())
}

func TestInputField_UniqueKey(t *testing.T) {
	m1 := &InputField{
		LocationID: 1,
		TextTag:    "2",
		Value:      "3",
		pseudoID:   4,
	}
	assert.Equal(t, "1_2", m1.UniqueKey())
	m1.Value = "Changed"
	m1.pseudoID = 12345
	assert.Equal(t, "1_2", m1.UniqueKey())
}

func TestInputField_Equals(t *testing.T) {
	m1 := &InputField{
		LocationID: 1,
		TextTag:    "2",
		Value:      "3",
		pseudoID:   4,
	}
	m1_1 := &InputField{
		LocationID: 1,
		TextTag:    "2",
		Value:      "3",
		pseudoID:   12345,
	}
	m2 := &InputField{
		LocationID: 3,
		TextTag:    "2",
		Value:      "3",
		pseudoID:   4,
	}
	assert.True(t, m1.Equals(m1_1))
	assert.False(t, m1.Equals(m2))
}

func TestInputField_RelatedEntries(t *testing.T) {
	db := &Database{
		InputField: []*InputField{
			nil,
			{
				LocationID: 3,
				TextTag:    "A",
				Value:      "Bla",
			},
		},
		Location: []*Location{
			nil,
			nil,
			nil,
			{
				LocationID: 3,
				DocumentID: sql.NullInt64{12345, true},
				KeySymbol:  sql.NullString{"lffi", true},
			},
		},
	}

	assert.Equal(t, Related{}, db.InputField[1].RelatedEntries(nil))
	assert.Equal(t, Related{Location: db.Location[3]}, db.InputField[1].RelatedEntries(db))
}

func TestInputField_PrettyPrint(t *testing.T) {
	m1 := &InputField{
		LocationID: 3,
		TextTag:    "A",
		Value:      "Bla",
	}

	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	fmt.Fprint(w, "\nTextTag:\tA")
	fmt.Fprint(w, "\nValue:\tBla")
	w.Flush()
	expectedResult := buf.String()

	assert.Equal(t, expectedResult, m1.PrettyPrint(nil))
}

func TestInputField_MarshalJSON(t *testing.T) {
	m1 := &InputField{
		LocationID: 3,
		TextTag:    "A",
		Value:      "Bla",
		pseudoID:   12345,
	}
	result, err := json.Marshal(m1)
	assert.NoError(t, err)
	assert.Equal(t,
		`{"type":"InputField","locationId":3,"textTag":"A","value":"Bla"}`,
		string(result))
}
