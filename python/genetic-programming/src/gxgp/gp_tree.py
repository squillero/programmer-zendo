#   *        Giovanni Squillero's GP Toolbox
#  / \       ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# 2   +      A no-nonsense GP in pure Python
#    / \     Coded @ GECCO 2024
#  10   11   (Melburne, Australia)
#
# Copyright © 2024 Giovanni Squillero / Politecnico di Torino
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

import random
from typing import Collection

from .node import Node
from .utils import arity

__all__ = ["TreeGP"]


class TreeGP:
    def __init__(
        self,
        operators: Collection,
        variables: int | Collection,
        constants: int | Collection,
        *,
        seed=42,
    ):
        raise NotImplementedError
