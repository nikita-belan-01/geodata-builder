package strutil

func Shorten(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}

	const ellipsis = "..."
	if maxLen <= len(ellipsis) {
		return s[:maxLen]
	}

	cut := maxLen - len(ellipsis)
	for cut > 0 && !utf8Start(s[cut]) {
		cut--
	}

	return s[:cut] + ellipsis
}

func utf8Start(b byte) bool {
	return b&0xC0 != 0x80
}
