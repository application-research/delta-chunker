package cmd

import (
	"bufio"
	"bytes"
	"context"
	"delta-chunker/config"
	"delta-chunker/model"
	"delta-chunker/utils"
	"fmt"
	commcid "github.com/filecoin-project/go-fil-commcid"
	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

type CommpResult struct {
	commp     string
	pieceSize uint64
}

type Result struct {
	SourcePath      string                       `json:"source_path"`
	PayloadCid      string                       `json:"cid"`
	PieceCommitment PieceCommitment              `json:"piece_commitment"`
	Size            uint64                       `json:"size"`
	Miner           string                       `json:"miner"`
	CidMap          map[string]utils.CidMapValue `json:"cid_map"`
}

type PieceCommitment struct {
	PieceCID          string `json:"piece_cid"`
	PaddedPieceSize   uint64 `json:"padded_piece_size"`
	UnpaddedPieceSize uint64 `json:"unpadded_piece_size"`
}

func CarChunkRunnerCmd(config *config.DeltaChunkerConfig) []*cli.Command {
	var carCommands []*cli.Command
	carChunkerCmd := &cli.Command{
		Name:  "car-chunk-runner",
		Usage: "Generate car file(s) from a given file or directory",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "run-config",
				Usage: "path to run config file",
			},
			&cli.StringFlag{
				Name:  "output-dir",
				Usage: "path to output directory",
			},
		},
		Action: func(c *cli.Context) error {
			configFile := c.String("run-config")
			// open the db
			DB, err := model.OpenDatabase(config.Common.DBDSN)
			if err != nil {
				fmt.Println("Error opening database:", err)
				return err
			}

			data, err := os.ReadFile(configFile)
			if err != nil {
				fmt.Println("Error reading config file:", err)
				return err
			}

			// parse the YAML data into Config struct
			var cfg model.ChunkRunConfig
			err = yaml.Unmarshal(data, &cfg)
			if err != nil {
				fmt.Println("Error parsing YAML:", err)
				return err
			}

			//			DB.Create(&cfg)

			// access the individual chunk tasks
			for _, task := range cfg.ChunkTasks {
				fmt.Println("Task")
				fmt.Printf("Task name: %s\n", task.Name)
				fmt.Printf("Source: %s\n", task.Source)
				fmt.Printf("Output directory: %s\n", task.OutputDir)
				fmt.Printf("Split size: %d\n", task.SplitSize)
				fmt.Printf("Connection mode: %s\n", task.ConnectionMode)
				fmt.Printf("Miner: %s\n", task.Miner)
				fmt.Printf("Delta URL: %s\n", task.DeltaURL)
				fmt.Printf("Delta token: %s\n", task.DeltaApiKey)
				fmt.Printf("Delta wallet: %s\n", task.DeltaWallet)
				fmt.Printf("Delta metadata request: %s\n", task.DeltaMetadataReq)
				//go func() { // async!
				fmt.Println("Running task:", task.Name)
				// record on the database
				DB.Create(&task)

				// run it
				carChunkRunner(task, DB)

				// get all the results

				// create the delta metadata request

				// upload to delta

				// write the output to a file
				//}()

			}
			return nil
		},
	}
	carCommands = append(carCommands, carChunkerCmd)

	return carCommands
}

func carChunkRunner(chunkTask model.ChunkTask, DB *gorm.DB) error {
	ctx := context.Background()
	if _, err := os.Stat(chunkTask.OutputDir); os.IsNotExist(err) {
		return err
	}
	var input Input
	var outputs []Result
	if chunkTask.SplitSize == "" {
		chunkTask.SplitSize = "0"
	}

	fmt.Println("Split size: ", chunkTask.SplitSize)
	if chunkTask.SplitSize != "0" {
		splitSizeA, err := strconv.Atoi(chunkTask.SplitSize)
		if err != nil {
			return err
		}
		splitSize := int64(splitSizeA)

		err = filepath.Walk(chunkTask.Source, func(sourcePath string, info os.FileInfo, err error) error {
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
			chunkDir := filepath.Join(chunkTask.OutputDir, fileName)
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
				outPath := path.Join(chunkTask.OutputDir, outFilename)
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
					SourcePath: sourcePath,
				}

				if chunkTask.Miner != "" {
					output.Miner = chunkTask.Miner
				}

				if chunkTask.IncludeCommp {
					rawCommP, pieceSize, err := cp.Digest()
					if err != nil {
						return err
					}
					commCid, err := commcid.DataCommitmentV1ToCID(rawCommP)
					if err != nil {
						return err
					}
					err = os.Rename(outPath, path.Join(chunkTask.OutputDir, commCid.String()+".car"))
					if err != nil {
						return err
					}
					output.PieceCommitment.PieceCID = commCid.String()
					output.PieceCommitment.PaddedPieceSize = pieceSize
					output.Size = uint64(written)

					err = os.RemoveAll(outPath)
					if err != nil {
						return err
					}
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
		stat, err := os.Stat(chunkTask.Source)
		if err != nil {
			return err
		}
		if stat.IsDir() {
			err := filepath.Walk(chunkTask.Source, func(sourcePath string, info os.FileInfo, err error) error {
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
				outPath := path.Join(chunkTask.OutputDir, outFilename)
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
					SourcePath: sourcePath,
				}

				if chunkTask.Miner != "" {
					output.Miner = chunkTask.Miner
				}

				if chunkTask.IncludeCommp {
					rawCommP, pieceSize, err := cp.Digest()
					if err != nil {
						return err
					}
					commCid, err := commcid.DataCommitmentV1ToCID(rawCommP)
					if err != nil {
						return err
					}
					err = os.Rename(outPath, path.Join(chunkTask.OutputDir, commCid.String()+".car"))
					if err != nil {
						return err
					}
					output.PieceCommitment.PieceCID = commCid.String()
					output.PieceCommitment.PaddedPieceSize = pieceSize
					output.Size = uint64(info.Size())

					// remove the output path
					err = os.RemoveAll(outPath)
					if err != nil {
						return err
					}
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
				Path:  chunkTask.Source,
				Size:  stat.Size(),
				Start: 0,
				End:   stat.Size(),
			})
			outFilename := uuid.New().String() + ".car"
			outPath := path.Join(chunkTask.OutputDir, outFilename)
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

			if chunkTask.Miner != "" {
				output.Miner = chunkTask.Miner
			}

			if chunkTask.IncludeCommp {
				rawCommP, pieceSize, err := cp.Digest()
				if err != nil {
					return err
				}
				commCid, err := commcid.DataCommitmentV1ToCID(rawCommP)
				if err != nil {
					return err
				}
				err = os.Rename(outPath, path.Join(chunkTask.OutputDir, commCid.String()+".car"))
				if err != nil {
					return err
				}
				output.PieceCommitment.PieceCID = commCid.String()
				output.PieceCommitment.PaddedPieceSize = pieceSize
				output.Size = uint64(stat.Size())
				err = os.RemoveAll(outPath)
				if err != nil {
					return err
				}
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
}

func processOutput(deltaApiKey string, deltaEndpoint string, output Result, db *gorm.DB) error {

	return nil
}
