from FlagEmbedding import FlagReranker
# 设置 fp16 为True可以加快推理速度，效果会有可以忽略的下降
reranker = FlagReranker('BAAI/bge-reranker-large', use_fp16=True)

score = reranker.compute_score(
    ['就业创业', '《就业创业证》申领-信息查询A: 受理机构：市民之家2楼E区人社医保综合服务区27号综合窗口。办理地点：湖北省黄石市开发区•铁山区金山街办园博大道289黄石市民之家。办理时间：周一至周五上午8:30-12:00，下午14：00-17:00。国家法定节假日休息。咨询电话：0714-6521000。是否收费：否。投诉电话：0714-6510992。办理项承诺时限：1。跑腿次数：0跑腿。事项访问地址'])  # 计算 query 和 passage的相似度
print(score)

scores = reranker.compute_score(
    [['query 1', 'passage 1'], ['query 2', 'passage 2']])
print(scores)
