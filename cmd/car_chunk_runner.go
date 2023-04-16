package cmd

import (
	"bufio"
	"bytes"
	"context"
	"delta-chunker/utils"
	"fmt"
	commcid "github.com/filecoin-project/go-fil-commcid"
	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func CarChunkRunnerCmd() []*cli.Command {
	ctx := context.TODO()
	var carCommands []*cli.Command
	carChunkerCmd := &cli.Command{
		Name:  "car-chunk-runner",
		Usage: "Generate car file(s) from a given file or directory",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "run-config",
				Usage: "path to run config file",
			},
		},
		Action: func(c *cli.Context) error {
			sourceInput := c.String("source")
			splitSizeInput := c.String("split-size")
			outDir := c.String("output-dir")
			includeCommp := c.Bool("include-commp")
			minerAssignment := c.String("miner")

			if _, err := os.Stat(outDir); os.IsNotExist(err) {
				return err
			}
			var input Input
			var outputs []Result
			if splitSizeInput != "" {
				splitSizeA, err := strconv.Atoi(splitSizeInput)
				if err != nil {
					return err
				}
				splitSize := int64(splitSizeA)

				err = filepath.Walk(sourceInput, func(sourcePath string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						return nil
					}

					chunks := (info.Size() + splitSize - 1) / splitSize
					sourceFile, err := os.Open(sourcePath)
					if err != nil {
						return err
					}

					fileName := info.Name()
					chunkDir := filepath.Join(outDir, fileName)
					err = os.MkdirAll(chunkDir, 0755)
					if err != nil {
						return err
					}

					for i := int64(0); i < chunks; i++ {
						chunkFileName := fmt.Sprintf("%s_%04d", fileName, i)
						chunkFilePath := filepath.Join(chunkDir, chunkFileName)
						chunkFile, err := os.Create(chunkFilePath)
						if err != nil {
							return err
						}

						start := i * splitSize
						end := (i + 1) * splitSize
						if end > info.Size() {
							end = info.Size()
						}
						_, err = sourceFile.Seek(start, 0)
						if err != nil {
							return err
						}
						written, err := io.CopyN(chunkFile, sourceFile, end-start)
						if err != nil && err != io.EOF {
							return err
						}
						input = append(input, utils.Finfo{
							Path:  chunkFilePath,
							Size:  written,
							Start: 0,
							End:   written,
						})
						outFilename := uuid.New().String() + ".car"
						//outFilename := cid + ".car"
						outPath := path.Join(outDir, outFilename)
						carF, err := os.Create(outPath)
						if err != nil {
							return err
						}
						cp := new(commp.Calc)
						writer := bufio.NewWriterSize(io.MultiWriter(carF, cp), BufSize)
						_, cid, cidMap, err := utils.GenerateCar(ctx, input, "", "", writer)
						if err != nil {
							return err
						}
						err = writer.Flush()
						if err != nil {
							return err
						}
						output := Result{
							PayloadCid: cid,
							CidMap:     cidMap,
						}

						if minerAssignment != "" {
							output.Miner = minerAssignment
						}

						if includeCommp {
							rawCommP, pieceSize, err := cp.Digest()
							if err != nil {
								return err
							}
							commCid, err := commcid.DataCommitmentV1ToCID(rawCommP)
							if err != nil {
								return err
							}
							err = os.Rename(outPath, path.Join(outDir, commCid.String()+".car"))
							if err != nil {
								return err
							}
							output.PieceCommitment.PieceCID = commCid.String()
							output.PieceCommitment.PaddedPieceSize = pieceSize
							output.Size = uint64(written)
						}
						outputs = append(outputs, output)
					}
					return nil
				})
				var buffer bytes.Buffer
				err = utils.PrettyEncode(outputs, &buffer)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(buffer.String())
				if err != nil {
					panic(err)
				}
			} else {
				stat, err := os.Stat(sourceInput)
				if err != nil {
					return err
				}
				if stat.IsDir() {
					err := filepath.Walk(sourceInput, func(sourcePath string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if info.IsDir() {
							return nil
						}
						input = append(input, utils.Finfo{
							Path:  sourcePath,
							Size:  info.Size(),
							Start: 0,
							End:   info.Size(),
						})
						outFilename := uuid.New().String() + ".car"
						outPath := path.Join(outDir, outFilename)
						carF, err := os.Create(outPath)
						if err != nil {
							return err
						}
						cp := new(commp.Calc)
						writer := bufio.NewWriterSize(io.MultiWriter(carF, cp), BufSize)
						_, cid, cidMap, err := utils.GenerateCar(ctx, input, "", "", writer)
						if err != nil {
							return err
						}
						err = writer.Flush()
						if err != nil {
							return err
						}

						output := Result{
							PayloadCid: cid,
							CidMap:     cidMap,
						}

						if minerAssignment != "" {
							output.Miner = minerAssignment
						}

						if includeCommp {
							rawCommP, pieceSize, err := cp.Digest()
							if err != nil {
								return err
							}
							commCid, err := commcid.DataCommitmentV1ToCID(rawCommP)
							if err != nil {
								return err
							}
							err = os.Rename(outPath, path.Join(outDir, commCid.String()+".car"))
							if err != nil {
								return err
							}
							output.PieceCommitment.PieceCID = commCid.String()
							output.PieceCommitment.PaddedPieceSize = pieceSize
							output.Size = uint64(info.Size())
						}
						outputs = append(outputs, output)
						return nil
					})

					var buffer bytes.Buffer
					err = utils.PrettyEncode(outputs, &buffer)
					if err != nil {
						fmt.Println(err)
					}
					fmt.Println(buffer.String())
					return nil
					if err != nil {
						return err
					}
				} else {
					input = append(input, utils.Finfo{
						Path:  sourceInput,
						Size:  stat.Size(),
						Start: 0,
						End:   stat.Size(),
					})
					outFilename := uuid.New().String() + ".car"
					outPath := path.Join(outDir, outFilename)
					carF, err := os.Create(outPath)
					if err != nil {
						return err
					}
					cp := new(commp.Calc)
					writer := bufio.NewWriterSize(io.MultiWriter(carF, cp), BufSize)
					_, cid, cidMap, err := utils.GenerateCar(ctx, input, "", "", writer)
					if err != nil {
						return err
					}
					err = writer.Flush()
					if err != nil {
						return err
					}
					output := Result{
						PayloadCid: cid,
						CidMap:     cidMap,
					}

					if minerAssignment != "" {
						output.Miner = minerAssignment
					}

					if includeCommp {
						rawCommP, pieceSize, err := cp.Digest()
						if err != nil {
							return err
						}
						commCid, err := commcid.DataCommitmentV1ToCID(rawCommP)
						if err != nil {
							return err
						}
						err = os.Rename(outPath, path.Join(outDir, commCid.String()+".car"))
						if err != nil {
							return err
						}
						output.PieceCommitment.PieceCID = commCid.String()
						output.PieceCommitment.PaddedPieceSize = pieceSize
						output.Size = uint64(stat.Size())
					}
					if err != nil {
						return err
					}
					var buffer bytes.Buffer
					err = utils.PrettyEncode(output, &buffer)
					if err != nil {
						fmt.Println(err)
					}
					fmt.Println(buffer.String())
				}
				return nil
			}

			return nil
		},
	}
	carCommands = append(carCommands, carChunkerCmd)

	return carCommands
}
