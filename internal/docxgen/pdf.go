package docxgen

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/firebreather-heart/kyle/internal/models"
)

const (
	pdfMargin   = 22.0
	pdfContentW = 210.0 - 2*pdfMargin // A4 width minus margins
	bodySize    = 10.5
	lineH       = 6.0
	codeLineH   = 5.0
)

func GeneratePDF(filename string, rawJSON []byte) error {
	var blocks []models.AIBlock
	if err := json.Unmarshal(rawJSON, &blocks); err != nil {
		return fmt.Errorf("failed to parse AI JSON: %v", err)
	}

	pr, pg, pb := 31, 73, 125 // default primary colour: #1F497D

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	pdf.SetAutoPageBreak(true, pdfMargin)
	pdf.AddPage()

	for _, block := range blocks {
		switch block.Type {

		case "document_meta":
			if block.PrimaryColor != "" {
				pr, pg, pb = hexToRGB(block.PrimaryColor)
			}

		case "h1":
			pdf.Ln(4)
			pdf.SetFont("Helvetica", "B", 22)
			pdf.SetTextColor(pr, pg, pb)
			pdf.MultiCell(pdfContentW, 9, block.Content, "", "L", false)
			y := pdf.GetY()
			pdf.SetDrawColor(pr, pg, pb)
			pdf.SetLineWidth(0.5)
			pdf.Line(pdfMargin, y, pdfMargin+pdfContentW, y)
			pdf.Ln(4)
			pdf.SetTextColor(0, 0, 0)

		case "h2":
			pdf.Ln(5)
			pdf.SetFont("Helvetica", "B", 16)
			pdf.SetTextColor(pr, pg, pb)
			pdf.MultiCell(pdfContentW, 7, block.Content, "", "L", false)
			pdf.Ln(2)
			pdf.SetTextColor(0, 0, 0)

		case "h3":
			pdf.Ln(4)
			pdf.SetFont("Helvetica", "B", 13)
			pdf.SetTextColor(60, 60, 60)
			pdf.MultiCell(pdfContentW, 6, block.Content, "", "L", false)
			pdf.Ln(2)
			pdf.SetTextColor(0, 0, 0)

		case "paragraph":
			pdf.Ln(2)
			writeMixed(pdf, block.Content, bodySize, lineH)
			pdf.Ln(3)

		case "callout":
			pdf.Ln(3)
			startY := pdf.GetY()
			icon := "i  "
			if block.Icon == "warning" {
				icon = "!  "
			} else if block.Icon == "check" {
				icon = "v  "
			}
			pdf.SetX(pdfMargin + 5)
			pdf.SetFont("Helvetica", "B", bodySize)
			pdf.SetTextColor(pr, pg, pb)
			pdf.Write(lineH, icon)
			pdf.SetTextColor(50, 50, 50)
			writeMixedFrom(pdf, block.Content, bodySize, lineH, pdfMargin+5)
			endY := pdf.GetY() + lineH
			pdf.SetDrawColor(pr, pg, pb)
			pdf.SetLineWidth(1.2)
			pdf.Line(pdfMargin+1.5, startY, pdfMargin+1.5, endY)
			pdf.SetTextColor(0, 0, 0)
			pdf.Ln(4)

		case "code_block":
			pdf.Ln(3)
			lines := strings.Split(block.Content, "\n")
			blockH := float64(len(lines))*codeLineH + 6.0
			x := pdfMargin
			y := pdf.GetY()
			pdf.SetFillColor(245, 245, 245)
			pdf.SetDrawColor(200, 200, 200)
			pdf.SetLineWidth(0.3)
			pdf.Rect(x, y, pdfContentW, blockH, "FD")
			pdf.SetFont("Courier", "", 9)
			pdf.SetTextColor(40, 40, 40)
			pdf.SetY(y + 3)
			for _, line := range lines {
				pdf.SetX(x + 4)
				pdf.CellFormat(pdfContentW-8, codeLineH, line, "", 1, "L", false, 0, "")
			}
			pdf.SetTextColor(0, 0, 0)
			pdf.Ln(3)

		case "table":
			if len(block.Headers) == 0 {
				continue
			}
			pdf.Ln(4)
			n := len(block.Headers)
			colW := pdfContentW / float64(n)
			pdf.SetFont("Helvetica", "B", 9)
			pdf.SetFillColor(pr, pg, pb)
			pdf.SetTextColor(255, 255, 255)
			for _, h := range block.Headers {
				pdf.CellFormat(colW, 7, h, "1", 0, "C", true, 0, "")
			}
			pdf.Ln(-1)
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(30, 30, 30)
			for ri, row := range block.Rows {
				if ri%2 == 0 {
					pdf.SetFillColor(248, 248, 248)
				} else {
					pdf.SetFillColor(255, 255, 255)
				}
				for ci, cell := range row {
					if ci >= n {
						break
					}
					pdf.CellFormat(colW, 6, stripMarkdown(cell), "1", 0, "L", true, 0, "")
				}
				pdf.Ln(-1)
			}
			pdf.SetTextColor(0, 0, 0)
			pdf.Ln(4)

		case "unordered_list":
			pdf.Ln(2)
			for _, item := range block.Items {
				pdf.SetX(pdfMargin + 4)
				pdf.SetFont("Helvetica", "", bodySize)
				pdf.Write(lineH, "\x95 ") // bullet
				writeMixedFrom(pdf, item, bodySize, lineH, pdfMargin+4)
				pdf.Ln(1)
			}
			pdf.Ln(2)

		case "ordered_list":
			pdf.Ln(2)
			for i, item := range block.Items {
				pdf.SetX(pdfMargin + 4)
				pdf.SetFont("Helvetica", "B", bodySize)
				pdf.Write(lineH, strconv.Itoa(i+1)+". ")
				writeMixedFrom(pdf, item, bodySize, lineH, pdfMargin+4)
				pdf.Ln(1)
			}
			pdf.Ln(2)
		}
	}

	return pdf.OutputFileAndClose(filename)
}

// writeMixed writes **bold** markdown inline starting at the left margin.
func writeMixed(pdf *fpdf.Fpdf, text string, size, h float64) {
	pdf.SetX(pdfMargin)
	writeMixedFrom(pdf, text, size, h, pdfMargin)
}

// writeMixedFrom writes **bold** markdown inline continuing from the current position.
func writeMixedFrom(pdf *fpdf.Fpdf, text string, size, h float64, _ float64) {
	parts := strings.Split(text, "**")
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i%2 != 0 {
			pdf.SetFont("Helvetica", "B", size)
		} else {
			pdf.SetFont("Helvetica", "", size)
		}
		pdf.Write(h, part)
	}
}

func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 31, 73, 125
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return int(r), int(g), int(b)
}
