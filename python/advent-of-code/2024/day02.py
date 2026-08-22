# Advent of Code 2024 | https://adventofcode.com/2024/day/1
# Copyright © 2024 Giovanni Squillero / Politecnico di Torino
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.


from itertools import pairwise
from icecream import ic

INPUT_FILE = "day02-test.txt"
# INPUT_FILE = "day02-input.txt"


def check_report_safety(report):
    r"""Test wether `report` is safe"""
    return all(1 <= abs(e1 - e2) <= 3 for e1, e2 in pairwise(report)) and (
        report == sorted(report) or report == sorted(report, reverse=True)
    )


def main():
    # --- Part One & Two ---
    safe_count_1 = safe_count_2 = 0
    for line in open(INPUT_FILE):
        report = [int(r) for r in line.split()]

        safe = check_report_safety(report)
        safe_count_1 += safe

        candidate_reports = (report[:i] + report[i + 1 :] for i in range(len(report)))
        safe_count_2 += safe or any(check_report_safety(r) for r in candidate_reports)

    ic(safe_count_1)
    ic(safe_count_2)


if __name__ == "__main__":
    main()
