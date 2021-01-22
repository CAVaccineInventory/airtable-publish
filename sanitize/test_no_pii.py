#!/usr/bin/env python3
from subprocess import check_output
import json


def test_no_personal_info():
    input_file = "test_data/reduced_locations.json"
    stdout = check_output(["go", "run", "main.go", input_file])

    # Basic sanity check.
    assert b"@gmail.com" not in stdout

    result = json.loads(stdout)
    for loc in result:
        assert "Last report author" not in loc, loc["Last report author"]

        # Check no "Internal notes" field
        assert "Internal notes" not in loc


def test_no_personal_info():
    input_file = "test_data/reduced_locations.json"
    stdout = check_output(["go", "run", "main.go", input_file])

    with open(input_file, "r") as fh:
        input_data = json.load(fh)
    result = json.loads(stdout)
    assert len(input_data) == len(result)