from unstructured.partition.auto import partition
import json
from unstructured.chunking.title import chunk_by_title
from typing import Callable, List
from unstructured.documents.elements import (
    CompositeElement,
    Element,
    ElementMetadata,
    Table,
    Text,
    Title,
)
import argparse


def writeFile(fileName, content, fileFormat):
    out_f = open("./unstructured-io/{}.{}".format(fileName, fileFormat), 'w')
    out_f.write(content)
    out_f.close()


def writeElements(elements: List[Element], fileName: str):
    contents = []
    outJson = {}
    objects = []
    for el in elements:
        contents.append(str(el))
        objects.append(el.to_dict())
    # txt
    writeFile(fileName, "\n\n".join(contents), "txt")
    # json
    outJson["results"] = objects
    outJsonStr = json.dumps(outJson, ensure_ascii=False)
    writeFile(fileName, outJsonStr, "json")


def parser_pdf(pdfpath):
    fileName = pdfpath.split("/")[-1]
    elements = partition(filename=pdfpath, content_type="application/pdf")

    # origin
    writeElements(elements, fileName)
    # chunks
    elements = chunk_by_title(
        elements, combine_under_n_chars=384, new_after_n_chars=384)
    writeElements(elements, "{}.chunked".format((fileName)))


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--input_pdf', type=str,
                        default='./b_data/pdf/BCD-621WDCAU10918poc测试版.pdf', help='input file')
    args = parser.parse_args()
    input_file = args.input_pdf
    parser_pdf(input_file)
