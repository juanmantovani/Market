package domain

import "strings"

var accentFolding = strings.NewReplacer(
	"á", "a", "à", "a", "ä", "a", "â", "a", "ã", "a",
	"é", "e", "è", "e", "ë", "e", "ê", "e",
	"í", "i", "ì", "i", "ï", "i", "î", "i",
	"ó", "o", "ò", "o", "ö", "o", "ô", "o", "õ", "o",
	"ú", "u", "ù", "u", "ü", "u", "û", "u",
	"ñ", "n", "ç", "c",
)

// NormalizeText folds case, accents and spacing, so searches and uniqueness
// checks treat "  Café   Molido" and "cafe molido" as the same text.
func NormalizeText(text string) string {
	folded := accentFolding.Replace(strings.ToLower(text))

	return strings.Join(strings.Fields(folded), " ")
}
