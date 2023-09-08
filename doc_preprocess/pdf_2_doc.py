from pdf2docx import Converter
import argparse
import os


def pdf2docx(pdf_file, docx_file):
    # convert pdf to docx
    cv = Converter(pdf_file)
    cv.convert(docx_file)      # all pages by default
    cv.close()


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--input_path', type=str,
                        default='./b_data', help='input file path')
    parser.add_argument('--output_path', type=str,
                        default='./pdf2doc', help='output file path')
    args = parser.parse_args()
    data_dir = args.input_path
    test_result_path = args.output_path
    file_dir_list = os.listdir(data_dir)
    for file_dir in file_dir_list:
        if not os.path.isdir(os.path.join(data_dir, file_dir)):
            continue
        file_list = os.listdir(os.path.join(data_dir, file_dir))
        file_result_path = os.path.join(test_result_path, file_dir)
        os.makedirs(file_result_path, exist_ok=True)
        print("=" * 50)
        for file_name in file_list:
            file_path = os.path.join(
                data_dir, os.path.join(file_dir, file_name))
            pdf2docx(file_path, os.path.join(
                file_result_path, file_name + ".docx"))
