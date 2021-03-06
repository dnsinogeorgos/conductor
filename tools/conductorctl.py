#!/usr/bin/env python3
"""
usage: conductorctl [-h] [-f] {list,create,delete,refresh,help} [cast] [replica]

positional arguments:
  {list,create,delete,refresh,help}
                        Action to take.
  cast                  Name of cast.
  replica               Name of replica.

optional arguments:
  -h, --help            show this help message and exit
  -f, --force           Whether to force action by deleting child replicas or parent
  cast.
"""

import sys
from argparse import ArgumentParser
from http.client import responses as rsp
from prettytable import PrettyTable, MARKDOWN
import requests


# CONSTANTS
URL = "http://localhost:8080"
ACTIONS = ["list", "create", "delete", "refresh", "help"]


# ARGUMENT PARSING
PARSER = ArgumentParser(prog="conductorctl")
PARSER.add_argument("action", type=str, choices=ACTIONS, help="Action to take.")
PARSER.add_argument("cast", type=str, nargs="?", default=None, help="Name of cast.")
PARSER.add_argument(
    "replica", type=str, nargs="?", default=None, help="Name of replica."
)
PARSER.add_argument(
    "-f",
    "--force",
    action="store_true",
    help="Whether to force action by deleting child replicas or parent cast.",
)
ARGS = PARSER.parse_args()


# LOGIC
def print_table(cast_id=None):
    """Prints conductor service table."""
    table = PrettyTable(["Timestamp", "Cast", "Replica", "Port"])
    for row in populate_table(cast_id):
        table.add_row(row)
    table.set_style(MARKDOWN)
    print(table.get_string(sortby="Timestamp", reversesort=True))


def put(cast_id, replica_id, force):
    """Creates a new cast or replica."""
    if replica_id is None:
        create_cast(cast_id)
    else:
        if force is True:
            create_cast(cast_id)
        create_replica(cast_id, replica_id)


def update(cast_id, replica_id, force):
    """Recreates an already existing cast or replica."""
    if replica_id is None:
        if force is True:
            force_delete_cast(cast_id)
        else:
            delete_cast(cast_id)
        create_cast(cast_id)
    else:
        if force is True:
            force_delete_cast(cast_id)
            create_cast(cast_id)
            create_replica(cast_id, replica_id)
        else:
            delete_replica(cast_id, replica_id)
            create_replica(cast_id, replica_id)


def delete(cast_id, replica_id, force):
    """Deletes a cast or replica."""
    if replica_id is None:
        if force is True:
            force_delete_cast(cast_id)
        else:
            delete_cast(cast_id)
    else:
        delete_replica(cast_id, replica_id)


def force_delete_cast(cast_id):
    """Forcefully deletes a cast."""
    replicas = get_replicas(cast_id)
    if replicas:
        prompt("This WILL delete all running replicas on this cast. Are you sure?")
        for replica in replicas:
            delete_replica(cast_id, replica["id"])
    delete_cast(cast_id)


def populate_table(cast_id):
    """Populates the conductor table."""
    rows = []
    if cast_id is None:
        for cast in get_casts():
            replicas = get_replicas(cast["id"])
            for replica in replicas:
                rows.append(
                    [cast["timestamp"], cast["id"], replica["id"], replica["port"]]
                )
            if not replicas:
                rows.append([cast["timestamp"], cast["id"], "-", ""])
    else:
        cast = get_cast(cast_id)
        for replica in get_replicas(cast["id"]):
            rows.append([cast["timestamp"], cast["id"], replica["id"], replica["port"]])
    return rows


def print_response(replica):
    """Prints the conductor response."""
    print(
        "Received HTTP {} response '{}'.".format(
            replica.status_code, rsp[replica.status_code]
        )
    )


def prompt(msg):
    """Prompts for action confirmation."""
    response = input(msg + " (y/n): ")
    if response == "y":
        return
    sys.exit(1)


# API CALLS
def get_casts():
    """Retrieves the list of casts from the conductor service."""
    req = requests.get("{}/casts".format(URL))
    if req.status_code == 200:
        return req.json()
    print_response(req)
    sys.exit(1)


def get_cast(cast_id):
    """Retrieves a certain cast from the conductor service."""
    req = requests.get("{}/casts/{}".format(URL, cast_id))
    if req.status_code == 200:
        return req.json()
    print_response(req)
    sys.exit(1)


def get_replicas(cast_id):
    """Retrieves the list of replicas from the conductor service."""
    req = requests.get("{}/replicas/{}".format(URL, cast_id))
    if req.status_code == 200:
        return req.json()
    print_response(req)
    sys.exit(1)


def create_cast(cast_id):
    """Creates a cast at the conductor service."""
    req = requests.post("{}/casts/{}".format(URL, cast_id))
    if req.status_code == 201:
        print("Created cast {}.".format(cast_id))
    else:
        print_response(req)
        sys.exit(1)


def create_replica(cast_id, replica_id):
    """Creates a replica at the conductor service."""
    req = requests.post("{}/replicas/{}/{}".format(URL, cast_id, replica_id))
    if req.status_code == 201:
        print("Created replica {}/{}.".format(cast_id, replica_id))
    else:
        print_response(req)
        sys.exit(1)


def delete_cast(cast_id):
    """Deletes a cast at the conductor service."""
    req = requests.delete("{}/casts/{}".format(URL, cast_id))
    if req.status_code == 204:
        print("Deleted cast {}.".format(cast_id))
    else:
        print_response(req)
        sys.exit(1)


def delete_replica(cast_id, replica_id):
    """Deletes a replica at the conductor service."""
    req = requests.delete("{}/replicas/{}/{}".format(URL, cast_id, replica_id))
    if req.status_code == 204:
        print("Deleted replica {}/{}.".format(cast_id, replica_id))
    else:
        print_response(req)
        sys.exit(1)


# WIRING
if ARGS.action == "help":
    PARSER.print_help()

if ARGS.action == "list":
    if ARGS.replica or ARGS.force is True:
        PARSER.error("action {} accepts only --cast argument".format(ARGS.action))

if ARGS.action in ["add", "refresh", "rm"]:
    if ARGS.cast is None:
        PARSER.error("action {} requires --cast argument".format(ARGS.action))
    if ARGS.cast == "":
        PARSER.error("--cast argument cannot be empty string")
    if ARGS.replica == "":
        PARSER.error("--cast argument cannot be empty string")

if ARGS.replica:
    if ARGS.cast is None:
        PARSER.error("--replica argument requires --cast")

if ARGS.action == "list":
    print_table(ARGS.cast)

if ARGS.action == "create":
    put(ARGS.cast, ARGS.replica, ARGS.force)

if ARGS.action == "delete":
    delete(ARGS.cast, ARGS.replica, ARGS.force)

if ARGS.action == "refresh":
    update(ARGS.cast, ARGS.replica, ARGS.force)
