# Author: Giovanni Squillero <giovanni.squillero@polito.it>
# Copyright © 2026 Giovanni Squillero / Politecnico di Torino
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

from itertools import product
from string import ascii_uppercase

import pytest

from trie.trie import Node, RadixTrie


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


def make_trie(*words):
    trie = RadixTrie()
    for w in words:
        trie.add(w)
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


# common_prefix


@pytest.mark.parametrize(
    "s1, s2, expected",
    [
        ("", "", ""),
        ("", "A", ""),
        ("A", "", ""),
        ("A", "A", "A"),
        ("A", "B", ""),
        ("ABC", "ABC", "ABC"),
        ("AB", "ABC", "AB"),
        ("ABC", "AB", "AB"),
        ("XYZ", "ABC", ""),
        ("A$", "A", "A"),
        ("A", "A$", "A"),
        ("A$", "A$", "A$"),
    ],
)
def test_common_prefix(s1, s2, expected):
    assert RadixTrie.common_prefix(s1, s2) == expected


# insertion and listing


def test_single_word():
    assert list(make_trie("word")) == ["WORD"]


def test_single_char():
    assert list(make_trie("a")) == ["A"]


@pytest.mark.parametrize("word", ["0", "42", "abc123", "Ünïcødé"])
def test_alnum_accepted(word):
    assert list(make_trie(word)) == [word.upper()]


@pytest.mark.parametrize("word", ["", " ", "$", "a$", "a b", "a-b", "cat\n"])
def test_non_alnum_rejected(word):
    trie = RadixTrie()
    with pytest.raises(ValueError, match="expected only letters and numbers"):
        trie.add(word)


def test_rejected_word_leaves_trie_unchanged():
    trie = make_trie("cat", "dog")
    assert sorted(trie) == ["CAT", "DOG"]


def test_duplicate():
    assert list(make_trie("cat", "cat")) == ["CAT"]


def test_duplicate_deep():
    assert sorted(make_trie("cat", "dog", "cat", "dog")) == ["CAT", "DOG"]


def test_duplicate_case_insensitive():
    assert list(make_trie("cat", "CAT", "CaT")) == ["CAT"]


def test_uppercase_unicode():
    assert list(make_trie("straße")) == ["STRASSE"]


@pytest.mark.parametrize(
    "words",
    [
        ("ca", "car", "card"),
        ("card", "car", "ca"),
        ("car", "ca", "card"),
    ],
)
def test_prefix_chain(words):
    assert sorted(make_trie(*words)) == ["CA", "CAR", "CARD"]


def test_split_at_root():
    trie = make_trie("cat", "dog")
    assert sorted(trie) == ["CAT", "DOG"]
    assert sorted(trie.items(separator=":")) == [":CAT", ":DOG"]


def test_split_mid_node():
    trie = make_trie("romane", "romulus")
    assert sorted(trie.items(separator=":")) == ["ROM:ANE", "ROM:ULUS"]


def test_long_word():
    word = "A" * 10_000
    assert list(make_trie(word, word)) == [word]


def test_long_words_sharing_prefix():
    w1 = "A" * 5_000 + "B"
    w2 = "A" * 5_000 + "C"
    assert sorted(make_trie(w1, w2)) == [w1, w2]


def test_many_words():
    words = ["".join(p) for p in product(ascii_uppercase, repeat=2)]
    assert sorted(make_trie(*words)) == words


def test_iter_equals_items(sample_trie):
    assert list(sample_trie) == list(sample_trie.items())


# items(separator=...)


def test_default_separator(sample_trie):
    assert list(sample_trie.items()) == list(sample_trie.items(separator=""))


def test_multichar_separator():
    assert sorted(make_trie("roman", "romulus").items(separator="<->")) == [
        "ROM<->AN",
        "ROM<->ULUS",
    ]


def test_separator_single_word():
    assert list(make_trie("a").items(separator=":")) == [":A"]


# print_tree


def test_print_tree(capsys):
    make_trie("cat", "dog").print_tree()
    assert capsys.readouterr().out == "> ''\n>> 'DOG$'\n>> 'CAT$'\n\n"


def test_print_tree_empty(capsys):
    RadixTrie().print_tree()
    assert capsys.readouterr().out == "> ''\n\n"


# Node


def test_node_bool():
    assert not Node("")
    assert Node("A")
    parent = Node("")
    parent.add_child(Node("A"))
    assert parent


def test_node_children():
    node = Node("A")
    assert node.data == "A"
    assert node.is_leaf
    assert list(node) == []
    children = [Node("B"), Node("C")]
    for c in children:
        node.add_child(c)
    assert not node.is_leaf
    assert list(node) == children


# empty trie


def test_empty_trie():
    assert list(RadixTrie()) == []
