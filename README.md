# Reference architecture using SingleStore for realtime esports analytics

**Attention**: The code in this repository is intended for experimental use only and is not fully tested, documented, or supported by SingleStore. Visit the [SingleStore Forums](https://www.singlestore.com/forum/) to ask questions about this repository.

The goal of this reference architecture is to build an esports dashboard displaying low-latency analytics for commentators.

## Background

Extremely experienced commentators can reflect on famous events and games, but often finding complex patterns when multiple events are happening at the same time is difficult.

In order to display high quality analytics, we need real data. The data set includes 10,000 StarCraft 2 replays where the players are top ranked professional players. If you follow SC2, you may recognize many of the names.

StarCraft 2 (SC2) is a realtime strategy (RTS) video game developed and published by Blizzard Entertainment with esports matches beginning in 2010. SC2 has a large community and several professional leagues.  There are three races with unique units and strategies: terran, zerg, protoss. The goal of each match is to destroy all of the opponents buildings or create a situation where there is no chance of winning and the opponent forfeits.

# Run the demo yourself!

This project is easy to run and I encourage you to play with it a bit. Once you get it setup you can easily add additional commands or play with different ways to optimize the queries or schema. Let's get started!

## Dependencies

This git repo includes a [VS Code development container][vscode-devcontainer] configuration. This means that if you open this repo using VS Code the entire development environment can be automatically setup for you.

If you want to run the demo without using a dev container you will need to setup golang, nodejs and, yarn.

## Setting up SingleStore

Next, we need a SingleStore cluster to connect to. This demo is quite computationally intensive, and in general we only run it using a S-10 on the managed service. You can set up a S-10 like so:

1. [Sign up][try-free] for $500 in free managed service credits.
2. Create a S-10 sized cluster in [the portal][portal]
3. Copy `config.example.toml` to a new file and edit the `[singlestore]` section to match:

```toml
[database]
    host = "THE CONNECTION ENDPOINT"
    port = "3306"
    username = "admin"
    password = "THE ADMIN PASSWORD"
    database = "sc2"
```

## Initialize the schema

Using the SQL editor (in the [portal][portal]) or via the mysql CLI run the contents of [schema.sql](schema.sql) and [pipelines.sql](pipelines.sql) against the database. Here is how I would do this using the mysql CLI:

```bash
SINGLESTORE_HOST="XXXXXXXXXXXXXXXXXXXXXX"
SINGLESTORE_PASSWORD="XXXXXXXXXXXXXX"
mysql -u admin -h "${SINGLESTORE_HOST}" -p"${SINGLESTORE_PASSWORD}" <schema.sql <pipelines.sql
```

## Run the Demo

In vscode you can run the demo using tasks. Press `CTRL-SHIFT-P` and then search for "Run Task". Select the task named "start all" to start the web interface and API service.

Alternatively you can start services in two terminals:

In the first terminal run:

```bash
cd web
yarn dev
```

In the second terminal run:

```bash
cd src
go build -o bin/player/__bin bin/player/main.go
bin/player/__bin --confing ../config.example.toml --config ../config.toml
```

Then just open up http://localhost:3001 in your browser.

## Navigating the demo UI

On the home screen, each card represents a unique StartCraft 2 replay from professional players.

1.  Select the first game on the homescreen: Zest (Protoss) vs Reynor (Zerg)
2.  Hit the play button in the top-right hand corner.
    The first timeline at the top represents a live game in realtime. We are monitoring for new events.
    The timelines below this one are similar games that match the realtime match's build composition.

    To start, almost every Protoss player will create the same initial
    units. The first fork begins at about **2m30s**.

    In this match, we can see two possible forks. The Protoss player may aim for building stalkers, going for a mothership, or getting an early stargate.

    This game is about managing resources. A building unit consumes resources of crystals and gas.
    How those resources are allocated is key to late game strategies.

    In this case, the player has chosen a stargate. As soon as the stargate has been chosen, this eliminated the other strategies. Notice how the timelines have changed.

As we can see, a commentator can look at this dashboard with specific context and provide
informative commentatory for things the audience can look for.

With additional features, we could even have a fully automated commentator.

# Details

## Analytics

We have a “buildcomp“ table, short for build composition. For each game, for each player, in one game loop (1/16th of second). We are looking for the kind and number of units, e.g. kind = zergling, num = 1.

Buildcomp reflects the delta of the kinds of events that are occurring.

The compvecs table is a prebuilt result of all of the games. This produces a SingleStore floating point vector that reflects the composition at that point.

## Searching

Per event we compute <60ms (assuming correctly sized SingleStore cluster)
k-nearest neighbor search of 20
million build composition vectors.
Each one of those vectors, is a small window of one of the 10,000 games
we loaded into our database.

We search for the 5 most similar games.

This is done by dynamically producing the buildcomp vector of the realtime game,
then we look for that same vector in the 20 million vector dataset.

We aggregate down to the 5 most similar games.
The second timeline is the most similar game.

Right now, we don't treat units with different weights, so there are definitely pieces that could be improved, but the bones are there.

We are running potentially hundreds of simularity searches within milliseconds. Because of the power of singlestore, all of this is done in realtime.

<!-- link index -->

[s2]: https://www.singlestore.com
[vscode-devcontainer]: https://code.visualstudio.com/docs/remote/containers
[try-free]: https://www.singlestore.com/try-free/
[ciab]: https://github.com/memsql/deployment-docker
[portal]: https://portal.singlestore.com/
[s2-forums]: https://www.singlestore.com/forum/
[gh-issue]: issues
