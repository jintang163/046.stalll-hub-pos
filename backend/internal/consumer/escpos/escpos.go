package escpos

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

type Printer struct {
	buf *bytes.Buffer
}

func NewPrinter() *Printer {
	return &Printer{
		buf: bytes.NewBuffer(nil),
	}
}

func (p *Printer) Bytes() []byte {
	return p.buf.Bytes()
}

func (p *Printer) Reset() {
	p.buf.Reset()
}

func (p *Printer) write(b []byte) {
	p.buf.Write(b)
}

func (p *Printer) Initialize() {
	p.write([]byte{0x1B, 0x40})
}

func (p *Printer) Cut() {
	p.write([]byte{0x1D, 0x56, 0x00})
}

func (p *Printer) CutPartial() {
	p.write([]byte{0x1D, 0x56, 0x01})
}

func (p *Printer) Feed(n int) {
	p.write([]byte{0x1B, 0x64, byte(n)})
}

func (p *Printer) NewLine() {
	p.write([]byte{0x0A})
}

func (p *Printer) SetTextNormal() {
	p.write([]byte{0x1B, 0x21, 0x00})
}

func (p *Printer) SetTextBold(enable bool) {
	if enable {
		p.write([]byte{0x1B, 0x45, 0x01})
	} else {
		p.write([]byte{0x1B, 0x45, 0x00})
	}
}

func (p *Printer) SetTextSize(width, height int) {
	w := byte(width - 1)
	h := byte(height - 1)
	if w > 7 {
		w = 7
	}
	if h > 7 {
		h = 7
	}
	p.write([]byte{0x1D, 0x21, (w << 4) | h})
}

func (p *Printer) SetTextDoubleHeight(enable bool) {
	if enable {
		p.write([]byte{0x1B, 0x21, 0x10})
	} else {
		p.write([]byte{0x1B, 0x21, 0x00})
	}
}

func (p *Printer) SetTextDoubleWidth(enable bool) {
	if enable {
		p.write([]byte{0x1B, 0x21, 0x20})
	} else {
		p.write([]byte{0x1B, 0x21, 0x00})
	}
}

func (p *Printer) SetUnderline(enable bool) {
	if enable {
		p.write([]byte{0x1B, 0x2D, 0x01})
	} else {
		p.write([]byte{0x1B, 0x2D, 0x00})
	}
}

func (p *Printer) SetAlign(align string) {
	var b byte
	switch strings.ToLower(align) {
	case "center":
		b = 0x01
	case "right":
		b = 0x02
	default:
		b = 0x00
	}
	p.write([]byte{0x1B, 0x61, b})
}

func (p *Printer) SetFont(font string) {
	var b byte
	switch strings.ToLower(font) {
	case "b":
		b = 0x01
	default:
		b = 0x00
	}
	p.write([]byte{0x1B, 0x4D, b})
}

func (p *Printer) PrintText(text string) {
	p.write([]byte(text))
}

func (p *Printer) PrintLine(text string) {
	p.PrintText(text)
	p.NewLine()
}

func (p *Printer) PrintLineWithAlign(text string, align string) {
	p.SetAlign(align)
	p.PrintLine(text)
	p.SetAlign("left")
}

func (p *Printer) PrintTitle(text string) {
	p.SetTextBold(true)
	p.SetTextSize(2, 2)
	p.SetAlign("center")
	p.PrintLine(text)
	p.SetTextNormal()
	p.Feed(1)
}

func (p *Printer) PrintSeparator() {
	p.PrintLine("--------------------------------")
}

func (p *Printer) PrintDoubleSeparator() {
	p.PrintLine("================================")
}

func (p *Printer) PrintItem(name string, quantity int, price string, amount string) {
	line := fmt.Sprintf("%-20s %3d %6s %8s", name, quantity, price, amount)
	p.PrintLine(line)
}

func (p *Printer) PrintItemWithWidth(name string, quantity int, price string, amount string, nameWidth int) {
	format := fmt.Sprintf("%%-%ds %%3d %%6s %%8s", nameWidth)
	line := fmt.Sprintf(format, name, quantity, price, amount)
	p.PrintLine(line)
}

func (p *Printer) PrintSummary(label string, value string) {
	line := fmt.Sprintf("%-20s %14s", label, value)
	p.SetTextBold(true)
	p.PrintLine(line)
	p.SetTextBold(false)
}

func (p *Printer) PrintHeader(storeName, orderNo string) {
	p.Initialize()
	p.Feed(2)
	p.PrintTitle(storeName)
	p.PrintLineWithAlign("--------------------------------", "center")
	p.SetAlign("left")
	p.PrintLine(fmt.Sprintf("订单号: %s", orderNo))
	p.PrintLine(fmt.Sprintf("时间: %s", time.Now().Format("2006-01-02 15:04:05")))
	p.PrintSeparator()
	p.Feed(1)
}

func (p *Printer) PrintFooter(footer string) {
	p.Feed(1)
	p.PrintSeparator()
	if footer != "" {
		p.SetAlign("center")
		p.PrintLine(footer)
		p.SetAlign("left")
	}
	p.Feed(3)
	p.Cut()
}

func (p *Printer) PrintBarcode(code string, barcodeType byte) {
	p.write([]byte{0x1D, 0x6B, barcodeType})
	p.write([]byte(code))
	p.write([]byte{0x00})
}

func (p *Printer) PrintQRCode(code string, size byte) {
	if size < 1 {
		size = 8
	}
	if size > 16 {
		size = 16
	}
	p.write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x43, size})
	p.write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x45, 0x30})
	codeLen := len(code) + 3
	pl := byte(codeLen % 256)
	ph := byte(codeLen / 256)
	p.write([]byte{0x1D, 0x28, 0x6B, pl, ph, 0x31, 0x50, 0x30})
	p.write([]byte(code))
	p.write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x51, 0x30})
}

func (p *Printer) OpenCashDrawer() {
	p.write([]byte{0x1B, 0x70, 0x00, 0x3C, 0x78})
}
