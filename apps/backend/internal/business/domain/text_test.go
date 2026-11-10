package domain

import "testing"

func TestNormalizeText(t *testing.T) {
	tests := map[string]string{
		"  Café   Molido ": "cafe molido",
		"PIÑA":             "pina",
		"Pingüino\tAzúcar": "pinguino azucar",
		"":                 "",
		"   ":              "",
		"Coca-Cola 500ml":  "coca-cola 500ml",
	}

	for input, want := range tests {
		if got := NormalizeText(input); got != want {
			t.Errorf("NormalizeText(%q) = %q, want %q", input, got, want)
		}
	}
}
