package rtl

import (
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
	Isolated   rune
	Final      rune
	Initial    rune
	Medial     rune
	CanConnect bool
}

// Arabic letter forms mapping
var arabicLetters = map[rune]ArabicLetter{
	// Basic Arabic letters
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
	'\u0649': {Isolated: '\uFEEF', Final: '\uFEF0', Initial: '\u0649', Medial: '\uFEF0', CanConnect: false}, // Alef Maksura

	// Ligatures (Lam-Alef)
	'\uFEF5': {Isolated: '\uFEF5', Final: '\uFEF6', Initial: '\uFEF5', Medial: '\uFEF6', CanConnect: false}, // Lam-Alef with Madda
	'\uFEF7': {Isolated: '\uFEF7', Final: '\uFEF8', Initial: '\uFEF7', Medial: '\uFEF8', CanConnect: false}, // Lam-Alef with Hamza Above
	'\uFEF9': {Isolated: '\uFEF9', Final: '\uFEFA', Initial: '\uFEF9', Medial: '\uFEFA', CanConnect: false}, // Lam-Alef with Hamza Below
	'\uFEFB': {Isolated: '\uFEFB', Final: '\uFEFC', Initial: '\uFEFB', Medial: '\uFEFC', CanConnect: false}, // Lam-Alef
}

// isArabicLetter checks if rune is an Arabic letter
func isArabicLetter(r rune) bool {
	_, ok := arabicLetters[r]
	return ok
}

// isArabicConnectable checks if character can connect to next
func isArabicConnectable(r rune) bool {
	if letter, ok := arabicLetters[r]; ok {
		return letter.CanConnect
	}
	return false
}

// isTransparent returns true if the rune is a diacritic/transparent character
func isTransparent(r rune) bool {
	// Arabic Tashkeel (diacritics)
	if (r >= 0x064B && r <= 0x065F) || r == 0x0670 || r == 0x0656 || r == 0x0657 {
		return true
	}
	// Hebrew Niqqud
	if r >= 0x0591 && r <= 0x05C7 {
		return true
	}
	return false
}

// ShapeArabic applies Arabic contextual shaping
func ShapeArabic(text string) string {
	if !ContainsRTL(text) {
		return text
	}

	runes := []rune(text)
	var combined []rune

	// First pass: Handle Lam-Alef ligatures
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\u0644' { // Lam
			alefIdx := -1
			// Look ahead for Alef, skipping transparent characters
			for j := i + 1; j < len(runes); j++ {
				if isTransparent(runes[j]) {
					continue
				}
				if runes[j] == '\u0622' || runes[j] == '\u0623' || runes[j] == '\u0625' || runes[j] == '\u0627' {
					alefIdx = j
				}
				break
			}

			if alefIdx != -1 {
				next := runes[alefIdx]
				var ligature rune
				switch next {
				case '\u0622':
					ligature = '\uFEF5'
				case '\u0623':
					ligature = '\uFEF7'
				case '\u0625':
					ligature = '\uFEF9'
				case '\u0627':
					ligature = '\uFEFB'
				}
				combined = append(combined, ligature)
				// Preserve transparent characters that were on the Lam
				for k := i + 1; k < alefIdx; k++ {
					combined = append(combined, runes[k])
				}
				i = alefIdx
				continue
			}
		}
		combined = append(combined, r)
	}

	// Second pass: Shaping
	result := make([]rune, len(combined))
	for i, r := range combined {
		if isTransparent(r) {
			result[i] = r
			continue
		}

		letter, isArabic := arabicLetters[r]
		if !isArabic {
			result[i] = r
			continue
		}

		hasPrev := false
		for j := i - 1; j >= 0; j-- {
			if isTransparent(combined[j]) {
				continue
			}
			hasPrev = isArabicConnectable(combined[j])
			break
		}

		hasNext := false
		for j := i + 1; j < len(combined); j++ {
			if isTransparent(combined[j]) {
				continue
			}
			hasNext = isArabicLetter(combined[j])
			break
		}

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

// Bidi Types for simplified reordering algorithm
type BidiType int

const (
	TypeR BidiType = iota
	TypeL
	TypeEN
	TypeN
)

func getBidiType(r rune) BidiType {
	if IsRTLChar(r) {
		return TypeR
	}
	if unicode.IsDigit(r) {
		return TypeEN // Numbers
	}
	if unicode.IsLetter(r) {
		return TypeL
	}
	return TypeN
}

// reverseRTLRun reverses the slice but keeps transparent characters attached to their logical preceding base
func reverseRTLRun(runes []rune) []rune {
	type cluster struct {
		runes []rune
	}
	var clusters []cluster

	for i := 0; i < len(runes); i++ {
		if isTransparent(runes[i]) {
			if len(clusters) > 0 {
				clusters[len(clusters)-1].runes = append(clusters[len(clusters)-1].runes, runes[i])
			} else {
				clusters = append(clusters, cluster{runes: []rune{runes[i]}})
			}
		} else {
			clusters = append(clusters, cluster{runes: []rune{runes[i]}})
		}
	}

	var result []rune
	for i := len(clusters) - 1; i >= 0; i-- {
		result = append(result, clusters[i].runes...)
	}
	return result
}

// simpleBidiReorder performs a simplified UBA logical-to-visual string reordering algorithm
func simpleBidiReorder(runes []rune, isBaseRTL bool) []rune {
	if len(runes) == 0 {
		return runes
	}

	baseDir := TypeL
	if isBaseRTL {
		baseDir = TypeR
	}

	types := make([]BidiType, len(runes))
	for i, r := range runes {
		types[i] = getBidiType(r)
	}

	// Resolve Neutrals (TypeN) based on surrounding context
	for i := 0; i < len(runes); i++ {
		if types[i] == TypeN {
			start := i
			end := i
			for end+1 < len(runes) && types[end+1] == TypeN {
				end++
			}

			prevStrong := baseDir
			for j := start - 1; j >= 0; j-- {
				if types[j] == TypeR {
					prevStrong = TypeR
					break
				}
				if types[j] == TypeL || types[j] == TypeEN {
					prevStrong = TypeL
					break
				}
			}

			nextStrong := baseDir
			for j := end + 1; j < len(runes); j++ {
				if types[j] == TypeR {
					nextStrong = TypeR
					break
				}
				if types[j] == TypeL || types[j] == TypeEN {
					nextStrong = TypeL
					break
				}
			}

			resolved := baseDir
			if prevStrong == nextStrong {
				resolved = prevStrong
			}

			for j := start; j <= end; j++ {
				types[j] = resolved
			}
			i = end
		}
	}

	// Treat remaining EN blocks as LTR for alignment purposes
	for i := 0; i < len(runes); i++ {
		if types[i] == TypeEN {
			types[i] = TypeL
		}
	}

	// Segment text into distinct runs
	type Run struct {
		dir   BidiType
		start int
		end   int
	}

	var runs []Run
	currentDir := types[0]
	start := 0

	for i := 1; i < len(runes); i++ {
		if types[i] != currentDir {
			runs = append(runs, Run{dir: currentDir, start: start, end: i - 1})
			currentDir = types[i]
			start = i
		}
	}
	runs = append(runs, Run{dir: currentDir, start: start, end: len(runes) - 1})

	var result []rune

	if baseDir == TypeR {
		// Base is RTL, assemble visual string processing runs backward
		for i := len(runs) - 1; i >= 0; i-- {
			run := runs[i]
			if run.dir == TypeR {
				result = append(result, reverseRTLRun(runes[run.start:run.end+1])...)
			} else {
				for j := run.start; j <= run.end; j++ {
					result = append(result, runes[j])
				}
			}
		}
	} else {
		// Base is LTR, assemble visual string processing runs forward
		for i := 0; i < len(runs); i++ {
			run := runs[i]
			if run.dir == TypeR {
				result = append(result, reverseRTLRun(runes[run.start:run.end+1])...)
			} else {
				for j := run.start; j <= run.end; j++ {
					result = append(result, runes[j])
				}
			}
		}
	}

	return result
}

// ReorderString reorders a shaped logical string into a Left-to-Right printable visual string
func ReorderString(text string, isBaseRTL bool) string {
	return string(simpleBidiReorder([]rune(text), isBaseRTL))
}

// ProcessRTLText shapes and reorders text directly (used for non-wrapped segments)
func ProcessRTLText(text string) string {
	if !ContainsRTL(text) {
		return text
	}

	shaped := ShapeArabic(text)
	return ReorderString(shaped, IsRTLText(text))
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
