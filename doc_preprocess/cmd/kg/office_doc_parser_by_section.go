package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/go-uuid"
)

// 按照一页处理
type OfficePageParserBySection struct {
	// pdf 全部 txt
	DocOutputTextTotalF *os.File
	// page total
	OutputTextTotalF          *os.File
	OutputTxtF                *os.File
	LogF                      *os.File
	IngnoreHeaderAfterPageOne bool
	Resp                      *OfficeJSONData
	// page table seq
	TableIndex int
	// output
	OutputJsons []*OutputJson
	DocTitles   []string
	// OutputJsonFormat
	OutputJsonFormat *OutputJsonFormat
	CombineF         func(*OutputJsonFormat, *OutputJsonFormat)
}

func CombinOutJsonFormatF(final *OutputJsonFormat, pageOut *OutputJsonFormat) {
	// 只保留第一页已识别有文本标题的标题内容
	if len(final.DocTitles) == 0 {
		final.DocTitles = append(final.DocTitles, pageOut.DocTitles...)
	}
	final.OutputJson = append(final.OutputJson, pageOut.OutputJson...)
}

func NewOfficePageParserBySection(resp *OfficeJSONData) *OfficePageParserBySection {
	return &OfficePageParserBySection{
		CombineF: CombinOutJsonFormatF,
	}
}

func (p *OfficePageParserBySection) SetCombineFinalFunc(f func(*OutputJsonFormat, *OutputJsonFormat)) {
	p.CombineF = f
}

func (p *OfficePageParserBySection) officeDataParseLayoutJsonFromatBySection() {
	totalF := p.OutputTextTotalF
	f := p.OutputTxtF
	docOutF := p.DocOutputTextTotalF
	ignoreHeader := p.IngnoreHeaderAfterPageOne
	resp := p.Resp

	sections := resp.Sections

	// 过滤出所有的header 中idx，全局按top位置排序输出
	var headerSections []Sections
	// var numberSections []Sections
	for _, section := range sections {
		attr := section.Attribute
		switch attr {
		case section_header:
			headerSections = append(headerSections, section)
		case section_number:
			// p.get_output_json_by_section_number(section)
		}
	}
	for _, section := range sections {
		attr := section.Attribute
		switch attr {
		case section_header:
			if !ignoreHeader {
				p.get_output_json_by_section_header(section)
			}
		case section_number:
			// p.get_output_json_by_section_number(section)
			// 所有该section下的para_idx, 应该包含了所有layout中的数据
		case section_section:
			p.get_output_json_by_section_section(section)
		}
	}

	outPutJsons := p.OutputJsons
	docTitles := p.DocTitles
	for _, o := range outPutJsons {
		o.Tokens = CountTokens(o.Content)
		endWith := endsWithPunctuation(o.Content)
		outContent := o.Content
		if endWith {
			outContent = o.Content + "\n"
		}
		// write txt
		if f != nil {
			writeFileBytes(f, []byte(outContent))
		}
		if totalF != nil {
			writeFileBytes(totalF, []byte(outContent))
		}
		if docOutF != nil {
			writeFileBytes(docOutF, []byte(outContent))
		}
	}
	formatOutJson := &OutputJsonFormat{
		DocTitles:  docTitles,
		OutputJson: outPutJsons,
	}
	p.OutputJsonFormat = formatOutJson
}

func (p *OfficePageParserBySection) get_output_json_by_section_header(section Sections) {
	resp := p.Resp
	results := resp.Results
	idx := section.SecIdx.Idx
	// parasIdx := section.SecIdx.ParaIdx
	var outString []string
	var pageNo int
	fmt.Printf("header Idx: %v \n", idx)
	fmt.Printf("header sorted Idx: %v \n", idx)
	sort.Slice(idx, func(i, j int) bool {
		vi := idx[i]
		vj := idx[j]
		return results[vi].Words.WordsLocation.Top < results[vj].Words.WordsLocation.Top
	})
	for _, ridx := range idx {
		result := results[ridx]
		text := result.Words.Word
		pageNo = result.PageNo
		outString = append(outString, text)
	}
	ret := &OutputJson{
		Type:    section_header,
		Pages:   pageNo,
		Content: strings.Join(outString, "\n"),
	}
	p.OutputJsons = append(p.OutputJsons, ret)
}

func (p *OfficePageParserBySection) get_output_json_by_section_number(section Sections) {
	resp := p.Resp
	results := resp.Results
	idx := section.SecIdx.Idx
	var outString []string
	var pageNo int
	for _, ridx := range idx {
		result := results[ridx]
		text := result.Words.Word
		pageNo = result.PageNo
		outString = append(outString, text)
	}
	ret := &OutputJson{
		Type:    section_number,
		Pages:   pageNo,
		Content: strings.Join(outString, "\n"),
	}
	p.OutputJsons = append(p.OutputJsons, ret)
}

func (p *OfficePageParserBySection) get_output_json_by_section_section(section Sections) {
	resp := p.Resp
	var docTitles []string
	results := resp.Results
	layouts := resp.Layouts
	// idx := section.SecIdx.Idx
	var pageNo int
	var outPutJsons []*OutputJson
	parasIdx := section.SecIdx.ParaIdx
	fmt.Printf("parasIdx: %v \n", parasIdx)
	// 需要根据layout的id来找到对应的layout，并按照layout_location来排序
	// top距离值越大的也在后面
	sort.Slice(parasIdx, func(i, j int) bool {
		vi := parasIdx[i]
		vj := parasIdx[j]
		return layouts[vi].LayoutLocation[0].Y <= layouts[vj].LayoutLocation[0].Y
	})
	fmt.Printf("sorted parasIdx: %v \n", parasIdx)
	for _, pidx := range parasIdx {
		lout := layouts[pidx]
		layoutType := lout.Layout
		if layoutType == layout_table ||
			layoutType == layout_content ||
			layoutType == layout_figure {
			switch layoutType {
			// 图表(识别出的文本丢掉)
			case layout_figure:
				imageIdStr, _ := uuid.GenerateUUID()
				outputJson := &OutputJson{
					Type:    layout_figure,
					Pages:   pageNo,
					Content: fmt.Sprintf("[image: %v] == start ", imageIdStr),
				}
				outPutJsons = addOutputJson(outPutJsons, outputJson)
				// 识别出来的图片内容，拿出来看看
				// for _, lidx := range lout.LayoutIdx {
				// 	ridx := lidx
				// 	result := results[ridx]
				// 	text := result.Words.Word
				// 	pageNo = result.PageNo
				// 	outputJson := &OutputJson{
				// 		Type:    layout_text,
				// 		Pages:   pageNo,
				// 		Content: text,
				// 	}
				// 	outPutJsons = addOutputJson(outPutJsons, outputJson)
				// }
				// outputJsonEnd := &OutputJson{
				// 	Type:    layout_figure,
				// 	Pages:   pageNo,
				// 	Content: fmt.Sprintf("[image: %v] == end ", imageIdStr),
				// }
				// outPutJsons = addOutputJson(outPutJsons, outputJsonEnd)
			// 表格
			case layout_table:
				table := buildTable(resp.TablesResult[p.TableIndex].Body)
				outputJson := &OutputJson{
					Type:    layout_table,
					Pages:   pageNo,
					Content: "\n" + writeTableStringBuffer(table) + "\n",
				}
				outPutJsons = addOutputJson(outPutJsons, outputJson)
				p.TableIndex++
			// 目录
			case layout_content:
				start := lout.LayoutIdx[0]
				end := lout.LayoutIdx[len(lout.LayoutIdx)-1]
				// add contents
				outputJson := &OutputJson{
					Type:    layout_content,
					Pages:   pageNo,
					Content: strings.Join(getContentParas(results, start, end), "\n"),
				}
				outPutJsons = addOutputJson(outPutJsons, outputJson)
			}
			continue
		}
		// 处理layout layoutIdx
		// 0引言
		// [pidx:17] LayoutIdx:[43]
		// 在现有的安全技术中
		// pidx:14] LayoutIdx:[53 55 57 59 61 63 65 67 69 71 73 75 77 79]
		fmt.Printf("--- [pidx:%v] LayoutIdx:%v \n", pidx, lout.LayoutIdx)
		for _, lidx := range lout.LayoutIdx {
			ridx := lidx
			result := results[ridx]
			text := result.Words.Word
			pageNo = result.PageNo
			switch layoutType {
			// 表格标题
			case layout_table_title:
				outputJson := &OutputJson{
					Type:    layout_table_title,
					Pages:   pageNo,
					Content: "\n" + text + "\n",
				}
				outPutJsons = addOutputJson(outPutJsons, outputJson)

				// 正常的文本
			case layout_text:
				outputJson := &OutputJson{
					Type:    layout_text,
					Pages:   pageNo,
					Content: text,
				}
				outPutJsons = addOutputJson(outPutJsons, outputJson)
				// 图标题
			case layout_figure_title:
				outputJson := &OutputJson{
					Type:    layout_figure_title,
					Pages:   pageNo,
					Content: "\n" + text + "\n",
				}
				outPutJsons = addOutputJson(outPutJsons, outputJson)
				// 文档标题
			case layout_doc_title:
				// 段落标题: 每个段落都有一个标题
				docTitles = append(docTitles, text)
			case layout_text_title:
				// add title
				outputJson := &OutputJson{
					Type:    layout_title,
					Pages:   pageNo,
					Content: "\n##" + text + "## \n",
				}
				outPutJsons = addOutputJson(outPutJsons, outputJson)
			}
		}
	}
	// output by page
	p.OutputJsons = append(p.OutputJsons, outPutJsons...)
	p.DocTitles = append(p.DocTitles, docTitles...)
}
