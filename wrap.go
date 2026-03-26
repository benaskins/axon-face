package face

import "strings"

// WordWrap wraps text at the given width on word boundaries.
func WordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	var sb strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if len(line) <= width {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(line)
			continue
		}
		words := strings.Fields(line)
		col := 0
		for i, w := range words {
			if i > 0 && col+1+len(w) > width {
				sb.WriteString("\n")
				col = 0
			} else if i > 0 {
				sb.WriteString(" ")
				col++
			}
			sb.WriteString(w)
			col += len(w)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
