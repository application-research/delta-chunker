# Δ Delta Chunker

**Delta chunker is currently under heavy development**

This repo contains the code for the delta-chunker binary. This binary is used to chunk files into smaller pieces (CAR files) for use with Delta deal making engine.

![image](https://user-images.githubusercontent.com/4479171/232341521-702b0530-5462-46e7-8393-70225369b368.png)


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

## Build 
```
make all
./dc car --help
```

## Run chunker basic mode
```
./dc car --source=<source-dir> \ 
--outpit-dir=<dest-dir> \ 
--split-size=<size-in-bytes>
--type=<e2e|import> \  
--miner=<miner-address> \
--delta-url=<delta-url> \ 
--delta-api-key=<delta-token> \ 
--delta-wallet=<delta-wallet> 
```

## Run chunker advance mode using a yml file.
### Prepare the configuration file
```
label: "sample-run"
chunk-tasks:
  - name: "chunk-task1"
    source: "myfiles_for_client_1"
    output-dir: "myfiles_for_client_2"
    split-size: 10000000
    connection_mode: "e2e"
    miner: "f1q2w3e4r5t6y7u8i9o0p1a2s3d4f5g6h7j8k9l0"
    delta-url: "https://node.delta.store"
    delta-token: "delta"
    delta-wallet: "delta" # hexed wallet address from boostd / lotus export
    delta-metadata-request: "{\"auto_retry\":true}" # no need to specify miner here, it will be taken from miner field
  - name: "chunk-task2"
    source: "myfiles_for_client_2"
    output-dir: "myfiles_for_client_2"
    split-size: 10000000
    connection_mode: "e2e"
    miner: "f1q2w3e4r5t6y7u8i9o0p1a2s3d4f5g6h7j8k9l0"
    delta-url: "http://localhost:1313"
    delta-token: "delta"
    delta-wallet: "delta" # hexed wallet address from boostd / lotus export
    delta-metadata-request: "{\"auto_retry\":true}" # no need to specify miner here, it will be taken from miner field

```
### Run the chunker
```
./dc car-chunk-runner --run-config=<run-config-file>
```

This will go thru each of the chunk-split tasks and run the delta chunker and make deals with the delta instance.

## Author
Protocol Labs Outercore Engineering.
