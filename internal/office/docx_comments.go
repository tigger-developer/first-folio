// ABOUTME: Converts exactly attributed DOCX footnotes into native Word comments.
// ABOUTME: Preserves ordinary notes and validates package relationships before publication.
package office

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	commentsRelationship         = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/comments"
	commentsExtendedRelationship = "http://schemas.microsoft.com/office/2011/relationships/commentsExtended"
	commentsContentType          = "application/vnd.openxmlformats-officedocument.wordprocessingml.comments+xml"
	commentsExtendedContentType  = "application/vnd.openxmlformats-officedocument.wordprocessingml.commentsExtended+xml"
	word2010Namespace            = "http://schemas.microsoft.com/office/word/2010/wordml"
	word2012Namespace            = "http://schemas.microsoft.com/office/word/2012/wordml"
)

var (
	attributedWithoutTime = regexp.MustCompile(`^\[([^\]\r\n]+)\] wrote: (.*)$`)
	attributedWithTime    = regexp.MustCompile(`^\[([^\]\r\n]+)\] wrote \(([^)\r\n]+)\): (.*)$`)
)

type attributedComment struct {
	footnoteID string
	commentID  int
	author     string
	initials   string
	date       time.Time
	paragraphs []string
}

// ConvertAttributedFootnotes converts the two documented attribution patterns
// and returns the original bytes unchanged when no eligible note exists.
func ConvertAttributedFootnotes(raw []byte, now time.Time) ([]byte, int, error) {
	if err := validatePackage(raw, "docx", defaultPackageLimits()); err != nil {
		return nil, 0, err
	}
	files, err := readPackage(raw, defaultPackageLimits())
	if err != nil {
		return nil, 0, err
	}
	footnoteRaw, ok := files["word/footnotes.xml"]
	if !ok {
		return raw, 0, nil
	}
	footnotes, err := parseXML(footnoteRaw)
	if err != nil {
		return nil, 0, fmt.Errorf("parse DOCX footnotes: %w", err)
	}
	candidates, err := attributedCandidates(footnotes, now.UTC().Truncate(time.Second))
	if err != nil {
		return nil, 0, err
	}
	if len(candidates) == 0 {
		return raw, 0, nil
	}

	document, err := parseXML(files["word/document.xml"])
	if err != nil {
		return nil, 0, fmt.Errorf("parse DOCX document: %w", err)
	}
	comments, err := commentsRoot(files["word/comments.xml"])
	if err != nil {
		return nil, 0, err
	}
	extended, err := commentsExtendedRoot(files["word/commentsExtended.xml"])
	if err != nil {
		return nil, 0, err
	}
	relationships, err := parseRequiredXML(files, "word/_rels/document.xml.rels")
	if err != nil {
		return nil, 0, err
	}
	contentTypes, err := parseRequiredXML(files, "[Content_Types].xml")
	if err != nil {
		return nil, 0, err
	}
	if err := ensureRelationship(relationships, commentsRelationship, "comments.xml"); err != nil {
		return nil, 0, err
	}
	if err := ensureRelationship(relationships, commentsExtendedRelationship, "commentsExtended.xml"); err != nil {
		return nil, 0, err
	}
	if err := ensureContentType(contentTypes, "/word/comments.xml", commentsContentType); err != nil {
		return nil, 0, err
	}
	if err := ensureContentType(contentTypes, "/word/commentsExtended.xml", commentsExtendedContentType); err != nil {
		return nil, 0, err
	}

	comments.set("xmlns:w14", word2010Namespace)
	nextCommentID, err := nextNumericID(comments, "w:comment", "w:id")
	if err != nil {
		return nil, 0, err
	}
	nextParaID, err := nextParagraphID(document, comments, extended)
	if err != nil {
		return nil, 0, err
	}
	existingExtended := map[string]bool{}
	extended.visit(func(n *docxNode) {
		if n.name == "w15:commentEx" {
			existingExtended[n.get("w15:paraId")] = true
		}
	})
	for i := range comments.children {
		comment := &comments.children[i]
		if comment.name != "w:comment" {
			continue
		}
		paragraph := directChild(comment, "w:p")
		if paragraph == nil {
			continue
		}
		paraID := paragraph.get("w14:paraId")
		if paraID == "" {
			paraID = formatParagraphID(nextParaID)
			nextParaID++
			paragraph.set("w14:paraId", paraID)
		}
		if !existingExtended[paraID] {
			extended.children = append(extended.children, commentExtended(paraID))
			existingExtended[paraID] = true
		}
	}
	for _, candidate := range candidates {
		candidate.commentID = nextCommentID
		nextCommentID++
		comment, paraID, next := commentNode(candidate, nextParaID)
		nextParaID = next
		comments.children = append(comments.children, comment)
		extended.children = append(extended.children, commentExtended(paraID))
	}

	byFootnote := map[string]*attributedComment{}
	for _, candidate := range candidates {
		byFootnote[candidate.footnoteID] = candidate
	}
	counts := map[string]int{}
	if err := transformCommentReferences(document, byFootnote, counts); err != nil {
		return nil, 0, err
	}
	for id := range byFootnote {
		if counts[id] != 1 {
			return nil, 0, fmt.Errorf("footnote %s expected 1 document reference, found %d", id, counts[id])
		}
	}
	filtered := footnotes.children[:0]
	for _, child := range footnotes.children {
		if child.name == "w:footnote" && byFootnote[child.get("w:id")] != nil {
			continue
		}
		filtered = append(filtered, child)
	}
	footnotes.children = filtered

	for name, root := range map[string]*docxNode{
		"word/document.xml":            document,
		"word/footnotes.xml":           footnotes,
		"word/comments.xml":            comments,
		"word/commentsExtended.xml":    extended,
		"word/_rels/document.xml.rels": relationships,
		"[Content_Types].xml":          contentTypes,
	} {
		encoded, encodeErr := docxEncode(*root, docxXMLLimit)
		if encodeErr != nil {
			return nil, 0, fmt.Errorf("encode %s: %w", name, encodeErr)
		}
		files[name] = encoded
	}
	converted, err := writePackage(files, "docx")
	if err != nil {
		return nil, 0, err
	}
	if err := validatePackage(converted, "docx", defaultPackageLimits()); err != nil {
		return nil, 0, err
	}
	return converted, len(candidates), nil
}

func attributedCandidates(root *docxNode, now time.Time) ([]*attributedComment, error) {
	var candidates []*attributedComment
	seen := map[string]bool{}
	for i := range root.children {
		note := &root.children[i]
		if note.name != "w:footnote" {
			continue
		}
		id := note.get("w:id")
		numeric, err := strconv.Atoi(id)
		if err != nil || numeric <= 0 {
			continue
		}
		if seen[id] {
			return nil, fmt.Errorf("duplicate DOCX footnote identifier %s", id)
		}
		seen[id] = true
		var paragraphs []string
		for j := range note.children {
			if note.children[j].name == "w:p" {
				paragraphs = append(paragraphs, strings.TrimSpace(wordText(&note.children[j])))
			}
		}
		if len(paragraphs) == 0 {
			continue
		}
		first := paragraphs[0]
		candidate := &attributedComment{footnoteID: id, date: now}
		if match := attributedWithTime.FindStringSubmatch(first); match != nil {
			candidate.author = strings.TrimSpace(match[1])
			candidate.date, err = time.Parse("2006-01-02 15:04:05Z", match[2])
			if err != nil {
				return nil, fmt.Errorf("footnote %s has invalid UTC timestamp %q", id, match[2])
			}
			candidate.paragraphs = append([]string{match[3]}, paragraphs[1:]...)
		} else if match := attributedWithoutTime.FindStringSubmatch(first); match != nil {
			candidate.author = strings.TrimSpace(match[1])
			candidate.paragraphs = append([]string{match[2]}, paragraphs[1:]...)
		} else {
			continue
		}
		if candidate.author == "" {
			continue
		}
		candidate.initials = authorInitials(candidate.author)
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

func wordText(root *docxNode) string {
	var out strings.Builder
	root.visit(func(n *docxNode) {
		if n.name == "w:t" {
			out.WriteString(nodeCharacterData(n))
		}
	})
	return out.String()
}

func nodeCharacterData(root *docxNode) string {
	var out strings.Builder
	if root.name == "" {
		out.WriteString(root.text)
	}
	for i := range root.children {
		out.WriteString(nodeCharacterData(&root.children[i]))
	}
	return out.String()
}

func authorInitials(author string) string {
	var out strings.Builder
	for _, field := range strings.Fields(author) {
		r, _ := utf8.DecodeRuneInString(field)
		if r != utf8.RuneError {
			out.WriteRune(r)
		}
	}
	return strings.ToUpper(out.String())
}

func commentsRoot(raw []byte) (*docxNode, error) {
	if len(raw) == 0 {
		root := dxRoot("comments")
		return &root, nil
	}
	root, err := parseXML(raw)
	if err != nil {
		return nil, fmt.Errorf("parse DOCX comments: %w", err)
	}
	if root.name != "w:comments" {
		return nil, fmt.Errorf("parse DOCX comments: unexpected root %s", root.name)
	}
	return root, nil
}

func commentsExtendedRoot(raw []byte) (*docxNode, error) {
	if len(raw) == 0 {
		root := docxNode{name: "w15:commentsEx"}
		root.set("xmlns:w15", word2012Namespace)
		return &root, nil
	}
	root, err := parseXML(raw)
	if err != nil {
		return nil, fmt.Errorf("parse DOCX extended comments: %w", err)
	}
	if root.name != "w15:commentsEx" {
		return nil, fmt.Errorf("parse DOCX extended comments: unexpected root %s", root.name)
	}
	return root, nil
}

func parseRequiredXML(files map[string][]byte, name string) (*docxNode, error) {
	raw, ok := files[name]
	if !ok {
		return nil, fmt.Errorf("missing office member %s", name)
	}
	root, err := parseXML(raw)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}
	return root, nil
}

func ensureRelationship(root *docxNode, relationshipType, target string) error {
	matches := 0
	for i := range root.children {
		rel := &root.children[i]
		if rel.name != "Relationship" {
			continue
		}
		typ, gotTarget := rel.get("Type"), rel.get("Target")
		if typ == relationshipType || gotTarget == target {
			if typ != relationshipType || gotTarget != target {
				return fmt.Errorf("DOCX comments relationship conflict for %s", target)
			}
			matches++
		}
	}
	if matches > 1 {
		return fmt.Errorf("duplicate DOCX comments relationship for %s", target)
	}
	if matches == 1 {
		return nil
	}
	root.children = append(root.children, docxNode{name: "Relationship"}.attr("Id", nextRelationshipID(root)).attr("Type", relationshipType).attr("Target", target))
	return nil
}

func nextRelationshipID(root *docxNode) string {
	used := map[string]bool{}
	for i := range root.children {
		if root.children[i].name == "Relationship" {
			used[root.children[i].get("Id")] = true
		}
	}
	for i := 1; ; i++ {
		id := fmt.Sprintf("rId%d", i)
		if !used[id] {
			return id
		}
	}
}

func ensureContentType(root *docxNode, partName, contentType string) error {
	matches := 0
	for i := range root.children {
		override := &root.children[i]
		if override.name != "Override" || override.get("PartName") != partName {
			continue
		}
		if override.get("ContentType") != contentType {
			return fmt.Errorf("DOCX comments content type conflict for %s", partName)
		}
		matches++
	}
	if matches > 1 {
		return fmt.Errorf("duplicate DOCX comments content type for %s", partName)
	}
	if matches == 1 {
		return nil
	}
	root.children = append(root.children, docxNode{name: "Override"}.attr("PartName", partName).attr("ContentType", contentType))
	return nil
}

func nextNumericID(root *docxNode, element, attribute string) (int, error) {
	maxID := -1
	var parseErr error
	seen := map[int]bool{}
	root.visit(func(n *docxNode) {
		if parseErr != nil || n.name != element {
			return
		}
		value, err := strconv.Atoi(n.get(attribute))
		if err != nil || value < 0 {
			parseErr = fmt.Errorf("invalid %s %q", attribute, n.get(attribute))
			return
		}
		if seen[value] {
			parseErr = fmt.Errorf("duplicate %s %d", attribute, value)
			return
		}
		seen[value] = true
		if value > maxID {
			maxID = value
		}
	})
	return maxID + 1, parseErr
}

func nextParagraphID(roots ...*docxNode) (uint64, error) {
	var maxID uint64
	var parseErr error
	for _, root := range roots {
		root.visit(func(n *docxNode) {
			if parseErr != nil {
				return
			}
			for _, key := range []string{"w14:paraId", "w15:paraId"} {
				if value := n.get(key); value != "" {
					parsed, err := strconv.ParseUint(value, 16, 32)
					if err != nil {
						parseErr = fmt.Errorf("invalid DOCX comment paragraph identifier %q", value)
						return
					}
					if parsed > maxID {
						maxID = parsed
					}
				}
			}
		})
	}
	if parseErr != nil {
		return 0, parseErr
	}
	if maxID >= 0xffffffff {
		return 0, fmt.Errorf("DOCX comment paragraph identifier range exhausted")
	}
	return maxID + 1, nil
}

func formatParagraphID(id uint64) string { return fmt.Sprintf("%08X", id) }

func directChild(root *docxNode, name string) *docxNode {
	for i := range root.children {
		if root.children[i].name == name {
			return &root.children[i]
		}
	}
	return nil
}

func commentNode(candidate *attributedComment, nextParaID uint64) (docxNode, string, uint64) {
	comment := dx("comment").attr("w:id", strconv.Itoa(candidate.commentID)).attr("w:author", candidate.author).attr("w:initials", candidate.initials).attr("w:date", candidate.date.UTC().Format(time.RFC3339))
	firstParaID := ""
	for i, text := range candidate.paragraphs {
		paraID := formatParagraphID(nextParaID)
		nextParaID++
		if firstParaID == "" {
			firstParaID = paraID
		}
		paragraph := dx("p", dx("pPr", dxVal("pStyle", "CommentText"))).attr("w14:paraId", paraID)
		if i == 0 {
			paragraph.children = append(paragraph.children, dx("r", dx("rPr", dxVal("rStyle", "CommentReference")), dx("annotationRef")))
		}
		paragraph.children = append(paragraph.children, docxRun(text, dx("rPr")))
		comment.children = append(comment.children, paragraph)
	}
	return comment, firstParaID, nextParaID
}

func commentExtended(paraID string) docxNode {
	return docxNode{name: "w15:commentEx"}.attr("w15:paraId", paraID).attr("w15:done", "0")
}

func transformCommentReferences(root *docxNode, candidates map[string]*attributedComment, counts map[string]int) error {
	if root.name == "w:p" {
		out := make([]docxNode, 0, len(root.children)+4)
		for _, child := range root.children {
			if child.name != "w:r" {
				out = append(out, child)
				continue
			}
			var id string
			references := 0
			unexpected := false
			for i := range child.children {
				switch child.children[i].name {
				case "w:footnoteReference":
					id = child.children[i].get("w:id")
					references++
				case "w:rPr":
				case "":
					unexpected = unexpected || strings.TrimSpace(child.children[i].text) != ""
				default:
					unexpected = true
				}
			}
			candidate := candidates[id]
			if candidate == nil {
				out = append(out, child)
				continue
			}
			if references != 1 || unexpected {
				return fmt.Errorf("footnote %s reference run contains unsupported mixed content", id)
			}
			counts[id]++
			commentID := strconv.Itoa(candidate.commentID)
			out = append(out,
				dx("commentRangeStart").attr("w:id", commentID),
				dx("commentRangeEnd").attr("w:id", commentID),
				dx("r", dx("rPr", dxVal("rStyle", "CommentReference")), dx("commentReference").attr("w:id", commentID)),
			)
		}
		root.children = out
		return nil
	}
	for i := range root.children {
		if err := transformCommentReferences(&root.children[i], candidates, counts); err != nil {
			return err
		}
	}
	return nil
}
