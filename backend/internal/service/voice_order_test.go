package service

import (
	"testing"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/internal/model"
)

func buildMenuProducts() map[string]model.Product {
	menu := []struct {
		id   uint
		name string
		sku  string
		price string
	}{
		{1, "老坛酸菜鱼", "标准份", "58.00"},
		{2, "宫保鸡丁", "大份", "38.00"},
		{3, "麻婆豆腐", "中份", "18.00"},
		{4, "清炒时蔬", "小份", "16.00"},
		{5, "红烧肉", "例牌", "48.00"},
		{6, "白米饭", "一碗", "3.00"},
		{7, "可口可乐", "瓶装330ml", "6.00"},
		{8, "珍珠奶茶", "中杯", "15.00"},
		{9, "扬州炒饭", "大份", "22.00"},
		{10, "西红柿鸡蛋面", "标准份", "26.00"},
		{11, "剁椒鱼头", "整个", "88.00"},
		{12, "凉拌黄瓜", "小份", "12.00"},
	}

	m := make(map[string]model.Product)
	for _, it := range menu {
		product := model.Product{
			BaseModel: model.BaseModel{ID: it.id},
			Name:      it.name,
			StoreID:   1,
			Status:    1,
			MainImage: "/mock/" + it.name + ".jpg",
		}
		sku := model.ProductSKU{
			BaseModel: model.BaseModel{ID: it.id*10 + 1},
			ProductID: it.id,
			SKUCode:   "SKU" + it.name,
			SpecName:  it.sku,
			Price:     decimal.RequireFromString(it.price),
			Stock:     99,
			Status:    1,
		}
		product.SKUs = append(product.SKUs, sku)
		m[it.name] = product
	}
	return m
}

func TestExtractItems_MultiDish(t *testing.T) {
	svc := NewVoiceOrderService()

	cases := []struct {
		input    string
		expected int
		check    []ParsedItem
	}{
		{
			input:    "来份老坛酸菜鱼",
			expected: 1,
			check: []ParsedItem{
				{Name: "老坛酸菜鱼", Quantity: 1},
			},
		},
		{
			input:    "来份酸菜鱼",
			expected: 1,
			check: []ParsedItem{
				{Name: "酸菜鱼", Quantity: 1},
			},
		},
		{
			input:    "两碗白米饭",
			expected: 1,
			check: []ParsedItem{
				{Name: "白米饭", Quantity: 2},
			},
		},
		{
			input:    "加一瓶可口可乐",
			expected: 1,
			check: []ParsedItem{
				{Name: "可口可乐", Quantity: 1},
			},
		},
		{
			input:    "来个宫保鸡丁和麻婆豆腐",
			expected: 2,
			check: []ParsedItem{
				{Name: "宫保鸡丁", Quantity: 1},
				{Name: "麻婆豆腐", Quantity: 1},
			},
		},
		{
			input:    "三杯珍珠奶茶",
			expected: 1,
			check: []ParsedItem{
				{Name: "珍珠奶茶", Quantity: 3},
			},
		},
		{
			input:    "要一份红烧肉再来个清炒时蔬",
			expected: 2,
			check: []ParsedItem{
				{Name: "红烧肉", Quantity: 1},
				{Name: "清炒时蔬", Quantity: 1},
			},
		},
		{
			input:    "来两份扬州炒饭，三瓶可乐",
			expected: 2,
			check: []ParsedItem{
				{Name: "扬州炒饭", Quantity: 2},
				{Name: "可乐", Quantity: 3},
			},
		},
		{
			input:    "剁椒鱼头一个凉拌黄瓜两份，宫保鸡丁大份再来四碗米饭",
			expected: 4,
			check: []ParsedItem{
				{Name: "剁椒鱼头", Quantity: 1},
				{Name: "凉拌黄瓜", Quantity: 2},
				{Name: "宫保鸡丁大份", Quantity: 1},
				{Name: "米饭", Quantity: 4},
			},
		},
		{
			input:    "要两碗米饭三碗面",
			expected: 2,
			check: []ParsedItem{
				{Name: "米饭", Quantity: 2},
				{Name: "面", Quantity: 3},
			},
		},
	}

	for _, c := range cases {
		got := svc.extractItems(c.input)
		if len(got) != c.expected {
			t.Errorf("extractItems(%q): expected %d items, got %d=%v", c.input, c.expected, len(got), got)
			continue
		}
		for i, chk := range c.check {
			if i >= len(got) {
				t.Errorf("extractItems(%q)[%d]: missing item, expected %q", c.input, i, chk.Name)
				break
			}
			gotItem := got[i]
			if gotItem.Quantity != chk.Quantity {
				t.Errorf("extractItems(%q)[%d]: qty expected %d got %d (name=%q)",
					c.input, i, chk.Quantity, gotItem.Quantity, gotItem.Name)
			}
		}
	}
}

func TestFuzzyMatch_RealMenu(t *testing.T) {
	svc := NewVoiceOrderService()
	productMap := buildMenuProducts()
	productList := make([]model.Product, 0, len(productMap))
	for _, p := range productMap {
		productList = append(productList, p)
	}

	cases := []struct {
		input     string
		wantName  string
		wantMatch bool
	}{
		{"酸菜鱼", "老坛酸菜鱼", true},
		{"老坛酸菜鱼", "老坛酸菜鱼", true},
		{"宫爆鸡丁", "宫保鸡丁", true},
		{"西红柿面", "西红柿鸡蛋面", true},
		{"炒时蔬", "清炒时蔬", true},
		{"红烧猪肉", "红烧肉", true},
		{"可乐", "可口可乐", true},
		{"珍珠奶茶", "珍珠奶茶", true},
		{"扬州饭", "扬州炒饭", true},
		{"米饭", "白米饭", true},
		{"鱼头", "剁椒鱼头", true},
		{"拍黄瓜", "凉拌黄瓜", true},
		{"佛跳墙", "", false},
	}

	for _, c := range cases {
		got := svc.fuzzyMatch(c.input, productMap, productList)
		if c.wantMatch {
			if got == nil {
				t.Errorf("fuzzyMatch(%q): expected match %q, got nil", c.input, c.wantName)
				continue
			}
			if got.Name != c.wantName {
				t.Errorf("fuzzyMatch(%q): expected %q, got %q (score=%.3f)",
					c.input, c.wantName, got.Name, got.matchScore)
			}
		} else {
			if got != nil {
				t.Errorf("fuzzyMatch(%q): expected no match, got %q (score=%.3f)",
					c.input, got.Name, got.matchScore)
			}
		}
	}
}
