// ABOUTME: Parses Org :letter: sections into recipient-specific cover letters.
// ABOUTME: Applies document and recipient placeholders without coupling parsing to rendering.
package letter

import (
	"fmt"
	"regexp"
	"strings"

	nativesource "github.com/tigger-developer/first-folio/internal/source"
)

type Letter struct {
	Sender       string
	Subject      string
	Body         string
	OfficeBody   string
	Footnotes    map[string][]string
	Closing      string
	Signoff      string
	Recipient    string
	Organization string
	Moniker      string
	Date         string
	Author       string
	Email        string
	Contact      string
}

var (
	frontMatterRE = regexp.MustCompile(`^#\+(\w+):\s*(.*)$`)
	taggedH2RE    = regexp.MustCompile(`^\*\*\s+(.*?)\s+:(sender|subject):\s*$`)
	taggedH3RE    = regexp.MustCompile(`^\*{3}\s+(.*?)\s+:(closing|signoff|to):\s*$`)
	taggedH4RE    = regexp.MustCompile(`^\*{4}\s+(.*?)\s+:(\w+):\s*$`)
	dateHeadingRE = regexp.MustCompile(`^\*{3}\s+[\[<](\d{4}-\d{2}-\d{2})`)
	placeholderRE = regexp.MustCompile(`\[(\w+)\]`)
)

func ParseOrg(source string) ([]Letter, error) {
	metadata := map[string]string{}
	var letters []Letter
	inLetters := false
	sender := ""
	subject, body, closing, signoff, date := "", "", "", "", ""
	var recipients []map[string]string

	flush := func() {
		if subject == "" || len(recipients) == 0 {
			return
		}
		trimmedBody := strings.TrimSpace(body)
		for _, recipient := range recipients {
			values := map[string]string{}
			for key, value := range metadata {
				values[key] = value
			}
			for key, value := range recipient {
				values[key] = value
			}
			resolved := placeholderRE.ReplaceAllStringFunc(trimmedBody, func(match string) string {
				key := placeholderRE.FindStringSubmatch(match)[1]
				if value, ok := values[key]; ok {
					return value
				}
				return match
			})
			officeBody, footnotes := nativesource.ExtractOrgFootnotes(resolved)
			letterClosing := closing
			if letterClosing == "" {
				letterClosing = "Yours sincerely"
			}
			letterSignoff := signoff
			if letterSignoff == "" {
				letterSignoff = metadata["author"]
			}
			letterSender := sender
			if letterSender == "" {
				letterSender = metadata["author"]
			}
			letters = append(letters, Letter{
				Sender: letterSender, Subject: subject, Body: resolved, OfficeBody: officeBody, Footnotes: footnotes, Closing: letterClosing, Signoff: letterSignoff,
				Recipient: recipient["address"], Organization: recipient["org"], Moniker: recipient["moniker"], Date: date,
				Author: metadata["author"], Email: metadata["email"], Contact: metadata["contact"],
			})
		}
		subject, body, closing, signoff, date = "", "", "", "", ""
		recipients = nil
	}

	for _, line := range strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n") {
		if match := frontMatterRE.FindStringSubmatch(line); match != nil {
			metadata[strings.ToLower(match[1])] = match[2]
			continue
		}
		if strings.HasPrefix(line, "* ") && strings.Contains(line, ":letter:") {
			inLetters = true
			continue
		}
		if strings.HasPrefix(line, "* ") && !strings.Contains(line, ":letter:") {
			flush()
			inLetters = false
			continue
		}
		if !inLetters {
			continue
		}
		if match := taggedH2RE.FindStringSubmatch(line); match != nil {
			if match[2] == "sender" {
				sender = match[1]
			} else {
				flush()
				subject = match[1]
			}
			continue
		}
		if match := taggedH3RE.FindStringSubmatch(line); match != nil {
			switch match[2] {
			case "closing":
				closing = match[1]
			case "signoff":
				signoff = match[1]
			case "to":
				recipients = append(recipients, map[string]string{"address": match[1]})
			}
			continue
		}
		if match := taggedH4RE.FindStringSubmatch(line); match != nil {
			if len(recipients) > 0 {
				recipients[len(recipients)-1][match[2]] = match[1]
			}
			continue
		}
		if match := dateHeadingRE.FindStringSubmatch(line); match != nil {
			date = match[1]
			continue
		}
		if strings.HasPrefix(line, "**") {
			continue
		}
		if subject != "" {
			body += line + "\n"
		}
	}
	flush()
	return letters, nil
}

func Slug(value string) string {
	var output strings.Builder
	dash := false
	for _, r := range strings.ToLower(value) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			if dash && output.Len() > 0 {
				output.WriteByte('-')
			}
			output.WriteRune(r)
			dash = false
		} else {
			dash = true
		}
	}
	return strings.Trim(output.String(), "-")
}

func Validate(letters []Letter, source string) error {
	if len(letters) == 0 {
		return fmt.Errorf("no :letter: tagged sections found in %s", source)
	}
	return nil
}
