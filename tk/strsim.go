package tk

import (
	"hash/fnv"
	"math"
	"regexp"
	"strings"
)

// Config 定义了相似度计算的配置选项
type SimilarityConfig struct {
	IgnoreSpaces   bool // 忽略空格
	IgnoreNewlines bool // 忽略换行符
	IgnorePunct    bool // 忽略标点符号
	IgnoreCase     bool // 忽略大小写
	NgramSize      int  // N-gram 的大小，默认为 2
	SimhashBits    int  // Simhash 的位数，默认为 64
}

// defaultConfig 返回默认配置
func defaultSimilarityConfig() *SimilarityConfig {
	return &SimilarityConfig{
		IgnoreSpaces:   true,
		IgnoreNewlines: true,
		IgnorePunct:    true,
		IgnoreCase:     true,
		NgramSize:      2,  // 默认使用 2-gram
		SimhashBits:    64, // 默认使用 64 位 SimHash
	}
}

// StringComparer 结构体用于计算字符串相似度
type StringSimilarity struct {
	config *SimilarityConfig
}

// NewStringComparer 创建一个新的 StringComparer 实例
func NewStringSimilarity() *StringSimilarity {
	return &StringSimilarity{
		config: defaultSimilarityConfig(),
	}
}

func (ss *StringSimilarity) Set(key string, value interface{}) *StringSimilarity {
	switch key {
	case "space":
		ss.config.IgnoreSpaces, _ = value.(bool)
	case "newline":
		ss.config.IgnoreNewlines, _ = value.(bool)
	case "punct":
		ss.config.IgnorePunct, _ = value.(bool)
	case "case":
		ss.config.IgnoreCase, _ = value.(bool)
	case "ngram":
		ss.config.NgramSize, _ = value.(int)
	case "simhash":
		ss.config.SimhashBits, _ = value.(int)
	}

	// 确保 NgramSize 和 SimhashBits 有效
	if ss.config.NgramSize <= 0 {
		ss.config.NgramSize = defaultSimilarityConfig().NgramSize
	}
	if ss.config.SimhashBits <= 0 || ss.config.SimhashBits > 64 {
		ss.config.SimhashBits = defaultSimilarityConfig().SimhashBits
	}

	return ss
}

// dotProduct 计算两个向量的点积
func dotProduct(vec1, vec2 map[string]int) float64 {
	sum := 0.0
	for k, v1 := range vec1 {
		if v2, ok := vec2[k]; ok {
			sum += float64(v1 * v2)
		}
	}
	return sum
}

// magnitude 计算向量的模
func magnitude(vec map[string]int) float64 {
	sumOfSquares := 0.0
	for _, v := range vec {
		sumOfSquares += float64(v * v)
	}
	return math.Sqrt(sumOfSquares)
}

func min(a int, b ...int) int {
	min := a
	for _, i := range b {
		if i < min {
			min = i
		}
	}
	return min
}

// preprocessString 根据配置对字符串进行预处理
func (ss *StringSimilarity) preprocessString(s string) string {
	if ss.config.IgnoreNewlines {
		s = strings.ReplaceAll(s, "\n", "")
		s = strings.ReplaceAll(s, "\r", "")
	}
	if ss.config.IgnoreSpaces {
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "\t", "")
	}
	if ss.config.IgnorePunct {
		// 使用正则表达式去除标点符号
		// 这里的 Unicode 标点符号范围可能需要根据实际情况调整
		// \p{P} 表示任何标点符号字符
		reg := regexp.MustCompile(`[^\p{L}\p{N}\s]`) // 保留字母、数字和空格 (如果 IgnoreSpaces 为 false)
		if ss.config.IgnoreSpaces {
			reg = regexp.MustCompile(`[^\p{L}\p{N}]`) // 保留字母、数字
		}
		s = reg.ReplaceAllString(s, "")
	}
	if ss.config.IgnoreCase {
		s = strings.ToLower(s)
	}
	return s
}

// generateNGrams 将字符串生成 n-grams 集合
// 对于中文，我们按字符生成 n-grams
// 对于英文，可以按字符或按词生成。这里统一按字符。
func generateNGrams(s string, ngramSize int) map[string]struct{} {
	grams := make(map[string]struct{})
	runes := []rune(s) // 使用 rune 处理 Unicode 字符（中文字符）

	if len(runes) == 0 {
		return grams
	}

	for i := 0; i <= len(runes)-ngramSize; i++ {
		gram := string(runes[i : i+ngramSize])
		grams[gram] = struct{}{}
	}

	// 对于短于 NgramSize 的字符串，特殊处理，将其自身作为一个 gram
	if len(runes) < ngramSize && len(runes) > 0 {
		grams[s] = struct{}{}
	}

	return grams
}

// generateNgramCounts 将字符串生成 n-grams 计数向量
func generateNgramCounts(s string, ngramSize int) map[string]int {
	counts := make(map[string]int)
	runes := []rune(s)

	if len(runes) == 0 {
		return counts
	}

	for i := 0; i <= len(runes)-ngramSize; i++ {
		gram := string(runes[i : i+ngramSize])
		counts[gram]++
	}

	// 对于短于 NgramSize 的字符串，特殊处理
	if len(runes) < ngramSize && len(runes) > 0 {
		counts[s]++
	}
	return counts
}

// JaccardSimilarity 计算两个字符串的 Jaccard 相似度
// 返回值在 0.0 到 1.0 之间
func (ss *StringSimilarity) JaccardSimilarity(s1, s2 string) float64 {
	ps1 := ss.preprocessString(s1)
	ps2 := ss.preprocessString(s2)

	if ps1 == "" && ps2 == "" {
		return 1.0 // 两个空字符串认为完全相似
	}
	if ps1 == "" || ps2 == "" {
		return 0.0 // 一个为空，一个不为空，认为不相似
	}

	grams1 := generateNGrams(ps1, ss.config.NgramSize)
	grams2 := generateNGrams(ps2, ss.config.NgramSize)

	// 计算交集
	intersectionCount := 0
	for gram := range grams1 {
		if _, exists := grams2[gram]; exists {
			intersectionCount++
		}
	}

	// 计算并集
	unionCount := len(grams1) + len(grams2) - intersectionCount

	if unionCount == 0 {
		return 0.0 // 避免除以零，通常发生在两个字符串都没有生成任何 gram 的情况
	}

	return float64(intersectionCount) / float64(unionCount)
}

// CosineSimilarity (可选实现)
// Cosine 相似度通常需要将字符串向量化 (例如使用 TF-IDF 或 CountVectorizer)，
// 然后计算向量的余弦。这里我们简化为基于 n-gram 频率的向量。

// CosineSimilarity 计算两个字符串的余弦相似度
// 返回值在 0.0 到 1.0 之间
func (ss *StringSimilarity) CosineSimilarity(s1, s2 string) float64 {
	ps1 := ss.preprocessString(s1)
	ps2 := ss.preprocessString(s2)

	if ps1 == "" && ps2 == "" {
		return 1.0
	}
	if ps1 == "" || ps2 == "" {
		return 0.0
	}

	counts1 := generateNgramCounts(ps1, ss.config.NgramSize)
	counts2 := generateNgramCounts(ps2, ss.config.NgramSize)

	dp := dotProduct(counts1, counts2)
	mag1 := magnitude(counts1)
	mag2 := magnitude(counts2)

	if mag1 == 0 || mag2 == 0 {
		return 0.0 // 避免除以零
	}

	return dp / (mag1 * mag2)
}

// levenshteinDistance 计算两个字符串的 Levenshtein 距离
// 返回将 s1 转换为 s2 所需的最小编辑操作数 (插入、删除、替换)
// 注意：此函数假定输入字符串已经过预处理
func levenshteinDistance(ps1, ps2 []rune) int {
	len1 := len(ps1)
	len2 := len(ps2)

	// 处理空字符串情况
	if len1 == 0 {
		return len2
	}
	if len2 == 0 {
		return len1
	}

	// dp 数组，只保留两行以节省内存
	// dp[j] 表示 s1[:i] 与 s2[:j] 的距离
	// prev_dp[j] 表示 s1[:i-1] 与 s2[:j] 的距离
	dp := make([]int, len2+1)
	prev_dp := make([]int, len2+1)

	// 初始化第一行
	for j := 0; j <= len2; j++ {
		prev_dp[j] = j
	}

	for i := 1; i <= len1; i++ {
		dp[0] = i // 初始化第一列
		for j := 1; j <= len2; j++ {
			cost := 0
			if ps1[i-1] != ps2[j-1] {
				cost = 1
			}
			dp[j] = min(prev_dp[j]+1, // 删除
				dp[j-1]+1,         // 插入
				prev_dp[j-1]+cost) // 替换 (或匹配)
		}
		// 将当前行的数据复制到 prev_dp，为下一行计算做准备
		copy(prev_dp, dp)
	}

	return dp[len2]
}

// LevenshteinSimilarity 计算两个字符串的 Levenshtein 相似度
// 返回值在 0.0 到 1.0 之间 (1.0 表示完全相同)
func (ss *StringSimilarity) LevenshteinSimilarity(s1, s2 string) float64 {
	ps1 := []rune(ss.preprocessString(s1))
	ps2 := []rune(ss.preprocessString(s2))

	len1 := len(ps1)
	len2 := len(ps2)

	maxLength := float64(math.Max(float64(len1), float64(len2)))

	if maxLength == 0 {
		return 1.0 // 两个空字符串被认为是完全相似
	}

	distance := float64(levenshteinDistance(ps1, ps2)) // 注意这里调用的是原始的 s1, s2，内部会再次预处理
	return (maxLength - distance) / maxLength
}

// jaroDistance 计算两个字符串的 Jaro 相似度 (内部函数)
// 注意：此函数假定输入字符串已经过预处理
func jaroDistance(s1, s2 []rune) float64 {
	len1 := len(s1)
	len2 := len(s2)

	if len1 == 0 && len2 == 0 {
		return 1.0
	}
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	// 匹配窗口大小
	matchWindow := int(math.Max(float64(len1), float64(len2))/2) - 1
	if matchWindow < 0 { // 确保匹配窗口至少为 0
		matchWindow = 0
	}

	// 标记匹配的字符
	s1Matches := make([]bool, len1)
	s2Matches := make([]bool, len2)

	matchingCharsCount := 0
	for i := 0; i < len1; i++ {
		start := int(math.Max(0.0, float64(i-matchWindow)))
		end := int(math.Min(float64(len2-1), float64(i+matchWindow)))

		for j := start; j <= end; j++ {
			if !s2Matches[j] && s1[i] == s2[j] {
				s1Matches[i] = true
				s2Matches[j] = true
				matchingCharsCount++
				break
			}
		}
	}

	if matchingCharsCount == 0 {
		return 0.0
	}

	// 收集匹配的字符序列
	s1MatchedChars := make([]rune, 0, matchingCharsCount)
	s2MatchedChars := make([]rune, 0, matchingCharsCount)

	for i := 0; i < len1; i++ {
		if s1Matches[i] {
			s1MatchedChars = append(s1MatchedChars, s1[i])
		}
	}
	for i := 0; i < len2; i++ {
		if s2Matches[i] {
			s2MatchedChars = append(s2MatchedChars, s2[i])
		}
	}

	// 计算错位次数 (transpositions)
	transpositions := 0
	for i := 0; i < matchingCharsCount; i++ {
		if s1MatchedChars[i] != s2MatchedChars[i] {
			transpositions++
		}
	}
	transpositions /= 2

	// 计算 Jaro 相似度
	jaro := (float64(matchingCharsCount)/float64(len1) +
		float64(matchingCharsCount)/float64(len2) +
		float64(matchingCharsCount-transpositions)/float64(matchingCharsCount)) / 3.0

	return jaro
}

// JaroSimilarity 计算两个字符串的 Jaro 相似度
func (ss *StringSimilarity) JaroSimilarity(s1, s2 string) float64 {
	ps1 := []rune(ss.preprocessString(s1))
	ps2 := []rune(ss.preprocessString(s2))

	// 获取 Jaro 相似度
	return jaroDistance(ps1, ps2)
}

// JaroWinklerSimilarity 计算两个字符串的 Jaro-Winkler 相似度
// p 是缩放因子，通常为 0.1 (默认)
func (ss *StringSimilarity) JaroWinklerSimilarity(s1, s2 string, p ...float64) float64 {
	ps1 := []rune(ss.preprocessString(s1))
	ps2 := []rune(ss.preprocessString(s2))

	// 获取 Jaro 相似度
	jaro := jaroDistance(ps1, ps2)

	// 获取缩放因子 p，如果未提供则使用默认值 0.1
	var scalingFactor float64
	if len(p) > 0 {
		scalingFactor = p[0]
	} else {
		scalingFactor = 0.1
	}

	// 计算公共前缀长度 (最多 4 个字符)
	prefixLen := 0
	for i := 0; i < int(math.Min(float64(len(ps1)), float64(len(ps2)))); i++ {
		if ps1[i] == ps2[i] {
			prefixLen++
			if prefixLen == 4 { // Winkler 算法通常只考虑最多 4 个字符的前缀
				break
			}
		} else {
			break
		}
	}

	// Jaro-Winkler 公式
	return jaro + float64(prefixLen)*scalingFactor*(1.0-jaro)
}

// HammingDistance 计算两个字符串之间的汉明距离(内部函数)。
// 注意：此函数假定输入字符串已经过预处理
// 注意：该算法仅适用于等长字符串。如果字符串长度不一致，将返回 -1 表示错误。
func hammingDistance(ps1, ps2 []rune) int {
	len1 := len(ps1)
	len2 := len(ps2)

	if len1 != len2 {
		return -1 // 字符串长度不一致，无法计算汉明距离
	}

	distance := 0
	for i := 0; i < len1; i++ {
		if ps1[i] != ps2[i] {
			distance++
		}
	}
	return distance
}

// HammingSimilarity 计算两个字符串的汉明相似度。
// 注意：该算法仅适用于**等长**字符串。如果字符串长度不一致，将返回 0.0。
func (ss *StringSimilarity) HammingSimilarity(s1, s2 string) float64 {
	ps1 := []rune(ss.preprocessString(s1))
	ps2 := []rune(ss.preprocessString(s2))

	len1 := len(ps1)
	len2 := len(ps2)

	if len1 == 0 && len2 == 0 {
		return 1.0 // 两个空字符串认为完全相似
	}
	if len1 != len2 {
		return 0.0 // 长度不一致，汉明相似度为 0
	}
	if len1 == 0 { // 此时len2也为0，被上面的情况处理了，这里只是为了代码的完整性
		return 1.0
	}

	distance := hammingDistance(ps1, ps2) // 内部会再次预处理，并处理长度不一致的情况

	// 如果 HammingDistance 返回 -1 (长度不一致)，则 Similarity 为 0.0
	if distance == -1 {
		return 0.0
	}

	return (float64(len1) - float64(distance)) / float64(len1)
}

// computeSimhash 计算给定字符串的 SimHash 指纹 (uint64)
func computeSimhash(ps []rune, ngramSize, simhashBits int) uint64 {
	if len(ps) == 0 {
		return 0 // 空字符串的 SimHash 可以约定为 0
	}

	// 1. 生成 N-grams 和它们的频率（作为权重）
	ngramCounts := generateNgramCounts(string(ps), ngramSize)

	// 2. 初始化 D 维向量
	v := make([]int, simhashBits)

	// 3. 为每个 N-gram 生成哈希值并加权叠加
	for gram, weight := range ngramCounts {
		h := fnv.New64a() // 使用 FNV-1a 64位哈希
		h.Write([]byte(gram))
		hashVal := h.Sum64()

		// 遍历哈希值的每一位
		for i := 0; i < simhashBits; i++ {
			if (hashVal>>uint(i))&1 == 1 { // 如果第 i 位是 1
				v[i] += weight
			} else { // 如果第 i 位是 0
				v[i] -= weight
			}
		}
	}

	// 4. 生成 SimHash 指纹
	var simhash uint64
	for i := 0; i < simhashBits; i++ {
		if v[i] > 0 {
			simhash |= (1 << uint(i))
		}
	}

	return simhash
}

// SimhashHammingDistance 计算两个 SimHash 指纹之间的汉明距离
func simhashHammingDistance(hash1, hash2 uint64, simhashBits int) int {
	xorResult := hash1 ^ hash2
	distance := 0
	for i := 0; i < simhashBits; i++ {
		if (xorResult>>uint(i))&1 == 1 {
			distance++
		}
	}
	return distance
}

// SimhashSimilarity 计算两个字符串的 SimHash 相似度
// 返回值在 0.0 到 1.0 之间 (1.0 表示完全相同)
func (ss *StringSimilarity) SimhashSimilarity(s1, s2 string) float64 {
	ps1 := []rune(ss.preprocessString(s1))
	ps2 := []rune(ss.preprocessString(s2))

	hash1 := computeSimhash(ps1, ss.config.NgramSize, ss.config.SimhashBits)
	hash2 := computeSimhash(ps2, ss.config.NgramSize, ss.config.SimhashBits)

	distance := float64(simhashHammingDistance(hash1, hash2, ss.config.SimhashBits))
	return (float64(ss.config.SimhashBits) - distance) / float64(ss.config.SimhashBits)
}

func StrSim(str1, str2 string, algo ...string) float64 {
	var algorithm string
	if len(algo) == 0 {
		algorithm = "jaccard"
	} else {
		algorithm = algo[0]
	}
	ss := NewStringSimilarity()
	switch algorithm {
	case "jaccard":
		return ss.JaccardSimilarity(str1, str2)
	case "cosine":
		return ss.CosineSimilarity(str1, str2)
	case "levenshtein":
		return ss.LevenshteinSimilarity(str1, str2)
	case "jaro":
		return ss.JaroSimilarity(str1, str2)
	case "jaro-winkler":
		return ss.JaroWinklerSimilarity(str1, str2)
	case "hamming":
		return ss.HammingSimilarity(str1, str2)
	case "simhash":
		return ss.SimhashSimilarity(str1, str2)
	default:
		return ss.JaccardSimilarity(str1, str2)
	}
}
