package domain

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

const MaxEmailLength = 254

var emailShape = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

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

// requiredText trims text and checks it is present and at most maxRunes
// long.
func requiredText(text string, maxRunes int, missing, tooLong *Error) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", missing
	}
	if utf8.RuneCountInString(trimmed) > maxRunes {
		return "", tooLong
	}

	return trimmed, nil
}

// optionalText trims text, treating a blank value as absent, and checks it
// is at most maxRunes long. The result never aliases the caller's string
// pointer.
func optionalText(text *string, maxRunes int, tooLong *Error) (*string, error) {
	if text == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*text)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > maxRunes {
		return nil, tooLong
	}

	return &trimmed, nil
}

// validateEmail trims an optional email and checks it looks like one. Empty
// means no email.
func validateEmail(email *string, invalid *Error) (*string, error) {
	if email == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*email)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > MaxEmailLength || !emailShape.MatchString(trimmed) {
		return nil, invalid
	}

	return &trimmed, nil
}
