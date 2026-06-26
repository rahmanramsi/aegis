package messaging

import (
	"regexp"
	"strings"
)

var mdv2Escapes = regexp.MustCompile(`([_*\[\]()~` + "`" + `>#+\-=|{}.!\\])`)

// FormatTelegramMarkdown converts standard Markdown to Telegram MarkdownV2.
func FormatTelegramMarkdown(text string) string {
	// 0) Protect code blocks from escaping
	protected := protectBlocks(text)

	// 1) Convert GFM pipe tables to bullet groups
	protected = wrapMarkdownTables(protected)

	// 2) Convert bold: **text** → *text*
	protected = reBold.ReplaceAllStringFunc(protected, func(m string) string {
		return "*" + escapeMDV2(m[2:len(m)-2]) + "*"
	})

	// 3) Convert links [text](url) → escaped display + clean URL
	protected = reLink.ReplaceAllStringFunc(protected, func(m string) string {
		parts := reLink.FindStringSubmatch(m)
		return "[" + escapeMDV2(parts[1]) + "](" + escapeMDV2URL(parts[2]) + ")"
	})

	// 4) Escape remaining special chars, then restore protected blocks
	result := escapeMDV2(protected)
	result = restoreBlocks(result)
	return result
}

var reBold = regexp.MustCompile(`\*\*(.+?)\*\*`)
var reLink = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
var reFenced = regexp.MustCompile("(?s)```.*?```")
var reInlineCode = regexp.MustCompile("`[^`\n]+`")

var blockStore []string

func protectBlocks(text string) string {
	blockStore = nil
	text = reFenced.ReplaceAllStringFunc(text, func(m string) string {
		blockStore = append(blockStore, m)
		return "\x00FENCE" + itoa(len(blockStore)-1) + "\x00"
	})
	text = reInlineCode.ReplaceAllStringFunc(text, func(m string) string {
		blockStore = append(blockStore, m)
		return "\x00CODE" + itoa(len(blockStore)-1) + "\x00"
	})
	return text
}

func restoreBlocks(text string) string {
	for i := len(blockStore) - 1; i >= 0; i-- {
		text = strings.ReplaceAll(text, "\x00FENCE"+itoa(i)+"\x00", blockStore[i])
		text = strings.ReplaceAll(text, "\x00CODE"+itoa(i)+"\x00", blockStore[i])
	}
	return text
}

func escapeMDV2(s string) string {
	return mdv2Escapes.ReplaceAllString(s, `\$1`)
}

func escapeMDV2URL(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `)`, `\)`)
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// ---- Table handling ----

var reTableSep = regexp.MustCompile(`^\s*\|?\s*:?-+:?\s*(?:\|\s*:?-+:?\s*)+`)
var reTableData = regexp.MustCompile(`^\s*\|.+\|\s*$`)

func wrapMarkdownTables(text string) string {
	lines := strings.Split(text, "\n")
	var out []string
	i := 0
	for i < len(lines) {
		if reTableSep.MatchString(lines[i]) && i > 0 {
			headers := splitCells(lines[i-1])
			out = out[:len(out)-1] // remove header line
			j := i + 1
			var rows [][]string
			for j < len(lines) && reTableData.MatchString(lines[j]) {
				rows = append(rows, splitCells(lines[j]))
				j++
			}
			out = append(out, renderTable(headers, rows))
			i = j
			continue
		}
		out = append(out, lines[i])
		i++
	}
	return strings.Join(out, "\n")
}

func splitCells(line string) []string {
	line = strings.Trim(line, "| \t")
	var cells []string
	for _, c := range strings.Split(line, "|") {
		cells = append(cells, strings.TrimSpace(c))
	}
	return cells
}

func renderTable(headers []string, rows [][]string) string {
	var out []string
	for _, row := range rows {
		var parts []string
		for h := 0; h < len(headers) && h < len(row); h++ {
			if row[h] != "" && row[h] != "-" {
				parts = append(parts, "*"+headers[h]+"*: "+row[h])
			}
		}
		if len(parts) > 0 {
			out = append(out, strings.Join(parts, "\n"))
		}
	}
	return strings.Join(out, "\n")
}
