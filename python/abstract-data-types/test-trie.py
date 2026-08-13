# Copyright © 2026 Giovanni Squillero <giovanni.squillero@polito.it>
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

import pytest
from trie.trie import RadixTrie


@pytest.fixture
def sample_trie():
    trie = RadixTrie()

    trie.add("romane")
    trie.add("romanus")
    trie.add("romulus")
    trie.add("rubens")
    trie.add("ruber")
    trie.add("rubicon")
    trie.add("rubicundus")
    trie.add("roman")

    return trie


def test_items(sample_trie):
    assert list(sample_trie.items(separator=":")) == [
        "RUBIC:UNDUS",
        "RUBIC:ON",
        "RUBE:R",
        "RUBE:NS",
        "ROM:ULUS",
        "ROMAN:US",
        "ROMAN:E",
        "ROMAN:",
    ]


def foo():
    print("\nITEM LIST (unsorted):")

    print("\nTREE:")
    trie.print_tree()


if __name__ == "__main__":
    main()
