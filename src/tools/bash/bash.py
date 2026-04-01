import json
import subprocess
import shlex
import argparse


def run() -> None:
    """
    Executes a bash tool call based on the provided payload.

    Expected payload structure:
    {
        "message": str,
        "command": str,
        "working_directory": str (optional),
        "timeout_seconds": int (optional)
    }
    """

    parser = argparse.ArgumentParser()

    parser.add_argument("--message", type=str)
    parser.add_argument("--command", type=str)
    parser.add_argument("-w","--working_directory", type=str, default=".")
    parser.add_argument("-t","--timeout_seconds", type=int, default=10)

    args = parser.parse_args()

    # Validate required fields
    if "message" not in args or "command" not in args:
        print({
            "ok": False,
            "error": "Missing required fields: message, command"
        })

        exit(1)

    command = args.command # pyright: ignore[reportIndexIssue]
    cwd = args.working_directory
    timeout = args.timeout_seconds

    try:
        # Use shell=False with shlex.split for safety
        result = subprocess.run(
            shlex.split(command),
            shell=False,
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=timeout
        )

        print(json.dumps({
            "ok": True,
            "exit_code": result.returncode,
            "stdout": result.stdout,
            "stderr": result.stderr
        }))

    except subprocess.TimeoutExpired as e:
        print(json.dumps({
            "ok": False,
            "error": "Command timed out",
            "stdout": e.stdout,
            "stderr": e.stderr
        }))

    except Exception as e:
        print(json.dumps({
            "ok": False,
            "error": str(e)
        }))


run()