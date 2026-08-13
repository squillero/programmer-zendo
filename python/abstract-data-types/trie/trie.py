# Copyright © 2026 Giovanni Squillero <giovanni.squillero@polito.it>
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

from collections.abc import Generator, Iterator


class Node:
    _data: str
    _children: list["Node"]

    def __init__(self, data: str) -> None:
        self._data = data
        self._children = list()

    def __iter__(self):
        return iter(self._children)

    def __bool__(self) -> bool:
        return bool(self._data or self._children)

    @property
    def data(self) -> str:
        return self._data

    @property
    def is_leaf(self) -> bool:
        return not self._children

    def add_child(self, child: "Node") -> None:
        self._children.append(child)


class RadixTrie:
    _root: Node

    def __init__(self) -> None:
        self._root = Node("")

    def __iter__(self) -> Iterator[str]:
        # a generator is also and iterator...
        return RadixTrie._items("", self._root, "")

    def print_tree(self) -> None:
        # call recursive/static function
        RadixTrie._preorder_print_tree(self._root, ">")
        print()

    def items(self, *, separator: str = "") -> Generator[str]:
        # call recursive/static function
        return RadixTrie._items("", self._root, separator)

    def add(self, data: str) -> None:
        # call recursive/static function
        RadixTrie._add(self._root, data.upper() + "$")

    @staticmethod
    def _items(prefix: str, node: Node, separator: str) -> Generator[str]:
        if node.is_leaf:
            yield f"{prefix}{separator}{node._data[:-1]}"
        else:
            for child in node:
                yield from RadixTrie._items(prefix + node.data, child, separator)

    @staticmethod
    def _preorder_print_tree(node: Node, prefix: str) -> None:
        print(f"{prefix} {node.data!r}")
        for n in node:
            RadixTrie._preorder_print_tree(n, prefix + ">")

    @staticmethod
    def _add(node: Node, data: str) -> None:
        prefix = RadixTrie.common_prefix(node._data, data)
        if not node:
            node._data = data  # trie was empty, terrible!
        elif len(prefix) < len(node._data):
            old_node = Node(node._data[len(prefix) :])
            old_node._children = node._children
            new_node = Node(data[len(prefix) :])
            node._data = prefix
            node._children = [new_node, old_node]
        elif prefix == node._data and prefix != data:
            new_data = data[len(prefix) :]
            for c in node:
                if RadixTrie.common_prefix(c.data, new_data):
                    RadixTrie._add(c, new_data)
                    break
            else:
                node.add_child(Node(new_data))

    @staticmethod
    def common_prefix(s1: str, s2: str) -> str:
        for i, (c1, c2) in enumerate(zip(s1 + "$", s2 + "$")):
            if c1 != c2:
                break
        return s1[:i]
