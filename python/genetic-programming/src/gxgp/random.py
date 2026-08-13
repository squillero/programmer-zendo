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

assert "gxgp_random" not in globals(), "Paranoia check: gxgp_random already initialized"
gxgp_random = random.Random(42)
