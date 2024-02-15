# openmcs

An open source server emulator for your favorite online medieval clicking simulator.

## Building

You'll need at least Go 1.22 to work with this project.

You can build an executable for the server simply by running `make` in the top-level directory. This will run unit tests
and build the code into a binary which you can then run.

## Running

> **Warning**
> Be aware that operating your own server may be against the terms of service of the official game!

This project implements the protocol for revision 317 of the game.

Before you can run the server, you will need to place the game cache files into the `data` directory. These are non-free
and therefore not provided by this project. There is one, main cache file and 5 index files that need to be placed in
said directory.

Runtime configuration is in the `config.yaml` file, and it's recommended to at least review the default settings.

Finally, to run the server execute the following command:

`$ ./bin/openmcs`

This will start a server that listens on the port configured in `config.yaml`.

To kickstart the server, you can then run `make seed-sqlite3` to initialize the database with some basic data, including
a few player accounts you can use to log in right off the bat. It's recommended to restart the server after adding
seed data.

To connect to the server, you'll need a client with the same game revision.

## Content

Since this project is intended to be a framework rather than a complete, out-of-the-box product, there is limited
content provided in this repository. You're free to add your own content as you see fit, and to that extent several
tools are included to help accomplish that.

### Items

Items from the game cache can be referenced in the game engine by using their ID numbers. Each item must have a unique
ID, and may have optional attributes that influence how it's used in-game.

Items can be made known to the game by inserting records into the `ITEM_ATTRIBUTES` database table, and populating
relevant columns depending on the type of item (weapon, editable, etc.). Further, if you want to use the item in Lua
scripts, it's recommended to include the item's ID number in a constants file instead of using the raw number in code.

This project comes with an `itemgen` binary, which will generate Lua and (eventually) SQL scripts for items and their 
attributes. You'll need to provide an item definition file in JSON format (several are available on GitHub), that 
adheres to the following format:

```json
{
  "0": {
    "id": 0,
    "name": "My Item",
    "duplicate": false
  },
  "1": { ... },
  "2": { ... }
}
```

The `duplicate` key can be used to flag items that should be ignored and not processed. This is useful for skipping
items from the game cache that are useless or dupes of others.

The format to run the binary is then as follows:

`$ ./itemgen -item-file items-complete.json -script-output-dir ./scripts/generated -max-item-id 7619`

This will write a `scripts/generated/items.lua` file containing Lua constants for each item parsed from the definition
file. The `-max-item-id` can be used to limit how many items are loaded, especially if you are using a definition file
from later game revisions.

### Spells

In-game spells and magic are handled by customizable Lua scripts. This project provides some example spells from the
available spell books in the supported game revision, but you're free to add your own.

Each spell has a unique ID as determined by the spell book interface. Different spell books use different parent
interfaces, so it's easy to distinguish which spell was cast from which spell book. Spells are located in the 
`scripts/spells` directory, and new spells can be added at runtime.

## Monitoring

The server is instrumented with Prometheus metrics, available at http://localhost:2112/metrics. In addition to standard
Go metrics, the server exposes several of its own metrics to help provide insights into the various processes which
could impact performance.

* `game_state_update_duration_bucket`: a histogram describing how long game state updates take to complete
* `users_online_total`: a gauge for the current, active player count

This project comes with a Docker Compose stack consisting of a Prometheus and Grafana instance with prebuilt dashboards.

You can start the stack by running `make start-monitoring`, and bring it down by running `make stop-monitoring`. 

Once up, you can access the components at the following locations:

* Grafana: http://localhost:3000
* Prometheus: http://localhost:9090

The default Grafana login is `admin`/`admin`. Prometheus will store its metrics data under the `data/prometheus` 
directory, so it's safe to stop and start the stack as necessary without losing data.
