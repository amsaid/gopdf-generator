package rtl

import (
	"strings"
	"unicode"
)

// IsRTLChar checks if a rune is an RTL character
func IsRTLChar(r rune) bool {
	// Hebrew block: U+0590 to U+05FF
	if r >= 0x0590 && r <= 0x05FF {
		return true
	}
	// Arabic block: U+0600 to U+06FF
	if r >= 0x0600 && r <= 0x06FF {
		return true
	}
	// Arabic Supplement: U+0750 to U+077F
	if r >= 0x0750 && r <= 0x077F {
		return true
	}
	// Arabic Extended-A: U+08A0 to U+08FF
	if r >= 0x08A0 && r <= 0x08FF {
		return true
	}
	// Hebrew presentation forms: U+FB1D to U+FB4F
	if r >= 0xFB1D && r <= 0xFB4F {
		return true
	}
	// Arabic presentation forms A: U+FB50 to U+FDFF
	if r >= 0xFB50 && r <= 0xFDFF {
		return true
	}
	// Arabic presentation forms B: U+FE70 to U+FEFF
	if r >= 0xFE70 && r <= 0xFEFF {
		return true
	}
	// RLM, RTL marks
	if r == 0x200F || r == 0x202E || r == 0x202B {
		return true
	}
	return false
}

// ContainsRTL checks if text contains RTL characters
func ContainsRTL(text string) bool {
	for _, r := range text {
		if IsRTLChar(r) {
			return true
		}
	}
	return false
}

// IsRTLText checks if text is primarily RTL
func IsRTLText(text string) bool {
	rtlCount := 0
	totalCount := 0
	
	for _, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		totalCount++
		if IsRTLChar(r) {
			rtlCount++
		}
	}
	
	if totalCount == 0 {
		return false
	}
	return float64(rtlCount)/float64(totalCount) >= 0.5
}

// ArabicLetter represents an Arabic letter with its contextual forms
type ArabicLetter struct {
	Isolated  rune
	Final     rune
	Initial   rune
	Medial    rune
	CanConnect bool
}

// Arabic letter forms mapping
var arabicLetters = map[rune]ArabicLetter{
	// Arabic letters with their contextual forms
	'\u0627': {Isolated: '\uFE8D', Final: '\uFE8E', Initial: '\u0627', Medial: '\uFE8E', CanConnect: false}, // Alef
	'\u0628': {Isolated: '\uFE8F', Final: '\uFE90', Initial: '\uFE91', Medial: '\uFE92', CanConnect: true},  // Beh
	'\u062A': {Isolated: '\uFE95', Final: '\uFE96', Initial: '\uFE97', Medial: '\uFE98', CanConnect: true},  // Teh
	'\u062B': {Isolated: '\uFE99', Final: '\uFE9A', Initial: '\uFE9B', Medial: '\uFE9C', CanConnect: true},  // Theh
	'\u062C': {Isolated: '\uFE9D', Final: '\uFE9E', Initial: '\uFE9F', Medial: '\uFEA0', CanConnect: true},  // Jeem
	'\u062D': {Isolated: '\uFEA1', Final: '\uFEA2', Initial: '\uFEA3', Medial: '\uFEA4', CanConnect: true},  // Hah
	'\u062E': {Isolated: '\uFEA5', Final: '\uFEA6', Initial: '\uFEA7', Medial: '\uFEA8', CanConnect: true},  // Khah
	'\u062F': {Isolated: '\uFEA9', Final: '\uFEAA', Initial: '\u062F', Medial: '\uFEAA', CanConnect: false}, // Dal
	'\u0630': {Isolated: '\uFEAB', Final: '\uFEAC', Initial: '\u0630', Medial: '\uFEAC', CanConnect: false}, // Thal
	'\u0631': {Isolated: '\uFEAD', Final: '\uFEAE', Initial: '\u0631', Medial: '\uFEAE', CanConnect: false}, // Reh
	'\u0632': {Isolated: '\uFEAF', Final: '\uFEB0', Initial: '\u0632', Medial: '\uFEB0', CanConnect: false}, // Zain
	'\u0633': {Isolated: '\uFEB1', Final: '\uFEB2', Initial: '\uFEB3', Medial: '\uFEB4', CanConnect: true},  // Seen
	'\u0634': {Isolated: '\uFEB5', Final: '\uFEB6', Initial: '\uFEB7', Medial: '\uFEB8', CanConnect: true},  // Sheen
	'\u0635': {Isolated: '\uFEB9', Final: '\uFEBA', Initial: '\uFEBB', Medial: '\uFEBC', CanConnect: true},  // Sad
	'\u0636': {Isolated: '\uFEBD', Final: '\uFEBE', Initial: '\uFEBF', Medial: '\uFEC0', CanConnect: true},  // Dad
	'\u0637': {Isolated: '\uFEC1', Final: '\uFEC2', Initial: '\uFEC3', Medial: '\uFEC4', CanConnect: true},  // Tah
	'\u0638': {Isolated: '\uFEC5', Final: '\uFEC6', Initial: '\uFEC7', Medial: '\uFEC8', CanConnect: true},  // Zah
	'\u0639': {Isolated: '\uFEC9', Final: '\uFECA', Initial: '\uFECB', Medial: '\uFECC', CanConnect: true},  // Ain
	'\u063A': {Isolated: '\uFECD', Final: '\uFECE', Initial: '\uFECF', Medial: '\uFED0', CanConnect: true},  // Ghain
	'\u0641': {Isolated: '\uFED1', Final: '\uFED2', Initial: '\uFED3', Medial: '\uFED4', CanConnect: true},  // Feh
	'\u0642': {Isolated: '\uFED5', Final: '\uFED6', Initial: '\uFED7', Medial: '\uFED8', CanConnect: true},  // Qaf
	'\u0643': {Isolated: '\uFED9', Final: '\uFEDA', Initial: '\uFEDB', Medial: '\uFEDC', CanConnect: true},  // Kaf
	'\u0644': {Isolated: '\uFEDD', Final: '\uFEDE', Initial: '\uFEDF', Medial: '\uFEE0', CanConnect: true},  // Lam
	'\u0645': {Isolated: '\uFEE1', Final: '\uFEE2', Initial: '\uFEE3', Medial: '\uFEE4', CanConnect: true},  // Meem
	'\u0646': {Isolated: '\uFEE5', Final: '\uFEE6', Initial: '\uFEE7', Medial: '\uFEE8', CanConnect: true},  // Noon
	'\u0647': {Isolated: '\uFEE9', Final: '\uFEEA', Initial: '\uFEEB', Medial: '\uFEEC', CanConnect: true},  // Heh
	'\u0648': {Isolated: '\uFEED', Final: '\uFEEE', Initial: '\u0648', Medial: '\uFEEE', CanConnect: false}, // Waw
	'\u064A': {Isolated: '\uFEF1', Final: '\uFEF2', Initial: '\uFEF3', Medial: '\uFEF4', CanConnect: true},  // Yeh
	'\u0621': {Isolated: '\uFE80', Final: '\uFE80', Initial: '\u0621', Medial: '\uFE80', CanConnect: false}, // Hamza
	'\u0622': {Isolated: '\uFE81', Final: '\uFE82', Initial: '\u0622', Medial: '\uFE82', CanConnect: false}, // Alef with Madda
	'\u0623': {Isolated: '\uFE83', Final: '\uFE84', Initial: '\u0623', Medial: '\uFE84', CanConnect: false}, // Alef with Hamza Above
	'\u0625': {Isolated: '\uFE87', Final: '\uFE88', Initial: '\u0625', Medial: '\uFE88', CanConnect: false}, // Alef with Hamza Below
	'\u0626': {Isolated: '\uFE89', Final: '\uFE8A', Initial: '\uFE8B', Medial: '\uFE8C', CanConnect: true},  // Yeh with Hamza Above
	'\u0629': {Isolated: '\uFE93', Final: '\uFE94', Initial: '\u0629', Medial: '\uFE94', CanConnect: false}, // Teh Marbuta
}

// isArabicConnectable checks if character can connect to next
func isArabicConnectable(r rune) bool {
	if letter, ok := arabicLetters[r]; ok {
		return letter.CanConnect
	}
	return false
}

// ShapeArabic applies Arabic contextual shaping
func ShapeArabic(text string) string {
	if !ContainsRTL(text) {
		return text
	}
	
	runes := []rune(text)
	result := make([]rune, len(runes))
	
	for i, r := range runes {
		letter, isArabic := arabicLetters[r]
		if !isArabic {
			result[i] = r
			continue
		}
		
		// Determine position
		hasPrev := i > 0 && isArabicConnectable(runes[i-1])
		hasNext := i < len(runes)-1 && isArabicLetter(runes[i+1])
		
		switch {
		case hasPrev && hasNext:
			result[i] = letter.Medial
		case hasPrev:
			result[i] = letter.Final
		case hasNext:
			result[i] = letter.Initial
		default:
			result[i] = letter.Isolated
		}
	}
	
	return string(result)
}

// isArabicLetter checks if rune is an Arabic letter
func isArabicLetter(r rune) bool {
	_, ok := arabicLetters[r]
	return ok
}

// ReverseString reverses a string (for RTL display)
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ProcessRTLText processes text for RTL display
// This includes reversing the text and applying Arabic shaping
func ProcessRTLText(text string) string {
	if !ContainsRTL(text) {
		return text
	}
	
	// Apply Arabic shaping
	shaped := ShapeArabic(text)
	
	// For pure RTL text, reverse for proper display
	if IsRTLText(text) {
		// Split into words and reverse order
		words := strings.Fields(shaped)
		for i, j := 0, len(words)-1; i < j; i, j = i+1, j-1 {
			words[i], words[j] = words[j], words[i]
		}
		return strings.Join(words, " ")
	}
	
	return shaped
}

// GetTextDirection returns the direction of text
func GetTextDirection(text string) string {
	if IsRTLText(text) {
		return "rtl"
	}
	return "ltr"
}

// RTLString wraps text with RTL marks if needed
func RTLString(text string) string {
	if IsRTLText(text) {
		// Add RTL mark at start and end
		return "\u202B" + text + "\u202C"
	}
	return text
}

// LTRString wraps text with LTR marks if needed
func LTRString(text string) string {
	return "\u202A" + text + "\u202C"
}