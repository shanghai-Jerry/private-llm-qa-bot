from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain.document_loaders import TextLoader
import json
from langchain.text_splitter import CharacterTextSplitter
from tqdm import tqdm


def split_documents(file_path):
    loader = TextLoader(file_path)
    documents = loader.load()
    text_splitter = RecursiveCharacterTextSplitter(
        separators=["\n\n", "\n", ".", "。", ",", "，", " "], chunk_size=384, chunk_overlap=150)
    texts = text_splitter.split_documents(documents)
    seg = 1
    for page in tqdm(texts[:]):
        text = page.page_content
        # 将文本保存到文档中
        # text = re.sub('[^\w\s]\n', " ", text)
        # text = re.sub('\n', "", text)
        info = {
            "content": text,
        }
        data_str = json.dumps(info, ensure_ascii=False)
        out_f = open("./custom_test/paras/{}.txt".format(seg), 'w')
        out_f.write(text)
        out_f.close()
        out_f_json = open("./custom_test/paras/{}.json".format(seg), 'w')
        out_f_json.write(data_str)
        out_f_json.close()
        seg += 1


def split(content):
    text_splitter = RecursiveCharacterTextSplitter(
        chunk_size=384,
        chunk_overlap=0,
        separators=["\n\n", "\n", ".", "。", ",", "，", " "],
    )
    chunks = text_splitter.create_documents([content])
    result = []
    for chunk in chunks:
        info = {
            "content": chunk.page_content,
        }
        result.append(info)
    return result


if __name__ == '__main__':

    split_documents(
        "~/Downloads/guowangdata/pdf/output/txt/1-DLT 1586-2016 12kV固体绝缘金属封闭开关设备和控制设备-out.txt")
