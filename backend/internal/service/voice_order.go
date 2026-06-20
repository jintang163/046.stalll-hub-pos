package service

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
)

type VoiceOrderService struct {
	productRepo *repository.ProductRepository
}

func NewVoiceOrderService() *VoiceOrderService {
	return &VoiceOrderService{
		productRepo: repository.NewProductRepository(),
	}
}

type ParsedItem struct {
	Name     string
	Quantity int
}

func (s *VoiceOrderService) ParseVoiceText(storeID uint, text string) (*dto.VoiceParseResponse, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("语音文本不能为空")
	}

	items := s.extractItems(text)

	products, err := s.productRepo.List(storeID, 0, "", intPtr(1), nil, nil, 0, 1000)
	if err != nil {
		return nil, fmt.Errorf("获取商品列表失败: %w", err)
	}

	productMap := make(map[string]model.Product)
	for _, p := range products {
		productMap[p.Name] = p
	}

	var results []dto.VoiceMatchResult
	var unmatched []string

	for _, item := range items {
		matched := s.fuzzyMatch(item.Name, productMap)
		if matched != nil {
			sku := s.getDefaultSKU(*matched)
			results = append(results, dto.VoiceMatchResult{
				ProductID:   matched.ID,
				ProductName: matched.Name,
				SKUID:       sku.ID,
				SKUName:     sku.SpecName,
				Price:       sku.Price,
				Quantity:    item.Quantity,
				MatchScore:  matched.matchScore,
				Image:       matched.MainImage,
			})
		} else {
			unmatched = append(unmatched, item.Name)
		}
	}

	return &dto.VoiceParseResponse{
		OriginalText: text,
		Items:        results,
		Unmatched:    unmatched,
	}, nil
}

type fuzzyMatchResult struct {
	model.Product
	matchScore float64
}

func (s *VoiceOrderService) fuzzyMatch(name string, productMap map[string]model.Product) *fuzzyMatchResult {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}

	if p, ok := productMap[name]; ok {
		return &fuzzyMatchResult{Product: p, matchScore: 1.0}
	}

	lowerName := strings.ToLower(name)
	var bestMatch *fuzzyMatchResult

	for productName, product := range productMap {
		lowerProductName := strings.ToLower(productName)

		score := s.calcMatchScore(lowerName, lowerProductName)
		if score > 0 && (bestMatch == nil || score > bestMatch.matchScore) {
			bestMatch = &fuzzyMatchResult{Product: product, matchScore: score}
		}
	}

	if bestMatch != nil && bestMatch.matchScore >= 0.3 {
		return bestMatch
	}

	return nil
}

func (s *VoiceOrderService) calcMatchScore(input, target string) float64 {
	if input == target {
		return 1.0
	}

	if strings.Contains(target, input) {
		return 0.8 + 0.2*float64(len(input))/float64(len(target))
	}

	if strings.Contains(input, target) {
		return 0.7 + 0.2*float64(len(target))/float64(len(input))
	}

	inputRunes := []rune(input)
	targetRunes := []rune(target)

	commonCount := 0
	targetUsed := make([]bool, len(targetRunes))

	for _, ir := range inputRunes {
		for j, tr := range targetRunes {
			if !targetUsed[j] && ir == tr {
				commonCount++
				targetUsed[j] = true
				break
			}
		}
	}

	if commonCount == 0 {
		return 0
	}

	inputLen := len(inputRunes)
	targetLen := len(targetRunes)

	precision := float64(commonCount) / float64(inputLen)
	recall := float64(commonCount) / float64(targetLen)

	if precision+recall == 0 {
		return 0
	}

	f1 := 2 * precision * recall / (precision + recall)

	consecutiveScore := 0.0
	maxConsecutive := 0
	for i := 0; i < len(inputRunes); i++ {
		consecutive := 0
		ti := 0
		for j := i; j < len(inputRunes); j++ {
			found := false
			for k := ti; k < len(targetRunes); k++ {
				if inputRunes[j] == targetRunes[k] {
					consecutive++
					ti = k + 1
					found = true
					break
				}
			}
			if !found {
				break
			}
		}
		if consecutive > maxConsecutive {
			maxConsecutive = consecutive
		}
	}
	if maxConsecutive > 1 {
		consecutiveScore = float64(maxConsecutive) / float64(targetLen) * 0.3
	}

	return f1*0.7 + consecutiveScore
}

func (s *VoiceOrderService) extractItems(text string) []ParsedItem {
	text = normalizeText(text)

	patterns := []struct {
		regex   string
		nameIdx int
		qtyIdx  int
	}{
		{`(\d+)\s*[份个碗盘碟杯桶份只条块盒]?(.+?)`, 2, 1},
		{`(.+?)\s*(\d+)\s*[份个碗盘碟杯桶份只条块盒]?`, 1, 2},
		{`来\s*[份个碗盘碟杯桶只条块]?(\d+)?\s*(.+?)`, 2, 1},
		{`要\s*[份个碗盘碟杯桶只条块]?(\d+)?\s*(.+?)`, 2, 1},
		{`加\s*[份个碗盘碟杯桶只条块]?(\d+)?\s*(.+?)`, 2, 1},
		{`点\s*[份个碗盘碟杯桶只条块]?(\d+)?\s*(.+?)`, 2, 1},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.regex)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 0 {
			name := strings.TrimSpace(matches[p.nameIdx])
			qty := 1
			if p.qtyIdx < len(matches) && matches[p.qtyIdx] != "" {
				fmt.Sscanf(matches[p.qtyIdx], "%d", &qty)
			}
			if qty <= 0 {
				qty = 1
			}
			return []ParsedItem{{Name: name, Quantity: qty}}
		}
	}

	quantityPrefixes := []string{"来份", "来个", "要份", "要个", "加份", "加个", "点份", "点个",
		"来一", "要一", "加一", "点一", "来两", "要两", "加两", "点两"}
	for _, prefix := range quantityPrefixes {
		if strings.HasPrefix(text, prefix) {
			name := strings.TrimPrefix(text, prefix)
			qty := 1
			if strings.Contains(prefix, "两") {
				qty = 2
			}
			if name != "" {
				return []ParsedItem{{Name: name, Quantity: qty}}
			}
		}
	}

	separatorPattern := regexp.MustCompile(`[，,、；;和还再]`)
	parts := separatorPattern.Split(text, -1)

	var items []ParsedItem
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		qty := 1
		name := part

		qtyRe := regexp.MustCompile(`^(\d+)`)
		qtyMatches := qtyRe.FindStringSubmatch(part)
		if len(qtyMatches) > 1 {
			fmt.Sscanf(qtyMatches[1], "%d", &qty)
			name = strings.TrimLeft(part, "0123456789")
			name = strings.TrimLeft(name, "份个碗盘碟杯桶只条块")
		}

		name = strings.TrimPrefix(name, "来份")
		name = strings.TrimPrefix(name, "来个")
		name = strings.TrimPrefix(name, "要份")
		name = strings.TrimPrefix(name, "要个")
		name = strings.TrimPrefix(name, "加份")
		name = strings.TrimPrefix(name, "加个")

		if name != "" {
			items = append(items, ParsedItem{Name: name, Quantity: qty})
		}
	}

	if len(items) > 0 {
		return items
	}

	return []ParsedItem{{Name: text, Quantity: 1}}
}

func (s *VoiceOrderService) getDefaultSKU(product model.Product) model.ProductSKU {
	if len(product.SKUs) > 0 {
		for _, sku := range product.SKUs {
			if sku.Status == 1 && sku.Stock > 0 {
				return sku
			}
		}
		if len(product.SKUs) > 0 {
			return product.SKUs[0]
		}
	}
	return model.ProductSKU{}
}

func normalizeText(text string) string {
	var b strings.Builder
	prev := rune(0)
	for _, r := range text {
		if unicode.Is(unicode.Han, r) || unicode.IsLetter(r) || unicode.IsDigit(r) {
			if r != prev {
				b.WriteRune(r)
			}
			prev = r
		} else {
			b.WriteRune(r)
			prev = 0
		}
	}
	return b.String()
}

func intPtr(v int) *int {
	return &v
}
