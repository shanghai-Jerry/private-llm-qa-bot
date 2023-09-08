from modelscope.outputs import OutputKeys
from modelscope.pipelines import pipeline
from modelscope.utils.constant import Tasks
import argparse

p = pipeline(
    task=Tasks.document_segmentation,
    model='damo/nlp_bert_document-segmentation_chinese-base')


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--input_file', type=str,
                        default='/Users/youchaojiang/workspace/gopath/src/baidu/easydata/knowledgebase/cmd/kb/output.txt', help='input file')

    args = parser.parse_args()
    file_path = args.input_file
    with open(file_path, "rb") as f:
        content = f.read().decode("utf-8")

    result = p(documents=content)
    print("output:{}".format(result[OutputKeys.TEXT]))
