import boto3
import json
import os
import re
import statistics
import argparse
from bs4 import BeautifulSoup
from pdf2image import convert_from_path
from langchain.document_loaders import PDFMinerPDFasHTMLLoader
from langchain.docstore.document import Document
from langchain.text_splitter import RecursiveCharacterTextSplitter, CharacterTextSplitter
from langchain.docstore.document import Document

from tqdm import tqdm


class Elembbox(object):
    left = -1
    top = -1
    width = -1
    height = -1
    right = -1
    bottom = -1

    # top 增加是往下 bottom > top
    # right 增加是往右 right > left

    def __init__(self, left, top, width, height):
        self.left = left
        self.top = top
        self.width = width
        self.height = height
        self.right = left + width
        self.bottom = top + height

    def __str__(self):
        return "left:{}, top:{}, right:{}, bottom:{}, width:{}, height:{}".format(self.left, self.top, self.right, self.bottom, self.width, self.height)

    def to_span(self):
        return """<span style="position:absolute; border: red 1px solid; left:{}px; top:{}px; width:{}px; height:{}px;"></span>""".format(self.left, self.top, self.width, self.height)

    def overlap_area(self, bbox, h_margin=0, v_margin=0):
        if bbox is None:
            return 0

        # 解包bbox坐标
        x1_1, y1_1, x2_1, y2_1 = self.left-h_margin, self.top - \
            v_margin, self.right+h_margin, self.bottom+v_margin
        x1_2, y1_2, x2_2, y2_2 = bbox.left-h_margin, bbox.top - \
            v_margin, bbox.right+h_margin, bbox.bottom+v_margin
        # print((x1_1, y1_1, x2_1, y2_1))
        # print((x1_2, y1_2, x2_2, y2_2))

        # 计算重叠区域的坐标
        x_overlap = max(0, min(x2_1, x2_2) - max(x1_1, x1_2))
        y_overlap = max(0, min(y2_1, y2_2) - max(y1_1, y1_2))

        # print("x_overlap: {}".format(x_overlap))
        # print("y_overlap: {}".format(y_overlap))
        # 计算重叠面积
        area = x_overlap * y_overlap

        return area

    def get_area(self):
        return self.width * self.height

    def is_overlap(self, other):
        return self.overlap_area(other) > 0

    def is_beside(self, other):
        # only horizontal direction
        return self.overlap_area(other, h_margin=0, v_margin=8)


class ElemTable(object):
    table_bbox = None
    table_title = None
    table_footer = None
    table_data = None

    # top 增加是往下 bottom > top
    # right 增加是往右 right > left

    def __init__(self, bbox, title, footer, data):
        self.table_bbox = bbox
        self.table_title = title
        self.table_footer = footer
        self.table_data = data

    def __str__(self):
        x = {
            "title": self.table_title,
            "footer": self.table_footer,
            "data": self.table_data
        }
        return json.dumps(x)

    def max_cell_area(self):
        max_area = 0
        assert (self.table_data)

        for cell in self.table_data:
            cell_area = cell['bbox'].get_area()
            max_area = max(cell_area, max_area)
        return max_area

    def to_readable(self):
        assert (self.table_data)
        cell_list = self.table_data

        max_row_idx = max([cell['row_index'] for cell in cell_list])
        max_col_idx = max([cell['col_index'] for cell in cell_list])

        header_cells = [
            cell for cell in cell_list if cell['is_header'] == True]

        if len(header_cells) == 0:
            for cell in cell_list:
                if cell['row_index'] == 1:
                    cell["is_header"] = True
                    header_cells.append(cell)

        all_header_text = "".join([cell['text'] for cell in header_cells])

        head_dict = {str(cell['col_index']): cell['text']
                     for cell in header_cells}

        header_row_idx = statistics.mode(
            cell['row_index'] for cell in header_cells)

        data_cells = [
            cell for cell in cell_list if cell['row_index'] > header_row_idx]

        row_cnt = max([cell['row_index']
                      for cell in data_cells]) - header_row_idx

        data = [{} for i in range(row_cnt)]

        for cell in data_cells:
            idx = cell['row_index'] - header_row_idx - 1
            key = head_dict.get(str(cell['col_index']), 'row_key')
            # print(f"{idx}, {key}, {cell['text']}")
            data[idx][key] = cell['text']

        ret = {
            "table": self.table_title.get("text", ""),
            "footer": self.table_footer.get("text", ""),
            "data": None if len(all_header_text) == 0 else data
        }

        return ret


def find_page_bbox(html_soup):
    all_span = html_soup.find_all('span')
    # <span style="position:absolute; border: gray 1px solid; left:0px; top:50px; width:612px; height:792px;"></span>

    page_bboxes = []
    for idx, c in enumerate(all_span):
        style_attribute = c.get('style')

        attr_list = [p.split(':')
                     for p in style_attribute.strip(";").split('; ')]
        span_pos = {k: v for k, v in attr_list if k in [
            'left', 'top', 'width', 'height', 'font-size', 'border']}
        keys = span_pos.keys()
        if 'border' in keys and span_pos['border'].strip() == 'gray 1px solid':
            left = int(span_pos['left'][:-2])
            top = int(span_pos['top'][:-2])
            width = int(span_pos['width'][:-2])
            height = int(span_pos['height'][:-2])
            page_bboxes.append(Elembbox(left, top, width, height))

    return page_bboxes


def conver_pos_ratio2pixel(page_left_pixel, page_top_pixel, page_width_pixel, page_height_pixel, table_bbox_ratio):
    left_pixel = round(
        table_bbox_ratio['Left'] * page_width_pixel + page_left_pixel)
    top_pixel = round(table_bbox_ratio['Top']
                      * page_height_pixel + page_top_pixel)
    width_pixel = round(table_bbox_ratio['Width'] * page_width_pixel)
    height_pixel = round(table_bbox_ratio['Height'] * page_height_pixel)
    return Elembbox(left_pixel, top_pixel, width_pixel, height_pixel)


def get_cell_info(info_dict, id):
    cell_info = info_dict[id]
    row_index = cell_info['RowIndex']
    col_index = cell_info['ColumnIndex']
    row_span = cell_info['RowSpan']
    col_span = cell_info['ColumnSpan']
    bbox = cell_info['Geometry']['BoundingBox']
    header_flag = cell_info.get('EntityTypes', [''])
    is_header = header_flag[0] == "COLUMN_HEADER"
    cell_text = ''

    words = []

    if 'Relationships' in cell_info.keys():
        cell_word_node = cell_info['Relationships'][0]
        if cell_word_node['Type'] == 'CHILD':
            for word_id in cell_word_node['Ids']:
                if info_dict[word_id]['BlockType'] == 'WORD':
                    words.append(info_dict[word_id]['Text'])
            cell_text = ' '.join(words)

    ret_obj = {
        "text": cell_text,
        "bbox": bbox,
        "row_index": row_index,
        "col_index": col_index,
        "row_span": row_span,
        "col_span": col_span,
        "is_header": is_header
    }

    return ret_obj


def get_textline(info_dict, textline_id):
    textline_bbox = info_dict[textline_id]['Geometry']['BoundingBox']
    textline_word_ids = info_dict[textline_id]['Relationships'][0]['Ids']
    words = [info_dict[word_id]['Text']
             for word_id in textline_word_ids if info_dict[word_id]['BlockType'] == 'WORD']
    obj = {
        "text": " ".join(words),
        "bbox": textline_bbox
    }
    return obj


def extract_page_tables(response, pages_bbox):

    info_dict = {}
    for block_obj in response['Blocks']:
        info_dict[block_obj['Id']] = block_obj

    table_list = []
    table_bboxes = []
    for key, info in info_dict.items():
        if info['BlockType'] == 'TABLE':
            bbox = info['Geometry']['BoundingBox']
            page_left_pixel = pages_bbox.left
            page_top_pixel = pages_bbox.top
            page_width_pixel = pages_bbox.width
            page_height_pixel = pages_bbox.height
            bbox_pixel = conver_pos_ratio2pixel(
                page_left_pixel, page_top_pixel, page_width_pixel, page_height_pixel, bbox)
            table_bboxes.append(bbox_pixel)

            cell_list = []
            title_obj = {}
            footer_obj = {}
            for item in info['Relationships']:
                if item['Type'] == 'CHILD':
                    for cell_id in item['Ids']:
                        # iterate CELL
                        cell_obj = get_cell_info(info_dict, cell_id)
                        pixel_bbox = conver_pos_ratio2pixel(
                            page_left_pixel, page_top_pixel, page_width_pixel, page_height_pixel, cell_obj['bbox'])
                        cell_obj['bbox'] = pixel_bbox
                        cell_list.append(cell_obj)
                elif item['Type'] == 'TABLE_TITLE':
                    table_title_id = item['Ids'][0]
                    title_obj = get_textline(info_dict, table_title_id)
                    pixel_bbox = conver_pos_ratio2pixel(
                        page_left_pixel, page_top_pixel, page_width_pixel, page_height_pixel, title_obj['bbox'])
                    title_obj['bbox'] = pixel_bbox
                elif item['Type'] == 'TABLE_FOOTER':
                    table_footer_id = item['Ids'][0]
                    footer_obj = get_textline(info_dict, table_footer_id)
                    pixel_bbox = conver_pos_ratio2pixel(
                        page_left_pixel, page_top_pixel, page_width_pixel, page_height_pixel, footer_obj['bbox'])
                    footer_obj['bbox'] = pixel_bbox

            table_info = ElemTable(bbox_pixel, title_obj,
                                   footer_obj, cell_list)
            table_list.append(table_info)

    return table_list, table_bboxes


def split_pdf_to_snippet(soup, table_elem_bboxs):
    content = soup.find_all('div')

    cur_fs = 0
    cur_text = None
    snippets = []   # first collect all snippets that have the same font size

    print("table bbox count: {}".format(len(table_elem_bboxs)))
    skip_count = 0

    def overlap_with_table(table_elem_bboxs, div_elem_bbox):
        for elem_bbox in table_elem_bboxs:
            if div_elem_bbox.is_overlap(elem_bbox):
                return True

        return False

    table_text_objs = []

    previous_div_bbox = None
    snippet_start = True
    snippet_follow = False
    snippet_state = snippet_start
    for c in content:
        div_style_attribute = c.get('style')
        attr_list = [p.split(':')
                     for p in div_style_attribute.strip(";").split('; ')]
        div_pos = {
            k: int(v[:-2]) for k, v in attr_list if k in ['left', 'top', 'width', 'height']}
        keys = div_pos.keys()
        if 'left' not in keys or 'top' not in keys or 'width' not in keys or 'height' not in keys:
            continue
        div_elem_bbox = Elembbox(
            div_pos['left'], div_pos['top'], div_pos['width'], div_pos['height'])

        if overlap_with_table(table_elem_bboxs, div_elem_bbox):
            skip_count += 1
            table_text_objs.append({"text": c.text, "bbox": div_elem_bbox})
            continue

        # if these two div is not beside each other
        if not div_elem_bbox.is_beside(previous_div_bbox) and cur_text and cur_fs:
            snippets.append((cur_text, cur_fs, snippet_state))
            cur_fs = 0
            cur_text = None
            snippet_state = snippet_start

        previous_div_bbox = div_elem_bbox

        sp_list = c.find_all('span')
        if not sp_list:
            continue

        for sp in sp_list:
            st = sp.get('style')
            if not st:
                continue
            fs = re.findall('font-size:(\d+)px', st)

            if not fs:
                continue
            fs = int(fs[0])

            if not cur_fs and not cur_text:
                cur_fs = fs
                cur_text = sp.text
                snippet_state = snippet_start
                continue

            if fs == cur_fs:
                cur_text += sp.text
            else:
                snippets.append((cur_text, cur_fs, snippet_state))
                snippet_state = snippet_start if fs > cur_fs else snippet_follow
                cur_fs = fs
                cur_text = sp.text

    snippets.append((cur_text, cur_fs, snippet_follow))

    # merge snippet
    merged_snippets = []
    temp_list = []
    doc_title = ''
    max_font_size = max([item[1] for item in snippets])
    for snippet in snippets:
        if max_font_size == snippet[1]:
            doc_title = snippet[0]

        if len(temp_list) == 0 or snippet[2] == False:
            temp_list.append(snippet)
        else:
            content_list = [item[0] for item in temp_list]
            font_size_list = [item[1] for item in temp_list]
            content = "\n".join(content_list)
            font_size = max(font_size_list)
            temp_list.clear()
            temp_list.append(snippet)
            merged_snippets.append(
                {"content": content, "font_size": font_size})

    print("filter {} table text".format(skip_count))
    return merged_snippets, doc_title, table_text_objs


def split_pdf(soup, all_table_bboxes, page_table_list):
    semantic_snippets, doc_title, table_text_objs = split_pdf_to_snippet(
        soup, all_table_bboxes)
    text_splitter = RecursiveCharacterTextSplitter(
        chunk_size=1024,
        chunk_overlap=0,
        separators=["\n\n", "\n", ".", "。", ",", "，", " "],
    )

    for item in semantic_snippets:
        content = item["content"]
        chunks = text_splitter.create_documents([content])
        for chunk in chunks:
            snippet_info = {
                "content": chunk.page_content,
                "font_size": item["font_size"],
                "doc_title": doc_title
            }
            yield snippet_info

    # mapping the word to cell (if the language is chinese, textract will lose chinese characters)
    for i in range(len(page_table_list)):
        table_max_cell_area = page_table_list[i].max_cell_area()
        for j in range(len(page_table_list[i].table_data)):
            cell_obj = page_table_list[i].table_data[j]
            overlap_area_t1 = 0
            for table_text_obj in table_text_objs:
                # if text's area is bigger than biggest cell area, skip it
                cell_text_area = table_text_obj['bbox'].get_area()
                # print("cell_text_area: {}, table_max_cell_area:{}".format(cell_text_area, table_max_cell_area))
                if cell_text_area > table_max_cell_area:
                    # print("skip text: {}".format(table_text_obj['text']))
                    continue

                overlap_area_t2 = cell_obj['bbox'].overlap_area(
                    table_text_obj['bbox'])
                if overlap_area_t2 > overlap_area_t1:
                    overlap_area_t1 = overlap_area_t2
                    cell_obj['text'] = table_text_obj['text']

    # pdb.set_trace()
    for table in page_table_list:
        table_json = table.to_readable()

        # if there is no data in table, skip it
        if table_json['data'] == None:
            continue

        table_snippet = {
            "content": table_json,
            "doc_title": doc_title
        }

        yield table_snippet


# python3 pdf_spliter.py --input_file  ./b_data/pdf/2.1.pdf --output_path ./b_data/pdf
if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--input_file', type=str,
                        default='./common-stock-fs.pdf', help='input file')
    parser.add_argument('--output_path', type=str,
                        default='./pdf_spliter', help='input file')
    args = parser.parse_args()
    pdf_path = args.input_file
    out_path = args.output_path
    # convert pdf to image
    image_paths = []
    images = convert_from_path(pdf_path, 300)

    output_folder = os.path.basename(pdf_path).replace(".pdf", "")
    file_name = output_folder.split('/')[-1]

    output_folder = out_path + "/" + file_name
    if not os.path.exists(output_folder):
        os.makedirs(output_folder)

    for idx, image in tqdm(enumerate(images)):
        image_path = f'{output_folder}/page-{idx}.png'
        image_paths.append(image_path)
        image.save(image_path)

    # parse to html
    loader = PDFMinerPDFasHTMLLoader(pdf_path)
    data = loader.load()[0]
    html_path = f'{output_folder}/{file_name}.html'
    with open(html_path, 'w') as f:
        f.write(data.page_content)
        f.close()

    soup = BeautifulSoup(data.page_content, 'html.parser')
    page_bboxes = find_page_bbox(soup)

    # extract table by textract
    all_table_bboxes = []
    all_table_list = []
    session = boto3.Session()
    client = session.client('textract')

    for idx, imge_path in tqdm(enumerate(image_paths)):
        with open(imge_path, "rb") as document_data:
            response = client.analyze_document(
                Document={
                    'Bytes': document_data.read(),
                },
                FeatureTypes=["TABLES"])

            table_list, table_bboxes = extract_page_tables(
                response, page_bboxes[idx])
            all_table_list.extend(table_list)
            all_table_bboxes.extend(table_bboxes)

    output_path = f'{output_folder}/{file_name}.pdf.json'
    f_name = "{}".format(output_path)
    out_f = open(f_name, 'w')
    snippet_arr = []
    for snippet_info in split_pdf(soup, all_table_bboxes, all_table_list):
        snippet_arr.append(snippet_info)

    all_info = json.dumps(snippet_arr, ensure_ascii=False)
    out_f.write(all_info)
    out_f.close()
