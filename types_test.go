package hivego

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestParseAsset(t *testing.T) {
	tests := []struct {
		input     string
		amount    int64
		precision uint8
		symbol    string
	}{
		{"1.000 HIVE", 1000, 3, "HIVE"},
		{"5.321 HBD", 5321, 3, "HBD"},
		{"1234.567890 VESTS", 1234567890, 6, "VESTS"},
		{"100 HIVE", 100, 0, "HIVE"},
		{"0.000 HIVE", 0, 3, "HIVE"},
		{"0.000000 VESTS", 0, 6, "VESTS"},
		{"1000000.000 HBD", 1000000000, 3, "HBD"},
	}
	for _, tt := range tests {
		a, err := ParseAsset(tt.input)
		if err != nil {
			t.Errorf("ParseAsset(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if a.Amount != tt.amount || a.Precision != tt.precision || a.Symbol != tt.symbol {
			t.Errorf("ParseAsset(%q) = {%d, %d, %q}, want {%d, %d, %q}",
				tt.input, a.Amount, a.Precision, a.Symbol,
				tt.amount, tt.precision, tt.symbol)
		}
	}
}

func TestParseAssetError(t *testing.T) {
	invalid := []string{"bad", "1.000", "1.000HIVE", "", "   "}
	for _, s := range invalid {
		_, err := ParseAsset(s)
		if !errors.Is(err, ErrInvalidAsset) {
			t.Errorf("ParseAsset(%q): expected ErrInvalidAsset, got %v", s, err)
		}
	}
}

func TestAssetString(t *testing.T) {
	tests := []struct {
		asset Asset
		want  string
	}{
		{Asset{1000, 3, "HIVE"}, "1.000 HIVE"},
		{Asset{5321, 3, "HBD"}, "5.321 HBD"},
		{Asset{1234567890, 6, "VESTS"}, "1234.567890 VESTS"},
		{Asset{100, 0, "HIVE"}, "100 HIVE"},
		{Asset{0, 3, "HIVE"}, "0.000 HIVE"},
		{Asset{1000000000, 3, "HBD"}, "1000000.000 HBD"},
	}
	for _, tt := range tests {
		got := tt.asset.String()
		if got != tt.want {
			t.Errorf("Asset{%d,%d,%q}.String() = %q, want %q",
				tt.asset.Amount, tt.asset.Precision, tt.asset.Symbol, got, tt.want)
		}
	}
}

func TestAssetStringRoundTrip(t *testing.T) {
	inputs := []string{"1.000 HIVE", "5.321 HBD", "1234.567890 VESTS", "0.000000 VESTS"}
	for _, s := range inputs {
		a, err := ParseAsset(s)
		if err != nil {
			t.Fatalf("ParseAsset(%q): %v", s, err)
		}
		if got := a.String(); got != s {
			t.Errorf("round-trip %q: got %q", s, got)
		}
	}
}

func TestAssetMarshalJSON(t *testing.T) {
	b, err := json.Marshal(Asset{Amount: 1000, Precision: 3, Symbol: "HIVE"})
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if got, want := string(b), `"1.000 HIVE"`; got != want {
		t.Fatalf("Asset JSON = %s, want %s", got, want)
	}
}

func TestAssetUnmarshalJSONString(t *testing.T) {
	var a Asset
	if err := json.Unmarshal([]byte(`"5.321 HBD"`), &a); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if a != (Asset{Amount: 5321, Precision: 3, Symbol: "HBD"}) {
		t.Fatalf("decoded asset = %#v", a)
	}
}

func TestAssetUnmarshalJSONLegacyObject(t *testing.T) {
	var a Asset
	if err := json.Unmarshal([]byte(`{"Amount":1000,"Precision":3,"Symbol":"HIVE"}`), &a); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if a != (Asset{Amount: 1000, Precision: 3, Symbol: "HIVE"}) {
		t.Fatalf("decoded asset = %#v", a)
	}
}
