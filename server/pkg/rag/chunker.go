package rag

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

const (
	defaultChunkSize    = 500
	defaultChunkOverlap = 50
)

var markdownHeadingPattern = regexp.MustCompile(`^(#{1,2})\s+(.+)$`)

// SlidingWindowChunker implements DocumentSplitter with markdown-aware segmentation
// and sliding-window fallback for long paragraphs.
type SlidingWindowChunker struct {
	chunkSize    int
	chunkOverlap int
}

type semanticSegment struct {
	content  string
	metadata map[string]string
}

func NewSlidingWindowChunker(chunkSize, chunkOverlap int) (*SlidingWindowChunker, error) {
	if chunkSize <= 0 {
		chunkSize = defaultChunkSize
	}
	if chunkOverlap < 0 {
		chunkOverlap = defaultChunkOverlap
	}
	if chunkOverlap >= chunkSize {
		return nil, fmt.Errorf("chunkOverlap (%d) must be smaller than chunkSize (%d)", chunkOverlap, chunkSize)
	}

	return &SlidingWindowChunker{
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
	}, nil
}

func NewDefaultSlidingWindowChunker() *SlidingWindowChunker {
	chunker, _ := NewSlidingWindowChunker(defaultChunkSize, defaultChunkOverlap)
	return chunker
}

func (c *SlidingWindowChunker) Split(ctx context.Context, doc Document) ([]Chunk, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if c == nil {
		return nil, fmt.Errorf("chunker is nil")
	}
	if c.chunkSize <= 0 {
		return nil, fmt.Errorf("invalid chunkSize: %d", c.chunkSize)
	}
	if c.chunkOverlap < 0 || c.chunkOverlap >= c.chunkSize {
		return nil, fmt.Errorf("invalid chunkOverlap: %d", c.chunkOverlap)
	}

	content := strings.TrimSpace(doc.Content)
	if content == "" {
		return nil, fmt.Errorf("document content is empty")
	}

	docID := strings.TrimSpace(doc.ID)
	if docID == "" {
		docID = "doc"
	}

	segments := c.buildSemanticSegments(doc, content)
	if len(segments) == 0 {
		return nil, fmt.Errorf("no semantic segments produced for document")
	}

	chunks := make([]Chunk, 0, len(segments))
	chunkIndex := 0
	for _, seg := range segments {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		segmentText := strings.TrimSpace(seg.content)
		if segmentText == "" {
			continue
		}

		segmentChunks, err := c.splitSegment(ctx, segmentText)
		if err != nil {
			return nil, fmt.Errorf("split segment failed: %w", err)
		}

		for _, part := range segmentChunks {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			chunkMetadata := cloneMetadata(doc.Metadata)
			for k, v := range seg.metadata {
				chunkMetadata[k] = v
			}
			chunkMetadata["chunk_index"] = fmt.Sprintf("%d", chunkIndex)
			chunkMetadata["chunk_size"] = fmt.Sprintf("%d", len([]rune(part)))

			chunks = append(chunks, Chunk{
				ID:       fmt.Sprintf("%s_%d", docID, chunkIndex),
				Content:  part,
				Metadata: chunkMetadata,
			})
			chunkIndex++
		}
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks produced from document")
	}

	return chunks, nil
}

func (c *SlidingWindowChunker) buildSemanticSegments(doc Document, content string) []semanticSegment {
	if c.isMarkdown(doc, content) {
		segments := c.splitMarkdownByHeading(content)
		if len(segments) > 0 {
			return segments
		}
	}

	return []semanticSegment{{
		content:  content,
		metadata: map[string]string{"segment_type": "plain"},
	}}
}

func (c *SlidingWindowChunker) isMarkdown(doc Document, content string) bool {
	if doc.Metadata != nil {
		source := strings.ToLower(strings.TrimSpace(doc.Metadata["source"]))
		if strings.HasSuffix(source, ".md") || strings.HasSuffix(source, ".markdown") {
			return true
		}
	}

	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false
	}

	if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## ") {
		return true
	}
	if strings.Contains(trimmed, "\n# ") || strings.Contains(trimmed, "\n## ") {
		return true
	}
	return false
}

func (c *SlidingWindowChunker) splitMarkdownByHeading(content string) []semanticSegment {
	lines := strings.Split(content, "\n")
	segments := make([]semanticSegment, 0)
	currentHeading := ""
	currentLevel := "0"
	buffer := make([]string, 0, 16)

	flush := func() {
		joined := strings.TrimSpace(strings.Join(buffer, "\n"))
		if joined == "" {
			return
		}
		metadata := map[string]string{
			"segment_type":  "markdown",
			"heading":       strings.TrimSpace(currentHeading),
			"heading_level": currentLevel,
		}
		segments = append(segments, semanticSegment{content: joined, metadata: metadata})
	}

	for _, rawLine := range lines {
		line := strings.TrimRight(rawLine, "\r")
		trimmed := strings.TrimSpace(line)
		if matches := markdownHeadingPattern.FindStringSubmatch(trimmed); len(matches) == 3 {
			flush()
			buffer = buffer[:0]
			currentHeading = strings.TrimSpace(matches[2])
			currentLevel = fmt.Sprintf("%d", len(matches[1]))
			buffer = append(buffer, trimmed)
			continue
		}
		buffer = append(buffer, line)
	}
	flush()

	for i := range segments {
		segments[i].metadata["segment_index"] = fmt.Sprintf("%d", i)
	}
	return segments
}

func (c *SlidingWindowChunker) splitSegment(ctx context.Context, segment string) ([]string, error) {
	segment = strings.TrimSpace(segment)
	if segment == "" {
		return nil, nil
	}

	if len([]rune(segment)) <= c.chunkSize {
		return []string{segment}, nil
	}

	paragraphs := splitParagraphs(segment)
	if len(paragraphs) == 0 {
		return c.splitBySlidingWindow(ctx, segment), nil
	}

	chunks := make([]string, 0, len(paragraphs))
	var current strings.Builder

	appendCurrent := func() {
		text := strings.TrimSpace(current.String())
		if text != "" {
			chunks = append(chunks, text)
		}
		current.Reset()
	}

	for _, p := range paragraphs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		paragraph := strings.TrimSpace(p)
		if paragraph == "" {
			continue
		}

		paragraphLen := len([]rune(paragraph))
		if paragraphLen > c.chunkSize {
			appendCurrent()
			chunks = append(chunks, c.splitBySlidingWindow(ctx, paragraph)...)
			continue
		}

		if current.Len() == 0 {
			current.WriteString(paragraph)
			continue
		}

		candidate := current.String() + "\n\n" + paragraph
		if len([]rune(candidate)) <= c.chunkSize {
			current.WriteString("\n\n")
			current.WriteString(paragraph)
			continue
		}

		appendCurrent()
		current.WriteString(paragraph)
	}

	appendCurrent()
	if len(chunks) == 0 {
		return c.splitBySlidingWindow(ctx, segment), nil
	}

	return c.normalizeChunkOrder(chunks), nil
}

func (c *SlidingWindowChunker) splitBySlidingWindow(ctx context.Context, text string) []string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) == 0 {
		return nil
	}
	if len(runes) <= c.chunkSize {
		return []string{string(runes)}
	}

	step := c.chunkSize - c.chunkOverlap
	if step <= 0 {
		step = c.chunkSize
	}

	chunks := make([]string, 0, (len(runes)/step)+1)
	for start := 0; start < len(runes); start += step {
		if err := ctx.Err(); err != nil {
			break
		}

		end := start + c.chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		piece := strings.TrimSpace(string(runes[start:end]))
		if piece != "" {
			chunks = append(chunks, piece)
		}

		if end == len(runes) {
			break
		}
	}

	return c.normalizeChunkOrder(chunks)
}

func (c *SlidingWindowChunker) normalizeChunkOrder(chunks []string) []string {
	if len(chunks) <= 1 {
		return chunks
	}

	cleaned := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		cleaned = append(cleaned, chunk)
	}

	if len(cleaned) <= 1 {
		return cleaned
	}
	return cleaned
}

func splitParagraphs(text string) []string {
	parts := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func cloneMetadata(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
