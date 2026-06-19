'use strict';

const fs = require('fs');
const path = require('path');
const logger = require('../utils/logger');

const DEFAULT_TEMPLATES = {
  kitchen: {
    name: '后厨小票',
    type: 'kitchen',
    width: 80,
    encoding: 'UTF-8',
    copies: 1,
    autoCut: true,
    header: {
      align: 'center',
      bold: true,
      doubleWidth: true,
      lines: [
        '{storeName}',
        '后厨单',
      ],
    },
    infoSection: {
      lines: [
        '单号: {orderNo}',
        '{tableNo,桌号: {value}}',
        '类型: {orderType}',
        '时间: {createdAt}',
      ],
    },
    itemsSection: {
      header: {
        bold: true,
        line: '菜品              数量',
        separator: '----------------',
      },
      itemFormat: '{productName,16} {quantity,4,r}',
    },
    footer: {
      align: 'center',
      lines: [
        '请及时备菜',
      ],
      feedLines: 3,
    },
  },
  receipt: {
    name: '前台小票',
    type: 'receipt',
    width: 80,
    encoding: 'UTF-8',
    copies: 1,
    autoCut: true,
    header: {
      align: 'center',
      bold: true,
      doubleWidth: true,
      lines: [
        '{storeName}',
        '结账单',
      ],
    },
    infoSection: {
      lines: [
        '单号: {orderNo}',
        '{tableNo,桌号: {value}}',
        '类型: {orderType}',
        '下单时间: {createdAt}',
      ],
    },
    itemsSection: {
      header: {
        bold: true,
        line: '菜品        数量  单价  金额',
        separator: '---------------------------',
      },
      itemFormat: '{productName,10} {quantity,2,r} {price,5,r} {subtotal,6,r}',
    },
    summarySection: {
      separator: '---------------------------',
      lines: [
        '合计: {totalAmount}',
        '优惠: {discountAmount}',
        '应收: {payAmount}',
        '{payStatus,支付方式: {payMethod}}',
      ],
    },
    footer: {
      align: 'center',
      lines: [
        '欢迎下次光临',
      ],
      feedLines: 3,
    },
  },
  drink: {
    name: '饮品小票',
    type: 'drink',
    width: 80,
    encoding: 'UTF-8',
    copies: 1,
    autoCut: true,
    header: {
      align: 'center',
      bold: true,
      doubleWidth: true,
      lines: [
        '{storeName}',
        '饮品单',
      ],
    },
    infoSection: {
      lines: [
        '单号: {orderNo}',
        '{tableNo,桌号: {value}}',
        '时间: {createdAt}',
      ],
    },
    itemsSection: {
      header: {
        bold: true,
        line: '饮品              数量',
        separator: '----------------',
      },
      itemFormat: '{productName,16} {quantity,4,r}',
    },
    footer: {
      align: 'center',
      lines: [
        '请及时制作',
      ],
      feedLines: 3,
    },
  },
  cold: {
    name: '凉菜小票',
    type: 'cold',
    width: 80,
    encoding: 'UTF-8',
    copies: 1,
    autoCut: true,
    header: {
      align: 'center',
      bold: true,
      doubleWidth: true,
      lines: [
        '{storeName}',
        '凉菜单',
      ],
    },
    infoSection: {
      lines: [
        '单号: {orderNo}',
        '{tableNo,桌号: {value}}',
        '时间: {createdAt}',
      ],
    },
    itemsSection: {
      header: {
        bold: true,
        line: '凉菜              数量',
        separator: '----------------',
      },
      itemFormat: '{productName,16} {quantity,4,r}',
    },
    footer: {
      align: 'center',
      lines: [
        '请及时备菜',
      ],
      feedLines: 3,
    },
  },
};

class PrintTemplateManager {
  constructor() {
    this.templates = { ...DEFAULT_TEMPLATES };
    this.customTemplatesDir = path.join(process.cwd(), 'data', 'templates');
    this.ensureTemplatesDir();
    this.loadCustomTemplates();
  }

  ensureTemplatesDir() {
    if (!fs.existsSync(this.customTemplatesDir)) {
      fs.mkdirSync(this.customTemplatesDir, { recursive: true });
    }
  }

  loadCustomTemplates() {
    try {
      if (fs.existsSync(this.customTemplatesDir)) {
        const files = fs.readdirSync(this.customTemplatesDir).filter((f) => f.endsWith('.json'));
        for (const file of files) {
          try {
            const content = fs.readFileSync(path.join(this.customTemplatesDir, file), 'utf-8');
            const template = JSON.parse(content);
            if (template && template.type) {
              this.templates[template.type] = template;
              logger.info('[PrintTemplate] 已加载自定义模板: %s', template.name || template.type);
            }
          } catch (err) {
            logger.error('[PrintTemplate] 加载模板文件失败 %s: %s', file, err.message);
          }
        }
      }
    } catch (err) {
      logger.error('[PrintTemplate] 加载自定义模板失败: %s', err.message);
    }
  }

  getTemplate(type) {
    return this.templates[type] || this.templates.kitchen;
  }

  getAllTemplates() {
    return Object.values(this.templates).map((t) => ({
      type: t.type,
      name: t.name,
      width: t.width,
      copies: t.copies,
    }));
  }

  saveCustomTemplate(type, template) {
    this.templates[type] = { ...template, type };
    const filePath = path.join(this.customTemplatesDir, `${type}.json`);
    fs.writeFileSync(filePath, JSON.stringify(template, null, 2), 'utf-8');
    logger.info('[PrintTemplate] 已保存自定义模板: %s', type);
    return this.templates[type];
  }

  deleteCustomTemplate(type) {
    if (DEFAULT_TEMPLATES[type]) {
      this.templates[type] = DEFAULT_TEMPLATES[type];
    } else {
      delete this.templates[type];
    }
    const filePath = path.join(this.customTemplatesDir, `${type}.json`);
    if (fs.existsSync(filePath)) {
      fs.unlinkSync(filePath);
    }
    logger.info('[PrintTemplate] 已删除自定义模板: %s', type);
  }

  formatTemplate(template, data) {
    const result = [];

    if (template.header) {
      for (const line of template.header.lines) {
        result.push({
          type: 'text',
          content: this.renderTemplateString(line, data),
          align: template.header.align || 'left',
          bold: template.header.bold || false,
          doubleWidth: template.header.doubleWidth || false,
        });
      }
    }

    if (template.infoSection) {
      result.push({ type: 'separator', content: '----------------' });
      for (const line of template.infoSection.lines) {
        const rendered = this.renderConditionalTemplateString(line, data);
        if (rendered !== null) {
          result.push({ type: 'text', content: rendered });
        }
      }
    }

    if (template.itemsSection && data.items && data.items.length > 0) {
      if (template.itemsSection.header) {
        result.push({ type: 'separator', content: template.itemsSection.header.separator || '----------------' });
        result.push({
          type: 'text',
          content: template.itemsSection.header.line,
          bold: template.itemsSection.header.bold,
        });
        result.push({ type: 'separator', content: template.itemsSection.header.separator || '----------------' });
      }

      for (const item of data.items) {
        result.push({
          type: 'text',
          content: this.renderItemLine(template.itemsSection.itemFormat, item),
        });
      }
    }

    if (template.summarySection) {
      result.push({ type: 'separator', content: template.summarySection.separator || '----------------' });
      for (const line of template.summarySection.lines) {
        const rendered = this.renderConditionalTemplateString(line, data);
        if (rendered !== null) {
          result.push({ type: 'text', content: rendered, bold: true });
        }
      }
    }

    if (template.footer) {
      result.push({ type: 'feed', lines: 1 });
      for (const line of template.footer.lines) {
        result.push({
          type: 'text',
          content: this.renderTemplateString(line, data),
          align: template.footer.align || 'left',
        });
      }
      result.push({ type: 'feed', lines: template.footer.feedLines || 3 });
    }

    return result;
  }

  renderTemplateString(templateStr, data) {
    return templateStr.replace(/\{(\w+)\}/g, (match, key) => {
      if (data[key] !== undefined && data[key] !== null) {
        return String(data[key]);
      }
      return '';
    });
  }

  renderConditionalTemplateString(templateStr, data) {
    const match = templateStr.match(/^\{(\w+),(.+)\}$/);
    if (match) {
      const [, conditionKey, trueTemplate] = match;
      if (data[conditionKey] !== undefined && data[conditionKey] !== null && data[conditionKey] !== '') {
        return this.renderTemplateString(trueTemplate, { ...data, value: data[conditionKey] });
      }
      return null;
    }
    return this.renderTemplateString(templateStr, data);
  }

  renderItemLine(format, item) {
    return format.replace(/\{(\w+)(?:,(\d+)(?:,(l|r))?)?\}/g, (match, key, width, align) => {
      let value = item[key] !== undefined ? String(item[key]) : '';
      if (width) {
        const w = parseInt(width);
        if (align === 'r') {
          value = value.padStart(w);
        } else {
          value = value.padEnd(w);
        }
        if (value.length > w) {
          value = value.slice(0, w - 3) + '...';
        }
      }
      return value;
    });
  }
}

const templateManager = new PrintTemplateManager();

module.exports = {
  templateManager,
  DEFAULT_TEMPLATES,
};
