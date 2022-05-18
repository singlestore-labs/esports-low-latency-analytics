This repo contains code needed to process any sc2 replays you are interested in. Follow these steps to run the processor.

1. Download sc2 replays into a directory (./data/replays is suggested). The replays must have the extension `.SC2Replay`. Nested directories are supported.
2. Make sure you have a valid configuration file (see main readme)
3. Build the replay processor binary
    ```bash
    cd src
    go build -o bin/processor/__bin bin/processor/main.go
    ```
4. Run the replay processor
   ```bash
   src/bin/processor/__bin --config config.example.toml --config config.toml
   ```

This process can take quite some time for large numbers of replays. To construct the dataset documented in the [readme](README.md) took my computer a couple hours. If you want to scale this up, I suggest modifying the processor code to skip post-processing, scaling out the processor over many machines with different sets of replays, and then running post-processing once at the end. Post-processing is already designed to run in parallel inside of the SingleStore cluster, however splitting up the `prepareCompvecs` function into many parallel executions may also provide some performance boost. This is left as an exercise for the reader.