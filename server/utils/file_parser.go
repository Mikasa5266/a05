package utils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"your-project/config"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/unidoc/unioffice/document"
)

// findTesseract 尝试在系统中定位 tesseract 可执行文件与 tessdata 目录
func findTesseract() (exePath string, tessdata string) {
	// 0) Config (Priority)
	if cfg := config.GetConfig(); cfg != nil {
		if cfg.OCR.TesseractPath != "" {
			if st, err := os.Stat(cfg.OCR.TesseractPath); err == nil && !st.IsDir() {
				exePath = cfg.OCR.TesseractPath
			}
		}
		if cfg.OCR.TessdataPath != "" {
			if st, err := os.Stat(cfg.OCR.TessdataPath); err == nil && st.IsDir() {
				tessdata = cfg.OCR.TessdataPath
			}
		}
		// If both found from config, return immediately
		if exePath != "" && tessdata != "" {
			return
		}
	}

	// 1) PATH
	if exePath == "" {
		if p, _ := exec.LookPath("tesseract"); p != "" {
			exePath = p
		}
	}

	// Try to derive tessdata if not set
	if exePath != "" && tessdata == "" {
		dir := filepath.Dir(exePath)
		td := filepath.Join(dir, "tessdata")
		if st, err := os.Stat(td); err == nil && st.IsDir() {
			tessdata = td
		}
	}

	if exePath != "" && tessdata != "" {
		return
	}

	// 2) Environment Variables
	if exePath == "" {
		if v := os.Getenv("TESSERACT_PATH"); v != "" {
			if st, err := os.Stat(v); err == nil && !st.IsDir() {
				exePath = v
			}
		}
	}
	if tessdata == "" {
		if v := os.Getenv("TESSDATA_PREFIX"); v != "" {
			if st, err := os.Stat(v); err == nil && st.IsDir() {
				tessdata = v
			}
		}
	}

	// 3) Common Paths fallback
	if exePath == "" {
		candidates := []string{
			`C:\Program Files\Tesseract-OCR\tesseract.exe`,
			`C:\Program Files (x86)\Tesseract-OCR\tesseract.exe`,
			`D:\Program Files\Tesseract-OCR\tesseract.exe`,
			`D:\resource\tesseract\tesseract.exe`,
		}
		for _, c := range candidates {
			if st, err := os.Stat(c); err == nil && !st.IsDir() {
				exePath = c
				break
			}
		}
	}

	// Final attempt to find tessdata if we found exe but not data
	if exePath != "" && tessdata == "" {
		dir := filepath.Dir(exePath)
		td := filepath.Join(dir, "tessdata")
		if st, err := os.Stat(td); err == nil && st.IsDir() {
			tessdata = td
		}
	}

	return
}

// findPdftoppm locates pdftoppm executable
func findPdftoppm() string {
	// 0) Config
	if cfg := config.GetConfig(); cfg != nil {
		if cfg.OCR.PdftoppmPath != "" {
			if st, err := os.Stat(cfg.OCR.PdftoppmPath); err == nil && !st.IsDir() {
				return cfg.OCR.PdftoppmPath
			}
		}
	}
	// 1) PATH
	if p, err := exec.LookPath("pdftoppm"); err == nil {
		return p
	}
	return ""
}

// renderWithPdftoppm 使用 pdftoppm 将整页渲染为图片
func renderWithPdftoppm(pdfPath, outDir string) error {
	exe := findPdftoppm()
	if exe == "" {
		return fmt.Errorf("pdftoppm not found")
	}
	// 输出到 outDir/page-*.png
	prefix := filepath.Join(outDir, "page")
	cmd := exec.Command(exe, "-r", "300", "-png", pdfPath, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pdftoppm error: %v, stderr: %s", err, string(out))
	}
	return nil
}

// OCRStatus 返回 OCR 依赖状态
func OCRStatus() (string, string, []string, bool) {
	exe, tess := findTesseract()
	langs := []string{}
	if exe != "" {
		cmd := exec.Command(exe, "--list-langs")
		if tess != "" {
			cmd.Env = append(os.Environ(), "TESSDATA_PREFIX="+tess)
		}
		if out, err := cmd.CombinedOutput(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, ln := range lines {
				s := strings.TrimSpace(ln)
				if s == "" || strings.Contains(s, "List of available languages") {
					continue
				}
				langs = append(langs, s)
			}
		}
	}

	ppm := findPdftoppm()
	return exe, tess, langs, ppm != ""
}

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

	primary := strings.TrimSpace(textBuilder.String())
	if len([]rune(primary)) >= 50 {
		return primary, nil
	}

	log.Printf("Primary PDF text extraction is too short, trying OCR fallback")
	var ocrBuilder strings.Builder
	conf := model.NewDefaultConfiguration()
	imgOutDir, err := os.MkdirTemp("", "resume-imgs-*")
	if err == nil {
		defer os.RemoveAll(imgOutDir)
		// 先尝试直接提取内嵌图片
		extractErr := api.ExtractImagesFile(tmpFile.Name(), imgOutDir, nil, conf)
		entries, _ := os.ReadDir(imgOutDir)
		// 如果没有内嵌图片，尝试整页渲染
		if extractErr != nil || len(entries) == 0 {
			log.Printf("pdfcpu extracted %d images, trying pdftoppm page rendering", len(entries))
			if err := renderWithPdftoppm(tmpFile.Name(), imgOutDir); err != nil {
				log.Printf("pdftoppm rendering not available: %v", err)
			}
			entries, _ = os.ReadDir(imgOutDir)
		}

		tesPath, tessdata := findTesseract()
		if tesPath == "" {
			log.Printf("tesseract not found in PATH; skip OCR fallback")
		}

		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			l := strings.ToLower(name)
			if !(strings.HasSuffix(l, ".jpg") || strings.HasSuffix(l, ".jpeg") || strings.HasSuffix(l, ".png")) {
				continue
			}
			if tesPath == "" {
				break
			}
			imgPath := filepath.Join(imgOutDir, name)
			cmd := exec.Command(tesPath, imgPath, "stdout", "-l", "chi_sim+eng", "--psm", "3")
			if tessdata != "" {
				cmd.Env = append(os.Environ(), "TESSDATA_PREFIX="+tessdata)
			}
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("tesseract chi_sim+eng failed: %v, stderr: %s", err, string(out))
				// 尝试仅英文
				cmd2 := exec.Command(tesPath, imgPath, "stdout", "-l", "eng", "--psm", "3")
				if tessdata != "" {
					cmd2.Env = append(os.Environ(), "TESSDATA_PREFIX="+tessdata)
				}
				out2, err2 := cmd2.CombinedOutput()
				if err2 != nil {
					log.Printf("tesseract eng fallback failed: %v, stderr: %s", err2, string(out2))
					continue
				}
				ocrBuilder.WriteString(string(out2))
				ocrBuilder.WriteString("\n")
				continue
			}
			ocrBuilder.WriteString(string(out))
			ocrBuilder.WriteString("\n")
		}
	}

	ocrText := strings.TrimSpace(ocrBuilder.String())
	if ocrText != "" && len([]rune(ocrText)) > len([]rune(primary)) {
		return ocrText, nil
	}
	return primary, nil
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
