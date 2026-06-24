package msg

import (
	"regexp"
	"strings"
)

// mdv2Escapes matches characters that must be backslash-escaped in Telegram MarkdownV2.
var mdv2Escapes = regexp.MustCompile(`([_*\[\]()~` + "`" + `>#+\-=|{}.!\\])`)

// FormatTelegramMarkdown converts standard Markdown to Telegram MarkdownV2 format.
// Similar to Hermes' format_message().
func FormatTelegramMarkdown(text string) string {
	// 0) Convert GFM pipe tables to bullet groups (MarkdownV2 has no table syntax)
	text = wrapMarkdownTables(text)

	// 1) Protect fenced code blocks and inline code from escaping
	text = protectCodeBlocks(text)

	// 2) Convert bold: **text** → *text*
	text = reBold.ReplaceAllStringFunc(text, func(m string) string {
		inner := m[2 : len(m)-2]
		return "*" + escapeMDV2(inner) + "*"
	})

	// 3) Convert italic: *text* → _text_
	text = reItalic.ReplaceAllStringFunc(text, func(m string) string {
		inner := m[1 : len(m)-1]
		return "_" + escapeMDV2(inner) + "_"
	})

	// 4) Convert markdown links [text](url) → [text](url) with escaped display
	text = reLink.ReplaceAllStringFunc(text, func(m string) string {
		parts := reLink.FindStringSubmatch(m)
		display := escapeMDV2(parts[1])
		url := escapeMDV2URL(parts[2])
		return "[" + display + "](" + url + ")"
	})

	// 5) Escape remaining special characters (outside protected blocks)
	text = restoreCodeBlocks(text)

	return text
}

var (
	reBold   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	reItalic = regexp.MustCompile(`(?<!\*)\*(?!\*)(.+?)(?<!\*)\*(?!\*)`)
	reLink   = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
)

// Code block protection to prevent escaping inside code
var codeBlockID int

func protectCodeBlocks(text string) string {
	codeBlockID = 0
	// Fenced code blocks: ```...```
	text = reFenced.ReplaceAllStringFunc(text, func(m string) string {
		id := codeBlockID
		codeBlockID++
		return placeholder(id)
	})
	// Inline code: `...`
	text = reInlineCode.ReplaceAllStringFunc(text, func(m string) string {
		id := codeBlockID
		codeBlockID++
		return placeholder(id)
	})
	return text
}

func restoreCodeBlocks(text string) string {
	for i := 0; i < codeBlockID; i++ {
		text = strings.Replace(text, placeholder(i), codeBlocks[i], 1)
	}
	return text
}

var (
	reFenced     = regexp.MustCompile("```[\\s\\S]*?```")
	reInlineCode = regexp.MustCompile("`[^`\n]+`")
	codeBlocks   []string
)

func placeholder(id int) string {
	return "\x00CODE" + itoa(id) + "\x00"
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

func escapeMDV2(s string) string {
	return mdv2Escapes.ReplaceAllString(s, `\$1`)
}

func escapeMDV2URL(s string) string {
	// Only ) and \ need escaping inside URL
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `)`, `\)`)
	return s
}

// wrapMarkdownTables converts GFM pipe tables to Telegram-friendly bullet groups.
// Telegram MarkdownV2 has no table syntax.
func wrapMarkdownTables(text string) string {
	lines := strings.Split(text, "\n")
	var out []string
	i := 0
	for i < len(lines) {
		line := lines[i]
		if isTableDelimiter(line) && i > 0 && i+1 < len(lines) {
			// Found a table: header at i-1, delimiter at i, data rows follow
			header := lines[i-1]
			headers := splitTableRow(header)
			// Remove the header line from output
			if len(out) > 0 {
				out = out[:len(out)-1]
			}
			// Collect data rows
			var dataRows [][]string
			j := i + 1
			for j < len(lines) && isTableRow(lines[j]) {
				dataRows = append(dataRows, splitTableRow(lines[j]))
				j++
			}
			// Render as bullet groups
			out = append(out, renderTableAsBullets(headers, dataRows))
			i = j
			continue
		}
		out = append(out, line)
		i++
	}
	return strings.Join(out, "\n")
}

func isTableDelimiter(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.Contains(trimmed, "|") {
		return false
	}
	// Must contain only |, -, :, spaces
	for _, c := range trimmed {
		if c != '|' && c != '-' && c != ':' && c != ' ' {
			return false
		}
	}
	return strings.Contains(trimmed, "-")
}

func isTableRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed != "" && strings.Contains(trimmed, "|")
}

func splitTableRow(line string) []string {
	trimmed := strings.Trim(line, "| ")
	var cells []string
	for _, cell := range strings.Split(trimmed, "|") {
		cells = append(cells, strings.TrimSpace(cell))
	}
	return cells
}

func renderTableAsBullets(headers []string, rows [][]string) string {
	var out []string
	for _, row := range rows {
		for h := 0; h < len(headers) && h < len(row); h++ {
			if row[h] != "" && row[h] != "-" {
				out = append(out, "*"+escapeMDV2(headers[h])+"*: "+row[h])
			}
		}
	}
	return strings.Join(out, "\n")
}
