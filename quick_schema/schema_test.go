package quick_schema

import (
	"encoding/json"
	"testing"
	"time"

	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"

	"github.com/gofrs/uuid"
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v3"
)

type DataResponse[T dataer] struct {
	// Client sets this value and server echos data in the response
	Context string `json:"context,omitempty"`
	Data    T      `json:"data"`
}

type dataer interface {
	data()
}

type resp struct {
	A string `json:"a" description:"stuff aa" example:"here we go, travelling with jesus"`
	B string `json:"b"`
}

func (r resp) data() {

}

func diffJSON(a, b []byte) (string, error) {
	differ := diff.New()
	d, err := differ.Compare(a, b)
	if err != nil {
		return "", err
	}
	if d.Modified() {
		var aJson map[string]interface{}
		json.Unmarshal(a, &aJson)

		config := formatter.AsciiFormatterConfig{
			ShowArrayIndex: true,
			Coloring:       true,
		}

		formatter := formatter.NewAsciiFormatter(aJson, config)
		diffString, err := formatter.Format(d)
		if err != nil {
			return "", err
		}
		return diffString, nil
	}
	return "", nil
}

func TestGetSchema(t *testing.T) {
	expected := `{"Package":"github.com/pindamonhangaba/apiculi/quick_schema","Type":"DataResponse[github.com/pindamonhangaba/apiculi/quick_schema.resp]","Format":"object","Name":"","Description":"","Example":"","Children":[{"Package":"","Type":"string","Format":"string","Name":"context","Description":"","Example":"","Children":null},{"Package":"github.com/pindamonhangaba/apiculi/quick_schema","Type":"resp","Format":"object","Name":"data","Description":"","Example":"","Children":[{"Package":"","Type":"string","Format":"string","Name":"a","Description":"stuff aa","Example":"here we go, travelling with jesus","Children":null},{"Package":"","Type":"string","Format":"string","Name":"b","Description":"","Example":"","Children":null}]}]}`
	schema := GetSchema[DataResponse[resp]]()
	j, err := json.Marshal(schema)
	if err != nil {
		t.Error(err)
	}
	d, err := diffJSON([]byte(expected), j)
	if err != nil {
		t.Error(err)
	}
	if len(d) > 0 {
		t.Errorf("result not as expected:\n%v", d)
	}

}

func TestEmbededTypes(t *testing.T) {
	expected := `{"Package":"github.com/pindamonhangaba/apiculi/quick_schema","Type":"B","Format":"object","Name":"","Description":"","Example":"","Children":[{"Package":"","Type":"string","Format":"string","Name":"FirstMe","Description":"","Example":"","Children":null,"Omitempty":false},{"Package":"","Type":"int64","Format":"number","Name":"NUmberSutff","Description":"","Example":"","Children":null,"Omitempty":false},{"Package":"","Type":"int32","Format":"number","Name":"FirstMe","Description":"","Example":"","Children":null,"Omitempty":false},{"Package":"time","Type":"Time","Format":"string","Name":"NOthingHere","Description":"","Example":"","Children":null,"Omitempty":false}],"Omitempty":false}`
	type EMbedMe struct {
		FirstMe     string
		NUmberSutff int64
	}
	type B struct {
		EMbedMe
		FirstMe     int32
		NOthingHere time.Time
	}
	bodyTypeNodeSchema := GetSchema[B]()
	b, err := json.Marshal(bodyTypeNodeSchema)
	if err != nil {
		t.Error(err)
	}

	d, err := diffJSON([]byte(expected), b)
	if err != nil {
		t.Error(err)
	}
	if len(d) > 0 {
		t.Errorf("result not as expected:\n%v", d)
	}
}

type PushTokens struct {
	pq.StringArray
}

func TestComplexTypes(t *testing.T) {

	expected := `{"Package":"github.com/pindamonhangaba/apiculi/quick_schema","Type":"B","Format":"object","Name":"","Description":"","Example":"","Children":[{"Package":"github.com/gofrs/uuid","Type":"string","Format":"string","Name":"acveID","Description":"","Example":"","Children":null,"Omitempty":false},{"Package":"github.com/gofrs/uuid","Type":"UUID","Format":"string","Name":"ID","Description":"","Example":"","Children":null,"Omitempty":false},{"Package":"time","Type":"Time","Format":"string","Name":"createdAt","Description":"","Example":"","Children":null,"Omitempty":false},{"Package":"","Type":"string","Format":"string","Name":"opt","Description":"","Example":"","Children":null,"Omitempty":true},{"Package":"github.com/lib/pq","Type":"StringArray","Format":"slice","Name":"PushTokens","Description":"","Example":"","Children":[{"Package":"","Type":"string","Format":"string","Name":"","Description":"","Example":"","Children":null,"Omitempty":false}],"Omitempty":false},{"Package":"","Type":"","Format":"slice","Name":"PushTokens2","Description":"","Example":"","Children":[{"Package":"","Type":"string","Format":"string","Name":"","Description":"","Example":"","Children":null,"Omitempty":false}],"Omitempty":false},{"Package":"","Type":"","Format":"slice","Name":"PushTokens3","Description":"","Example":"","Children":[{"Package":"","Type":"uint8","Format":"number","Name":"","Description":"","Example":"","Children":null,"Omitempty":false}],"Omitempty":false}],"Omitempty":false}`

	type B struct {
		AcveID      uuid.UUID `db:"acve_id" json:"acveID" type:"string"`
		AnotherID   uuid.UUID `db:"id" json:"ID"`
		CreatedAt   time.Time `db:"created_at" json:"createdAt"`
		DeletedAt   null.Time `db:"deleted_at" json:"-"`
		Optional    string    `json:"opt,omitempty"`
		PushTokens  PushTokens
		PushTokens2 []string
		PushTokens3 []uint8
	}
	bodyTypeNodeSchema := GetSchema[B]()
	b, err := json.Marshal(bodyTypeNodeSchema)
	if err != nil {
		t.Error(err)
	}
	d, err := diffJSON([]byte(expected), b)
	if err != nil {
		t.Error(err)
	}
	if len(d) > 0 {
		t.Errorf("result not as expected:\n%v", d)
	}
}
