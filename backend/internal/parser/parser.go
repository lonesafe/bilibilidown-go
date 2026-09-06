package parser

import (
	"fmt"
	"regexp"
	"strings"

	"bilibilidown-go/internal/api"
	"bilibilidown-go/internal/model"
)

// Parser interface for parsing different Bilibili URL formats
type Parser interface {
	Matches(input string) bool
	ValidID(input string) string
	GetResult(input string, qn int, fetchLinks bool) (*model.VideoInfo, error)
}

// BaseParser provides shared logic
type BaseParser struct {
	api *api.BilibiliAPI
}

// BVParser handles BV IDs
type BVParser struct {
	BaseParser
	pattern *regexp.Regexp
}

func NewBVParser(bapi *api.BilibiliAPI) *BVParser {
	return &BVParser{
		BaseParser: BaseParser{api: bapi},
		pattern:    regexp.MustCompile(`BV[0-9A-Za-z]+`),
	}
}

func (p *BVParser) Matches(input string) bool {
	return p.pattern.MatchString(input)
}

func (p *BVParser) ValidID(input string) string {
	m := p.pattern.FindString(input)
	return m
}

func (p *BVParser) GetResult(input string, qn int, fetchLinks bool) (*model.VideoInfo, error) {
	bvid := p.ValidID(input)
	if bvid == "" {
		return nil, fmt.Errorf("无法解析BV号")
	}
	return p.api.GetVideoDetail(bvid, qn, fetchLinks)
}

// AVParser handles av IDs
type AVParser struct {
	BaseParser
	pattern *regexp.Regexp
}

func NewAVParser(bapi *api.BilibiliAPI) *AVParser {
	return &AVParser{
		BaseParser: BaseParser{api: bapi},
		pattern:    regexp.MustCompile(`av(\d+)`),
	}
}

func (p *AVParser) Matches(input string) bool {
	return p.pattern.MatchString(strings.ToLower(input))
}

func (p *AVParser) ValidID(input string) string {
	return p.pattern.FindString(strings.ToLower(input))
}

func (p *AVParser) GetResult(input string, qn int, fetchLinks bool) (*model.VideoInfo, error) {
	avid := p.pattern.FindStringSubmatch(strings.ToLower(input))
	if len(avid) < 2 {
		return nil, fmt.Errorf("无法解析av号")
	}
	// Convert av to bv
	avidNum := parseInt(avid[1])
	bvid := av2Bv(avidNum)
	return p.api.GetVideoDetail(bvid, qn, fetchLinks)
}

// B23Parser handles b23.tv short links
type B23Parser struct {
	BaseParser
	pattern *regexp.Regexp
}

func NewB23Parser(bapi *api.BilibiliAPI) *B23Parser {
	return &B23Parser{
		BaseParser: BaseParser{api: bapi},
		pattern:    regexp.MustCompile(`b23\.tv/[a-zA-Z0-9]+`),
	}
}

func (p *B23Parser) Matches(input string) bool {
	return p.pattern.MatchString(input)
}

func (p *B23Parser) ValidID(input string) string {
	return p.pattern.FindString(input)
}

func (p *B23Parser) GetResult(input string, qn int, fetchLinks bool) (*model.VideoInfo, error) {
	// b23 short links need HTTP redirect resolution - simplified: extract bvid from URL
	shortURL := p.ValidID(input)
	if !strings.HasPrefix(shortURL, "http") {
		shortURL = "https://" + shortURL
	}
	// In production, follow redirect to get real URL
	// For now, treat as BV if possible
	bvPattern := regexp.MustCompile(`BV[0-9A-Za-z]+`)
	if m := bvPattern.FindString(input); m != "" {
		return p.api.GetVideoDetail(m, qn, fetchLinks)
	}
	return nil, fmt.Errorf("b23短链需要重定向解析: %s", shortURL)
}

// SSParser handles season (ss) links for bangumi
type SSParser struct {
	BaseParser
	pattern *regexp.Regexp
}

func NewSSParser(bapi *api.BilibiliAPI) *SSParser {
	return &SSParser{
		BaseParser: BaseParser{api: bapi},
		pattern:    regexp.MustCompile(`ss(\d+)`),
	}
}

func (p *SSParser) Matches(input string) bool {
	return p.pattern.MatchString(strings.ToLower(input))
}

func (p *SSParser) ValidID(input string) string {
	return p.pattern.FindString(strings.ToLower(input))
}

func (p *SSParser) GetResult(input string, qn int, fetchLinks bool) (*model.VideoInfo, error) {
	return nil, fmt.Errorf("番剧(ss)解析暂未实现")
}

// EPParser handles episode (ep) links
type EPParser struct {
	BaseParser
	pattern *regexp.Regexp
}

func NewEPParser(bapi *api.BilibiliAPI) *EPParser {
	return &EPParser{
		BaseParser: BaseParser{api: bapi},
		pattern:    regexp.MustCompile(`ep(\d+)`),
	}
}

func (p *EPParser) Matches(input string) bool {
	return p.pattern.MatchString(strings.ToLower(input))
}

func (p *EPParser) ValidID(input string) string {
	return p.pattern.FindString(strings.ToLower(input))
}

func (p *EPParser) GetResult(input string, qn int, fetchLinks bool) (*model.VideoInfo, error) {
	return nil, fmt.Errorf("番剧(ep)解析暂未实现")
}

// MLParser handles media list (ml) links
type MLParser struct {
	BaseParser
	pattern *regexp.Regexp
}

func NewMLParser(bapi *api.BilibiliAPI) *MLParser {
	return &MLParser{
		BaseParser: BaseParser{api: bapi},
		pattern:    regexp.MustCompile(`ml(\d+)`),
	}
}

func (p *MLParser) Matches(input string) bool {
	return p.pattern.MatchString(strings.ToLower(input))
}

func (p *MLParser) ValidID(input string) string {
	return p.pattern.FindString(strings.ToLower(input))
}

func (p *MLParser) GetResult(input string, qn int, fetchLinks bool) (*model.VideoInfo, error) {
	return nil, fmt.Errorf("收藏夹(ml)解析暂未实现")
}

// InputParser selects the right parser for a given input
type InputParser struct {
	parsers []Parser
}

// NewInputParser creates an input parser with all registered parsers
func NewInputParser(bapi *api.BilibiliAPI) *InputParser {
	return &InputParser{
		parsers: []Parser{
			NewBVParser(bapi),
			NewAVParser(bapi),
			NewB23Parser(bapi),
			NewSSParser(bapi),
			NewEPParser(bapi),
			NewMLParser(bapi),
		},
	}
}

// SelectParser finds the matching parser for input
func (ip *InputParser) SelectParser(input string) Parser {
	for _, p := range ip.parsers {
		if p.Matches(input) {
			return p
		}
	}
	return nil
}

// ValidID extracts the valid video ID from input
func (ip *InputParser) ValidID(input string) string {
	p := ip.SelectParser(input)
	if p != nil {
		return p.ValidID(input)
	}
	return ""
}

// GetResult parses input and returns video info
func (ip *InputParser) GetResult(input string, qn int, fetchLinks bool) (*model.VideoInfo, error) {
	p := ip.SelectParser(input)
	if p == nil {
		return nil, fmt.Errorf("无法识别的链接格式: %s", input)
	}
	return p.GetResult(input, qn, fetchLinks)
}
