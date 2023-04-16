# Δ Delta Chunker

This repo contains the code for the delta-chunker binary. This binary is used to chunk files into smaller pieces (CAR files) for use with Delta deal making engine.

## Features
- Tracks all the files in a directory and chunks them into CAR files
- All CAR files can be streamed directly to a live delta instance
- Supports `e2e` and `import` deals
- Supports `split-size` to chunk files into smaller pieces
- Supports `delta-url` to stream CAR files to a live delta instance
- Supports `delta-token` to authenticate with a live delta instance
- Supports `delta-wallet` to specify the wallet to use for the deal
- Supports `type` to specify the type of deal to make

## Build 
```
make all
```

## Run chunker basic mode
```
./dc car --source=<source-dir> \ 
--outpit-dir=<dest-dir> \ 
--split-size=<size-in-bytes>
--type=<e2e|import> \  
--miner=<miner-address> \
--delta-url=<delta-url> \ 
--delta-token=<delta-token> \ 
--delta-wallet=<delta-wallet> 
```

## Run chunker advance mode using a yml file.
```
./dc car-chunk-runner --run-config=<run-config-file>
```

## Author
Protocol Labs Outercore Engineering.