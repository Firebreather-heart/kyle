package docxgen

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	docx "github.com/mmonterroca/docxgo/v2"
	"github.com/mmonterroca/docxgo/v2/domain"

	"github.com/firebreather-heart/kyle/internal/models"
)

func GenerateWordDoc(filename string, rawJSON []byte) error {
	var blocks []models.AIBlock
	if err := json.Unmarshal(rawJSON, &blocks); err != nil {
		return fmt.Errorf("failed to parse AI JSON: %v", err)
	}

	primaryColor := domain.Color{R: 0x1F, G: 0x49, B: 0x7D}
	headingFont := "Calibri"
	bodyFont := "Calibri"

	doc := docx.NewDocument()

	for _, block := range blocks {
		switch block.Type {

		case "document_meta":
			if block.PrimaryColor != "" {
				primaryColor = parseHexColor(block.PrimaryColor)
			}
			if block.HeadingFont != "" {
				headingFont = block.HeadingFont
			}
			if block.BodyFont != "" {
				bodyFont = block.BodyFont
			}

		case "h1":
			p, _ := doc.AddParagraph()
			p.SetSpacingBefore(240)
			p.SetSpacingAfter(120)
			r, _ := p.AddRun()
			r.SetText(block.Content)
			r.SetBold(true)
			r.SetFont(domain.Font{Name: headingFont})
			r.SetSize(48)
			r.SetColor(primaryColor)

		case "h2":
			p, _ := doc.AddParagraph()
			p.SetSpacingBefore(200)
			p.SetSpacingAfter(80)
			r, _ := p.AddRun()
			r.SetText(block.Content)
			r.SetBold(true)
			r.SetFont(domain.Font{Name: headingFont})
			r.SetSize(36)
			r.SetColor(primaryColor)

		case "h3":
			p, _ := doc.AddParagraph()
			p.SetSpacingBefore(160)
			p.SetSpacingAfter(60)
			r, _ := p.AddRun()
			r.SetText(block.Content)
			r.SetBold(true)
			r.SetFont(domain.Font{Name: headingFont})
			r.SetSize(28)

		case "paragraph":
			addMarkdownParagraph(doc, block.Content, bodyFont)

		case "callout":
			p, _ := doc.AddParagraph()
			p.SetSpacingBefore(120)
			p.SetSpacingAfter(120)
			border := domain.BorderStyle{Style: domain.BorderSingle, Width: 12, Color: primaryColor}
			p.SetBorderLeft(border)
			p.SetIndent(domain.Indentation{Left: 360})

			iconText := "ℹ  "
			if block.Icon == "warning" {
				iconText = "⚠  "
			} else if block.Icon == "check" {
				iconText = "✓  "
			}

			iconRun, _ := p.AddRun()
			iconRun.SetText(iconText)
			iconRun.SetBold(true)
			iconRun.SetFont(domain.Font{Name: bodyFont})
			iconRun.SetColor(primaryColor)

			addMarkdownRunsTo(p, block.Content, bodyFont)

		case "code_block":
			p, _ := doc.AddParagraph()
			p.SetSpacingBefore(120)
			p.SetSpacingAfter(120)
			p.SetIndent(domain.Indentation{Left: 360, Right: 360})
			border := domain.BorderStyle{Style: domain.BorderSingle, Width: 4, Color: domain.Color{R: 0xCC, G: 0xCC, B: 0xCC}}
			p.SetBorders(domain.ParagraphBorders{Top: border, Bottom: border, Left: border, Right: border})

			lines := strings.Split(block.Content, "\n")
			for i, line := range lines {
				r, _ := p.AddRun()
				r.SetFont(domain.Font{Name: "Courier New"})
				r.SetSize(18)
				r.SetText(line)
				if i < len(lines)-1 {
					r.AddBreak(domain.BreakTypeLine)
				}
			}

		case "table":
			if len(block.Headers) == 0 {
				continue
			}
			numCols := len(block.Headers)
			numRows := 1 + len(block.Rows)
			tbl, err := doc.AddTable(numRows, numCols)
			if err != nil {
				log.Printf("table error: %v", err)
				continue
			}
			tbl.SetStyle(domain.TableStyle{Name: domain.StyleIDTableGrid})

			hRow, _ := tbl.Row(0)
			for i, hdr := range block.Headers {
				cell, _ := hRow.Cell(i)
				cell.SetShading(primaryColor)
				p, _ := cell.AddParagraph()
				r, _ := p.AddRun()
				r.SetText(hdr)
				r.SetBold(true)
				r.SetFont(domain.Font{Name: headingFont})
				r.SetColor(domain.Color{R: 0xFF, G: 0xFF, B: 0xFF})
			}

			for ri, row := range block.Rows {
				dRow, _ := tbl.Row(ri + 1)
				for ci, cellText := range row {
					cell, _ := dRow.Cell(ci)
					p, _ := cell.AddParagraph()
					r, _ := p.AddRun()
					r.SetText(stripMarkdown(cellText))
					r.SetFont(domain.Font{Name: bodyFont})
				}
			}

		case "unordered_list":
			for _, item := range block.Items {
				p, _ := doc.AddParagraph()
				p.SetSpacingAfter(40)
				p.SetIndent(domain.Indentation{Left: 360, Hanging: 240})
				r, _ := p.AddRun()
				r.SetText("• ")
				r.SetFont(domain.Font{Name: bodyFont})
				addMarkdownRunsTo(p, item, bodyFont)
			}

		case "ordered_list":
			for i, item := range block.Items {
				p, _ := doc.AddParagraph()
				p.SetSpacingAfter(40)
				p.SetIndent(domain.Indentation{Left: 360, Hanging: 240})
				r, _ := p.AddRun()
				r.SetText(strconv.Itoa(i+1) + ". ")
				r.SetBold(true)
				r.SetFont(domain.Font{Name: bodyFont})
				addMarkdownRunsTo(p, item, bodyFont)
			}
		}
	}

	if err := doc.SaveAs(filename); err != nil {
		return fmt.Errorf("failed to save docx: %v", err)
	}
	log.Printf("SUCCESS: Word document generated at %s", filename)
	return nil
}

func parseHexColor(hex string) domain.Color {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return domain.Color{}
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return domain.Color{R: uint8(r), G: uint8(g), B: uint8(b)}
}

func addMarkdownParagraph(doc domain.Document, text string, fontName string) {
	p, _ := doc.AddParagraph()
	p.SetSpacingAfter(120)
	addMarkdownRunsTo(p, text, fontName)
}

func addMarkdownRunsTo(p domain.Paragraph, text string, fontName string) {
	parts := strings.Split(text, "**")
	for i, part := range parts {
		if part == "" {
			continue
		}
		r, _ := p.AddRun()
		r.SetFont(domain.Font{Name: fontName})
		r.SetText(part)
		if i%2 != 0 {
			r.SetBold(true)
		}
	}
}

func stripMarkdown(text string) string {
	return strings.ReplaceAll(text, "**", "")
}