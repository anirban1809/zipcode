import json
import subprocess
import argparse


def run() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--message", type=str, required=True)
    parser.add_argument("--path", type=str, required=True)

    args = parser.parse_args()

    try:
        with open(args.path, 'r') as file:
            content = file.read()   

        print(json.dumps({
            "ok": True,
            "file_content":content
        }))

    except Exception as e:
        print(json.dumps({
            "ok": False,
            "error": str(e)
        }))


run()