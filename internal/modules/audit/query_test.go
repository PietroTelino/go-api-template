package audit

import (
	"reflect"
	"testing"
)

func TestBuildWhereClause(t *testing.T) {
	testCases := []struct {
		name      string
		filters   Filters
		wantWhere string
		wantArgs  []any
	}{
		{
			name:      "sem filtros",
			filters:   Filters{},
			wantWhere: "",
			wantArgs:  []any{},
		},
		{
			name:      "um filtro",
			filters:   Filters{Action: "USER_LOGIN"},
			wantWhere: " WHERE a.action = $1",
			wantArgs:  []any{"USER_LOGIN"},
		},
		{
			name:      "vários filtros numerados em ordem",
			filters:   Filters{UserID: "id-1", Action: "USER_LOGIN", Entity: "USER"},
			wantWhere: " WHERE a.user_id = $1 AND a.action = $2 AND a.entity = $3",
			wantArgs:  []any{"id-1", "USER_LOGIN", "USER"},
		},
		{
			name:      "valor com aspas vira parâmetro, não SQL",
			filters:   Filters{Action: "x' OR 1=1--"},
			wantWhere: " WHERE a.action = $1",
			wantArgs:  []any{"x' OR 1=1--"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			gotWhere, gotArgs := buildWhereClause(testCase.filters)

			if gotWhere != testCase.wantWhere {
				t.Errorf("where = %q, esperado %q", gotWhere, testCase.wantWhere)
			}

			if !reflect.DeepEqual(gotArgs, testCase.wantArgs) {
				t.Errorf("args = %v, esperado %v", gotArgs, testCase.wantArgs)
			}
		})
	}
}
