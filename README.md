**Note: This repository is no longer maintained. For our new tool in generating cars, go to [delta-car-gen](https://github.com/alvin-reyes/delta-car-gen])**

# Δ Delta Chunker

[![CodeQL](https://github.com/application-research/delta-chunker/actions/workflows/codeql.yml/badge.svg)](https://github.com/application-research/delta-chunker/actions/workflows/codeql.yml)

*Delta chunker is currently under heavy development*

This repo contains the code for the delta-chunker binary. This binary is used to chunk files into smaller pieces (CAR files) for use with Delta deal making engine.

![image](https://user-images.githubusercontent.com/4479171/232383639-b52b36ce-9d13-4f7c-be80-bcd887e62891.png)


## Features
- Tracks all the files in a directory and chunks them into CAR files
- All CAR files can be streamed directly to a live delta instance
- Supports `e2e` and `import` deals
- Supports `split-size` to chunk files into smaller pieces
- Supports `delta-url` to stream CAR files to a live delta instance
- Supports `delta-api-key` to authenticate with a live delta instance
- Supports `delta-wallet` to specify the wallet to use for the deal. Pass the hexed wallet address from boostd/lotus export.
- Supports `type` to specify the type of deal to make
- For `e2e` type, it'll prepare the car and send the file over to the delta instance.
- For `import` type, it'll prepare the deal. You'll need to manually send the car file to the miner. *This is something we will automate soon with delta importer*
- API endpoint to allow external clients to trigger the chunker to run with a give large blob [WIP]

## Build 
```
make all
./dc car-chunk-runner --help
```

## Run chunker task
### Prepare the chunk task file
This will be the basis of the car-chunk-runner for making storage deals.
```
label: "sample-run"
chunk-tasks:
  - name: "chunk-task1"
    source: "./example/source"
    output-dir: "./example/output_dir"
    split-size: "0"
    connection-mode: "e2e"
    include-commp: true
    auto-retry: true
    miner: "f0137168"
    delta-url: "https://delta.estuary.tech"
    delta-api-key: "[API_KEY]"
    delta-wallet: "f1z5ykhx7qi2jeukwlve3mu2zbcuzrtewhc3iajmi" # hexed wallet address from boostd / lotus export
```
### Run the chunker
```
./dc car-chunk-runner --run-config=<run-config-file>
```

This will go thru each of the chunk-split tasks and run the delta chunker and make deals with the delta instance.

## Author
Protocol Labs Outercore Engineering.
