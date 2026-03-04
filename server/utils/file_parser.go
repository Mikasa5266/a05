package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/unidoc/unioffice/document"
)

// ExtractTextFromFile 从文件中提取文本内容
func ExtractTextFromFile(file io.Reader, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return extractTextFromPDF(file)
	case ".docx":
		return extractTextFromDOCX(file)
	case ".txt", ".md":
		return extractTextFromPlainText(file)
	default:
		return "", fmt.Errorf("unsupported file format: %s", ext)
	}
}

// extractTextFromPDF 从PDF文件中提取文本
func extractTextFromPDF(file io.Reader) (string, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF file: %w", err)
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "resume-*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // 清理临时文件
	defer tmpFile.Close()

	// 写入内容到临时文件
	if _, err := tmpFile.Write(content); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}

	// 使用ledongthuc/pdf库解析PDF
	// pdf.Open returns (*File, *Reader, error)
	f, r, err := pdf.Open(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer f.Close()

	var textBuilder strings.Builder
	n := r.NumPage()
	for i := 1; i <= n; i++ {
		p := r.Page(i)
		if p.V.IsNull() || p.V.Key("Contents").Kind() == pdf.Null {
			continue
		}

		rows, _ := p.GetTextByRow()
		for _, row := range rows {
			for _, word := range row.Content {
				textBuilder.WriteString(word.S)
				textBuilder.WriteString(" ")
			}
			textBuilder.WriteString("\n")
		}
	}

	return textBuilder.String(), nil
}

// extractTextFromDOCX 从DOCX文件中提取文本
func extractTextFromDOCX(file io.Reader) (string, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read DOCX file: %w", err)
	}

	doc, err := document.Read(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("failed to read DOCX file: %w", err)
	}

	var textBuilder strings.Builder

	for _, para := range doc.Paragraphs() {
		for _, run := range para.Runs() {
			textBuilder.WriteString(run.Text())
		}
		textBuilder.WriteString("\n")
	}

	return textBuilder.String(), nil
}

// extractTextFromPlainText 从纯文本文件中提取文本
func extractTextFromPlainText(file io.Reader) (string, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read text file: %w", err)
	}

	return string(content), nil
}
