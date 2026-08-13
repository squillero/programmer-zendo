#   *        Giovanni Squillero's GP Toolbox
#  / \       ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
# 2   +      A no-nonsense GP in pure Python
#    / \     Coded @ GECCO 2024
#  10   11   (Melburne, Australia)
#
# Copyright © 2024 Giovanni Squillero / Politecnico di Torino
# https://github.com/squillero/programmer-zendo
# Free under certain conditions — see the license for details.

try:
    from icecream import install

    install()
except ImportError:
    pass

from .draw import *
from .gp_common import *
from .gp_dag import *
from .gp_tree import *
from .node import *
from .random import gxgp_random
from .utils import *
