#   *        Giovanni Squillero's GP Toolbox
#  / \       ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# 2   +      A no-nonsense GP in pure Python
#    / \     Coded @ GECCO 2024
#  10   11   (Melburne, Australia)
#
# Copyright © 2024 Giovanni Squillero / Politecnico di Torino
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

import inspect
from collections.abc import Callable

__all__ = ["arity"]


def arity(f: Callable) -> int:
    """Return the number of expected parameter of the function"""
    return len(inspect.getfullargspec(f).args)
