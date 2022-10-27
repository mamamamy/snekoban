# Snekoban Game

import json
import typing

import subprocess

# NO ADDITIONAL IMPORTS! X


direction_vector = {
    "up": (-1, 0),
    "down": (+1, 0),
    "left": (0, -1),
    "right": (0, +1),
}


def call_go(opt, data):
    exepath = "./snekoban"
    args = exepath + " " + opt
    p = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, universal_newlines=True, shell=True)

    data = json.dumps(data)
    p.stdin.write(data)
    p.stdin.flush()
    p.stdin.close()

    output = p.stdout.read()
    showSet = {
        # "-new_game",
        # "-victory_check",
        # "-step_game",
        # "-dump_game",
        "-solve_puzzle",
    }
    if opt in showSet:
        print(opt, output)
    data = json.loads(output)
    if data["errCode"] != 0:
        raise RuntimeError( data["errMsg"])
    return data["data"]

def new_game(level_description):
    """
    Given a description of a game state, create and return a game
    representation of your choice.

    The given description is a list of lists of lists of strs, representing the
    locations of the objects on the board (as described in the lab writeup).

    For example, a valid level_description is:

    [
        [[], ['wall'], ['computer']],
        [['target', 'player'], ['computer'], ['target']],
    ]

    The exact choice of representation is up to you; but note that what you
    return will be used as input to the other functions.
    """
    return call_go("-new_game", level_description)


def victory_check(game):
    """
    Given a game representation (of the form returned from new_game), return
    a Boolean: True if the given game satisfies the victory condition, and
    False otherwise.
    """
    return call_go("-victory_check", game)


def step_game(game, direction):
    """
    Given a game representation (of the form returned from new_game), return a
    new game representation (of that same form), representing the updated game
    after running one step of the game.  The user's input is given by
    direction, which is one of the following: {'up', 'down', 'left', 'right'}.

    This function should not mutate its input.
    """
    return call_go("-step_game", {"game": game, "direction": direction})


def dump_game(game):
    """
    Given a game representation (of the form returned from new_game), convert
    it back into a level description that would be a suitable input to new_game
    (a list of lists of lists of strings).

    This function is used by the GUI and the tests to see what your game
    implementation has done, and it can also serve as a rudimentary way to
    print out the current state of your game for testing and debugging on your
    own.
    """
    return call_go("-dump_game", game)


def solve_puzzle(game):
    """
    Given a game representation (of the form returned from new game), find a
    solution.

    Return a list of strings representing the shortest sequence of moves ("up",
    "down", "left", and "right") needed to reach the victory condition.

    If the given level cannot be solved, return None.
    """
    return call_go("-solve_puzzle", game)


if __name__ == "__main__":
    pass
