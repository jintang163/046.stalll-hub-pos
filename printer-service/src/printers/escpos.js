'use strict';

const iconv = require('iconv-lite');

class EscPosPrinter {
  constructor() {
    this.buffer = Buffer.alloc(0);
    this.encoding = 'UTF-8';
  }

  raw(data) {
    if (typeof data === 'string') {
      data = iconv.encode(data, this.encoding);
    }
    this.buffer = Buffer.concat([this.buffer, data]);
    return this;
  }

  esc(cmd, ...args) {
    this.buffer = Buffer.concat([this.buffer, Buffer.from([0x1b, cmd, ...args])]);
    return this;
  }

  gs(cmd, ...args) {
    this.buffer = Buffer.concat([this.buffer, Buffer.from([0x1d, cmd, ...args])]);
    return this;
  }

  setEncoding(encoding) {
    this.encoding = encoding;
    return this;
  }

  setTextBold(bold = true) {
    return this.esc(0x45, bold ? 0x01 : 0x00);
  }

  setTextSize(width = 1, height = 1) {
    return this.gs(0x21, (width - 1) | ((height - 1) << 4));
  }

  setDoubleWidth(enable = true) {
    return this.setTextSize(enable ? 2 : 1, 1);
  }

  setTextAlign(align = 'left') {
    const alignMap = {
      left: 0x00,
      center: 0x01,
      right: 0x02,
      lt: 0x00,
      ct: 0x01,
      rt: 0x02,
    };
    return this.esc(0x61, alignMap[align] || 0x00);
  }

  align(align) {
    return this.setTextAlign(align);
  }

  style(style) {
    if (style === 'bu') {
      this.setTextBold(true);
    } else if (style === 'normal') {
      this.setTextBold(false);
    }
    return this;
  }

  size(width, height) {
    return this.setTextSize(width, height);
  }

  text(content) {
    return this.raw(content + '\n');
  }

  println(content) {
    return this.text(content);
  }

  PrintLine(content) {
    return this.text(content);
  }

  PrintSeparator(char = '-', len = 32) {
    return this.raw(char.repeat(len) + '\n');
  }

  SetTextBold(bold) {
    return this.setTextBold(bold);
  }

  Feed(lines = 1) {
    this.buffer = Buffer.concat([this.buffer, Buffer.from([0x1b, 0x64, lines])]);
    return this;
  }

  feed(lines = 1) {
    return this.Feed(lines);
  }

  Cut(mode = 'full') {
    const modeVal = mode === 'partial' ? 0x01 : 0x00;
    this.buffer = Buffer.concat([this.buffer, Buffer.from([0x1d, 0x56, modeVal])]);
    return this;
  }

  cut(mode = 'full') {
    return this.Cut(mode);
  }

  close() {
    return this;
  }

  PrintHeader(storeName, orderNo) {
    return this
      .setTextAlign('center')
      .setTextBold(true)
      .setTextSize(2, 2)
      .text(storeName)
      .setTextSize(1, 1)
      .text(`单号: ${orderNo}`)
      .Feed(1)
      .setTextAlign('left')
      .setTextBold(false);
  }

  PrintFooter(footerText) {
    return this
      .Feed(1)
      .setTextAlign('center')
      .text(footerText)
      .Feed(2);
  }

  PrintSummary(label, value) {
    return this
      .setTextBold(true)
      .text(`${label}${value}`)
      .setTextBold(false);
  }

  PrintItemWithWidth(name, qty, price, subtotal, nameWidth = 18) {
    const truncatedName = name.length > nameWidth ? name.slice(0, nameWidth - 3) + '...' : name;
    const line = `${truncatedName.padEnd(nameWidth)} ${String(qty).padStart(3)} ${String(price).padStart(6)} ${String(subtotal).padStart(8)}`;
    return this.text(line);
  }

  Bytes() {
    return this.buffer;
  }

  bytes() {
    return this.buffer;
  }
}

function NewPrinter() {
  return new EscPosPrinter();
}

module.exports = {
  EscPosPrinter,
  NewPrinter,
  Printer: EscPosPrinter,
};
