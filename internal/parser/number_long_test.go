package parser

import (
	"encoding/json"
	"testing"
)

func TestNumberLong(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    numberLong
		wantErr bool
	}{
		{"Plain number", `1006`, 1006, false},
		{"Negative number", `-42`, -42, false},
		{"Mongo format", `{"$numberLong":"123"}`, 123, false},
		{"Quoted string", `"456"`, 456, false},
		{"Null value", `null`, 0, false},
		{"Invalid string", `"abc"`, 0, true},
		{"Invalid format", `[]`, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nl numberLong
			err := json.Unmarshal([]byte(tt.input), &nl)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && nl != tt.want {
				t.Errorf("Got = %v, want %v", nl, tt.want)
			}
		})
	}
}

func TestMarshalNumberLong(t *testing.T) {
	nl := numberLong(789)
	got, err := json.Marshal(nl)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(got) != "789" {
		t.Errorf("Marshal = %v, want 789", string(got))
	}
}
