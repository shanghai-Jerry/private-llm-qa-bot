#!/usr/bin/env python
# -*- coding: utf-8 -*-

import json
from flask import Flask, request

app = Flask(__name__)


@app.route('/check', methods=['POST'])
def check():
    """
    check
    :return:
    """
    req = json.loads(request.data)
    app.logger.info(req)
    results = {
        "code": 0,
    }
    return json.dumps(results)


@app.route('/get', methods=['GET'])
def get_api():
    """
    get_api
    :return:
    """
    q = request.args.get('query')
    res = {
        "q": q,
        "results": {
            "label": {"key": 1, "value": 2},
            "others": {
                "value": "value",
            },
        }
    }
    return json.dumps(res)


@app.route('/check/health', methods=['GET'])
def check_health():
    """
    check health
    :return:
    """
    return "hello, world"


if __name__ == '__main__':
    app.run(host='127.0.0.1', port=8848, debug=True)
