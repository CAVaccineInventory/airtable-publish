#!/usr/bin/env python3

import fire
import re
from pprint import pprint

pattern = re.compile(r"p\[\"([\w+ \?]+)\"\]")


def parse_use(scriptname: str):
    """This a brittle script to extract all `p["Field name"]` from the `data.js` file."""
    with open(scriptname, "r") as fh:
        fields = set(
            [match.group(1) for line in fh for match in pattern.finditer(line)]
        )
    fields = list(fields)
    fields.sort()
    for field in fields:
        print(f'"{field}":  1,')


if __name__ == "__main__":
    fire.Fire(parse_use)
