#   *        Giovanni Squillero's GP Toolbox
#  / \       ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# 2   +      A no-nonsense GP in pure Python
#    / \     Coded @ GECCO 2024
#  10   11   (Melburne, Australia)
#
# Copyright © 2024 Giovanni Squillero / Politecnico di Torino
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

import operator
import math
from gxgp import Node


def test_t1():
    leaf_num = Node(1)
    leaf_var = Node("x")
    node_op = Node(operator.add, [leaf_num, leaf_var])
    node_op = Node(math.sin, [node_op])
    f = Node(operator.mul, [node_op, leaf_var])
    g = Node(lambda c, x, y: x if c > 0 else y, [f, f, node_op])

    assert f.short_name == "mul"
    assert f.long_name == "mul(sin(add(1, x)), x)"
    assert f.short_name == str(f)
