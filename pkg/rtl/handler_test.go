package rtl

import (
	"testing"
)

func TestIsRTLChar(t *testing.T) {
	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		{"Arabic alef", '\u0627', true},
		{"Arabic beh", '\u0628', true},
		{"Hebrew alef", '\u05D0', true},
		{"Latin A", 'A', false},
		{"Latin a", 'a', false},
		{"Digit 1", '1', false},
		{"Space", ' ', false},
		{"Arabic in presentation forms", '\uFE8D', true},
		{"RLM mark", '\u200F', true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRTLChar(tt.char)
			if result != tt.expected {
				t.Errorf("IsRTLChar(%U) = %v, expected %v", tt.char, result, tt.expected)
			}
		})
	}
}

func TestContainsRTL(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"Pure Arabic", "مرحبا", true},
		{"Pure Hebrew", "שלום", true},
		{"Pure English", "Hello World", false},
		{"Mixed Arabic-English", "Hello مرحبا", true},
		{"Empty string", "", false},
		{"Numbers only", "12345", false},
		{"Arabic with numbers", "١٢٣", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsRTL(tt.text)
			if result != tt.expected {
				t.Errorf("ContainsRTL(%q) = %v, expected %v", tt.text, result, tt.expected)
			}
		})
	}
}

func TestIsRTLText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"Pure Arabic", "مرحبا بالعالم", true},
		{"Pure Hebrew", "שלום עולם", true},
		{"Pure English", "Hello World", false},
		{"Mostly Arabic with some English", "مرحبا بالعالم Hello", true},
		{"Mostly English with some Arabic", "Hello World مرحبا", false},
		{"Empty string", "", false},
		{"Punctuation only", "...!?", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRTLText(tt.text)
			if result != tt.expected {
				t.Errorf("IsRTLText(%q) = %v, expected %v", tt.text, result, tt.expected)
			}
		})
	}
}

func TestReorderString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		isBaseRTL bool
		expected  string
	}{
		{"Simple ASCII", "Hello", false, "Hello"},
		{"Pure RTL", "مرحبا", true, "ابحرم"},
		{"Mixed RTL base", "Hello مرحبا", true, "ابحرم Hello"},
		{"Mixed LTR base", "Hello مرحبا", false, "Hello ابحرم"},
		{"Mixed RTL w/ Punctuation", "مرحبا!", true, "!ابحرم"},
		{"Mixed w/ Numbers RTL base", "123 مرحبا", true, "ابحرم 123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReorderString(tt.input, tt.isBaseRTL)
			if result != tt.expected {
				t.Errorf("ReorderString(%q, %v) = %q, expected %q", tt.input, tt.isBaseRTL, result, tt.expected)
			}
		})
	}
}

func TestShapeArabic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Arabic word", "مرحبا"},
		{"Arabic phrase", "السلام عليكم"},
		{"Non-Arabic", "Hello"},
		{"Empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShapeArabic(tt.input)
			if !ContainsRTL(tt.input) && result != tt.input {
				t.Errorf("ShapeArabic(%q) = %q, expected same for non-Arabic", tt.input, result)
			}
		})
	}
}

func TestGetTextDirection(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{"Arabic", "مرحبا", "rtl"},
		{"Hebrew", "שלום", "rtl"},
		{"English", "Hello", "ltr"},
		{"Empty", "", "ltr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTextDirection(tt.text)
			if result != tt.expected {
				t.Errorf("GetTextDirection(%q) = %q, expected %q", tt.text, result, tt.expected)
			}
		})
	}
}

func TestRTLString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Arabic", "مرحبا"},
		{"English", "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RTLString(tt.input)
			if IsRTLText(tt.input) {
				runes := []rune(result)
				if len(runes) == 0 || runes[0] != '\u202B' {
					t.Errorf("RTLString(%q) should start with RTL mark", tt.input)
				}
				if len(runes) == 0 || runes[len(runes)-1] != '\u202C' {
					t.Errorf("RTLString(%q) should end with PDF mark", tt.input)
				}
			}
		})
	}
}

func TestProcessRTLText(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Arabic phrase", "السلام عليكم"},
		{"Mixed content", "Hello مرحبا"},
		{"English only", "Hello World"},
		{"Empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProcessRTLText(tt.input)
			if tt.input != "" && result == "" {
				t.Errorf("ProcessRTLText(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestArabicLetterForms(t *testing.T) {
	arabicChars := []rune{
		'\u0627', // Alef
		'\u0628', // Beh
		'\u062A', // Teh
		'\u0645', // Meem
		'\u0646', // Noon
	}

	for _, char := range arabicChars {
		letter, exists := arabicLetters[char]
		if !exists {
			t.Errorf("Arabic letter %U not found in arabicLetters map", char)
			continue
		}

		if letter.Isolated == 0 {
			t.Errorf("Letter %U has no isolated form", char)
		}
		if letter.Final == 0 {
			t.Errorf("Letter %U has no final form", char)
		}
	}
}
