package service

import (
	"fmt"
	"regexp"
	"strconv"
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

var unitQuantifiers = []string{"份", "个", "碗", "盘", "碟", "杯", "桶", "只", "条", "块", "盒", "道", "瓶", "罐", "扎", "份儿", "杯装"}

var actionPrefixes = []string{"来份", "来个", "来碗", "来盘", "来杯", "来瓶",
	"要份", "要个", "要碗", "要盘", "要杯", "要瓶",
	"加份", "加个", "加碗", "加盘", "加杯", "加瓶",
	"点份", "点个", "点碗", "点盘", "点杯", "点瓶",
	"上份", "上个", "上碗", "上盘", "上杯", "上瓶",
	"来一", "要一", "加一", "点一", "上一",
	"来两", "要两", "加两", "点两", "上两"}

var splitWords = []string{"再来", "再要", "再点", "再加", "还要", "还来", "还加", "配上", "外加"}

var chiNums = map[rune]int{
	'零': 0, '一': 1, '二': 2, '两': 2, '三': 3, '四': 4,
	'五': 5, '六': 6, '七': 7, '八': 8, '九': 9, '十': 10,
}

func parseChiNum(s string) (int, bool) {
	runes := []rune(s)
	if len(runes) == 0 {
		return 0, false
	}
	total := 0
	current := 0
	for _, r := range runes {
		n, ok := chiNums[r]
		if !ok {
			return 0, false
		}
		if r == '十' {
			if current == 0 {
				current = 1
			}
			total += current * 10
			current = 0
		} else {
			current = n
		}
	}
	total += current
	if total == 0 && len(runes) == 1 {
		if n, ok := chiNums[runes[0]]; ok {
			return n, true
		}
	}
	return total, total > 0
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

	productSlices := products
	productMap := make(map[string]model.Product)
	productList := make([]model.Product, 0, len(productSlices))
	for _, p := range productSlices {
		productMap[p.Name] = p
		productList = append(productList, p)
	}

	var results []dto.VoiceMatchResult
	var unmatched []string

	for _, item := range items {
		matched := s.fuzzyMatch(item.Name, productMap, productList)
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

func (s *VoiceOrderService) fuzzyMatch(name string, productMap map[string]model.Product, productList []model.Product) *fuzzyMatchResult {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}

	if p, ok := productMap[name]; ok {
		return &fuzzyMatchResult{Product: p, matchScore: 1.0}
	}

	lowerName := strings.ToLower(name)
	runeName := []rune(name)

	var bestMatch *fuzzyMatchResult

	for _, product := range productList {
		productName := product.Name
		lowerProductName := strings.ToLower(productName)

		if lowerProductName == lowerName {
			return &fuzzyMatchResult{Product: product, matchScore: 1.0}
		}

		score := 0.0

		if strings.Contains(lowerProductName, lowerName) {
			score = 0.8 + 0.2*float64(len(runeName))/float64(len([]rune(productName)))
		} else if strings.Contains(lowerName, lowerProductName) {
			score = 0.7 + 0.2*float64(len([]rune(productName)))/float64(len(runeName))
		} else {
			score = s.calcMatchScore(lowerName, lowerProductName)
		}

		if score > 0 && (bestMatch == nil || score > bestMatch.matchScore) {
			bestMatch = &fuzzyMatchResult{Product: product, matchScore: score}
		}
	}

	if bestMatch != nil && bestMatch.matchScore >= 0.4 {
		return bestMatch
	}

	return nil
}

func (s *VoiceOrderService) calcMatchScore(input, target string) float64 {
	inputRunes := []rune(input)
	targetRunes := []rune(target)

	if len(inputRunes) == 0 || len(targetRunes) == 0 {
		return 0
	}

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

	f1 := 2 * precision * recall / (precision + recall)

	consecutiveScore := 0.0
	maxConsecutive := 0

	for tiStart := 0; tiStart < len(targetRunes); tiStart++ {
		for iiStart := 0; iiStart < len(inputRunes); iiStart++ {
			consecutive := 0
			i, j := iiStart, tiStart
			for i < len(inputRunes) && j < len(targetRunes) && inputRunes[i] == targetRunes[j] {
				consecutive++
				i++
				j++
			}
			if consecutive > maxConsecutive {
				maxConsecutive = consecutive
			}
		}
	}

	if maxConsecutive > 1 {
		consecutiveScore = float64(maxConsecutive) / float64(targetLen) * 0.35
	}

	return f1*0.65 + consecutiveScore
}

func (s *VoiceOrderService) extractItems(text string) []ParsedItem {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, ",", "，")
	text = strings.ReplaceAll(text, ";", "；")
	text = strings.ReplaceAll(text, ".", "。")

	for _, sw := range splitWords {
		text = strings.ReplaceAll(text, sw, "，")
	}

	var rawParts []string
	parenRe := regexp.MustCompile(`[，,、；;。！!？?\n]+`)
	rawParts = parenRe.Split(text, -1)

	var parts []string
	for _, p := range rawParts {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}

	var items []ParsedItem
	seen := make(map[string]bool)

	for _, part := range parts {
		parsed := s.parseSingleItem(part)
		for _, pi := range parsed {
			name := strings.TrimSpace(pi.Name)
			if name == "" {
				continue
			}
			key := fmt.Sprintf("%s_%d", name, pi.Quantity)
			if !seen[key] {
				items = append(items, pi)
				seen[key] = true
			}
		}
	}

	if len(items) == 0 && text != "" {
		parsed := s.parseSingleItem(text)
		if len(parsed) > 0 {
			items = parsed
		} else {
			items = []ParsedItem{{Name: text, Quantity: 1}}
		}
	}

	return items
}

func (s *VoiceOrderService) parseSingleItem(part string) []ParsedItem {
	part = strings.TrimSpace(part)
	if part == "" {
		return nil
	}

	qtyPrefixMap := map[string]int{
		"一": 1, "二": 2, "两": 2, "三": 3, "四": 4,
		"五": 5, "六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
		"十一": 11, "十二": 12, "十三": 13, "十四": 14, "十五": 15,
		"二十": 20, "半": 1, "小份": 1, "大份": 1,
	}

	for apIdx := len(actionPrefixes) - 1; apIdx >= 0; apIdx-- {
		prefix := actionPrefixes[apIdx]
		if strings.HasPrefix(part, prefix) {
			remaining := strings.TrimPrefix(part, prefix)
			remaining = strings.TrimSpace(remaining)
			if remaining == "" {
				continue
			}

			qty := 1
			for chiNum, num := range qtyPrefixMap {
				if strings.HasPrefix(remaining, chiNum) {
					qty = num
					remaining = strings.TrimPrefix(remaining, chiNum)
					break
				}
			}

			for _, uq := range unitQuantifiers {
				remaining = strings.TrimPrefix(remaining, uq)
			}

			remaining = strings.TrimSpace(remaining)
			if remaining != "" {
				return s.parseSingleItemWithQty(remaining, qty)
			}
		}
	}

	for cn, num := range qtyPrefixMap {
		if strings.HasPrefix(part, cn) {
			after := strings.TrimPrefix(part, cn)
			for _, uq := range unitQuantifiers {
				if strings.HasPrefix(after, uq) {
					after = strings.TrimPrefix(after, uq)
					after = strings.TrimSpace(after)
					if after != "" {
						return s.parseSingleItemWithQty(after, num)
					}
				}
			}
			after = strings.TrimSpace(after)
			if len(after) > 0 {
				if isDigitOrChinese(after) {
					return s.parseSingleItemWithQty(after, num)
				}
			}
		}
	}

	arabicQtyRe := regexp.MustCompile(`^(\d+)`)
	matches := arabicQtyRe.FindStringSubmatch(part)
	if len(matches) >= 2 {
		qty64, _ := strconv.ParseInt(matches[1], 10, 32)
		qty := int(qty64)
		if qty <= 0 {
			qty = 1
		}
		remaining := strings.TrimPrefix(part, matches[1])
		for _, uq := range unitQuantifiers {
			remaining = strings.TrimPrefix(remaining, uq)
		}
		remaining = strings.TrimSpace(remaining)
		if remaining != "" {
			return s.parseSingleItemWithQty(remaining, qty)
		}
	}

	trailingQtyRe := regexp.MustCompile(`(.+?)\s*(\d+)\s*([份个碗盘碟杯桶只条块盒道瓶罐扎]|份儿|杯装)?\s*$`)
	if trailingMatches := trailingQtyRe.FindStringSubmatch(part); len(trailingMatches) >= 3 {
		name := strings.TrimSpace(trailingMatches[1])
		qty64, _ := strconv.ParseInt(trailingMatches[2], 10, 32)
		qty := int(qty64)
		if qty <= 0 {
			qty = 1
		}
		name = trimTrailingUnits(name)
		if name != "" {
			return []ParsedItem{{Name: name, Quantity: qty}}
		}
	}

	chineseTrailRe := regexp.MustCompile(`^(.+?)\s*([零一二两三四五六七八九十]+)\s*([份个碗盘碟杯桶只条块盒道瓶罐扎]|份儿|杯装)?\s*$`)
	if chineseTrailMatches := chineseTrailRe.FindStringSubmatch(part); len(chineseTrailMatches) >= 3 {
		name := strings.TrimSpace(chineseTrailMatches[1])
		qtyStr := chineseTrailMatches[2]
		if qty, ok := parseChiNum(qtyStr); ok && qty > 0 {
			name = trimTrailingUnits(name)
			if name != "" {
				return []ParsedItem{{Name: name, Quantity: qty}}
			}
		}
	}

	part = trimTrailingUnits(part)
	part = strings.TrimSpace(part)
	if part != "" {
		return []ParsedItem{{Name: part, Quantity: 1}}
	}

	return nil
}

func (s *VoiceOrderService) parseSingleItemWithQty(name string, qty int) []ParsedItem {
	name = trimTrailingUnits(name)
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	if qty <= 0 {
		qty = 1
	}
	return []ParsedItem{{Name: name, Quantity: qty}}
}

func trimTrailingUnits(s string) string {
	s = strings.TrimSpace(s)
	changed := true
	for changed {
		changed = false
		for _, uq := range unitQuantifiers {
			if strings.HasSuffix(s, uq) {
				s = strings.TrimSuffix(s, uq)
				s = strings.TrimSpace(s)
				changed = true
				break
			}
		}
	}
	return s
}

func isChineseRune(r uint8) bool {
	return r >= 0x80
}

func isDigitOrChinese(s string) bool {
	if len(s) == 0 {
		return false
	}
	r := []rune(s)[0]
	return unicode.IsDigit(r) || unicode.Is(unicode.Han, r)
}

func (s *VoiceOrderService) getDefaultSKU(product model.Product) model.ProductSKU {
	if len(product.SKUs) == 0 {
		return model.ProductSKU{}
	}

	for _, sku := range product.SKUs {
		if sku.Status == 1 && sku.Stock > 0 {
			return sku
		}
	}

	for _, sku := range product.SKUs {
		if sku.Status == 1 {
			return sku
		}
	}

	return product.SKUs[0]
}

func (s *VoiceOrderService) getDefaultAttributeValues(product model.Product, sku model.ProductSKU) map[uint]uint {
	result := make(map[uint]uint)

	for _, attr := range product.Attributes {
		if attr.Status != 1 {
			continue
		}
		matched := false
		for _, skuAttrVal := range sku.AttributeValues {
			if skuAttrVal.AttributeID == attr.ID {
				result[attr.ID] = skuAttrVal.ValueID
				matched = true
				break
			}
		}
		if !matched && len(attr.Values) > 0 {
			for _, val := range attr.Values {
				if val.Status == 1 && (val.Stock == -1 || val.Stock > 0) {
					result[attr.ID] = val.ID
					break
				}
			}
			if result[attr.ID] == 0 && len(attr.Values) > 0 {
				result[attr.ID] = attr.Values[0].ID
			}
		}
	}

	return result
}

func intPtr(v int) *int {
	return &v
}
