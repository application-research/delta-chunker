package cmd

import (
	"bufio"
	"bytes"
	"context"
	"delta-chunker/config"
	"delta-chunker/model"
	"delta-chunker/utils"
	"encoding/json"
	"fmt"
	commcid "github.com/filecoin-project/go-fil-commcid"
	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
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

type WalletRequest struct {
	Id         uint64 `json:"id,omitempty"`
	Address    string `json:"address,omitempty"`
	Uuid       string `json:"uuid,omitempty"`
	KeyType    string `json:"key_type,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
}

type PieceCommitmentRequest struct {
	Piece             string `json:"piece_cid,omitempty"`
	PaddedPieceSize   uint64 `json:"padded_piece_size,omitempty"`
	UnPaddedPieceSize uint64 `json:"unpadded_piece_size,omitempty"`
}

type DealRequest struct {
	Cid                    string                 `json:"cid,omitempty"`
	Source                 string                 `json:"source,omitempty"`
	Miner                  string                 `json:"miner,omitempty"`
	Duration               int64                  `json:"duration,omitempty"`
	DurationInDays         int64                  `json:"duration_in_days,omitempty"`
	Wallet                 WalletRequest          `json:"wallet,omitempty"`
	PieceCommitment        PieceCommitmentRequest `json:"piece_commitment,omitempty"`
	ConnectionMode         string                 `json:"connection_mode,omitempty"`
	Size                   int64                  `json:"size,omitempty"`
	StartEpoch             int64                  `json:"start_epoch,omitempty"`
	StartEpochInDays       int64                  `json:"start_epoch_in_days,omitempty"`
	Replication            int                    `json:"replication,omitempty"`
	RemoveUnsealedCopy     bool                   `json:"remove_unsealed_copy"`
	SkipIPNIAnnounce       bool                   `json:"skip_ipni_announce"`
	AutoRetry              bool                   `json:"auto_retry"`
	Label                  string                 `json:"label,omitempty"`
	DealVerifyState        string                 `json:"deal_verify_state,omitempty"`
	UnverifiedDealMaxPrice string                 `json:"unverified_deal_max_price,omitempty"`
}

type DealE2EUploadResponse struct {
	Status          string `json:"status"`
	Message         string `json:"message"`
	ContentID       int    `json:"content_id"`
	DealRequestMeta struct {
		Cid    string `json:"cid"`
		Miner  string `json:"miner"`
		Wallet struct {
		} `json:"wallet"`
		PieceCommitment struct {
		} `json:"piece_commitment"`
		ConnectionMode     string `json:"connection_mode"`
		Replication        int    `json:"replication"`
		RemoveUnsealedCopy bool   `json:"remove_unsealed_copy"`
		SkipIpniAnnounce   bool   `json:"skip_ipni_announce"`
		AutoRetry          bool   `json:"auto_retry"`
	} `json:"deal_request_meta"`
	DealProposalParameterRequestMeta struct {
		ID                 int    `json:"ID"`
		Content            int    `json:"content"`
		Label              string `json:"label"`
		Duration           int    `json:"duration"`
		RemoveUnsealedCopy bool   `json:"remove_unsealed_copy"`
		SkipIpniAnnounce   bool   `json:"skip_ipni_announce"`
		VerifiedDeal       bool   `json:"verified_deal"`
		CreatedAt          string `json:"created_at"`
		UpdatedAt          string `json:"updated_at"`
	} `json:"deal_proposal_parameter_request_meta"`
	ReplicatedContents []struct {
		Status          string `json:"status"`
		Message         string `json:"message"`
		ContentID       int    `json:"content_id"`
		DealRequestMeta struct {
			Cid    string `json:"cid"`
			Miner  string `json:"miner"`
			Wallet struct {
			} `json:"wallet"`
			PieceCommitment struct {
			} `json:"piece_commitment"`
			ConnectionMode     string `json:"connection_mode"`
			Replication        int    `json:"replication"`
			RemoveUnsealedCopy bool   `json:"remove_unsealed_copy"`
			SkipIpniAnnounce   bool   `json:"skip_ipni_announce"`
			AutoRetry          bool   `json:"auto_retry"`
		} `json:"deal_request_meta"`
		DealProposalParameterRequestMeta struct {
			ID                 int    `json:"ID"`
			Content            int    `json:"content"`
			Label              string `json:"label"`
			Duration           int    `json:"duration"`
			RemoveUnsealedCopy bool   `json:"remove_unsealed_copy"`
			SkipIpniAnnounce   bool   `json:"skip_ipni_announce"`
			VerifiedDeal       bool   `json:"verified_deal"`
			CreatedAt          string `json:"created_at"`
			UpdatedAt          string `json:"updated_at"`
		} `json:"deal_proposal_parameter_request_meta"`
	} `json:"replicated_contents"`
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

					//err = os.RemoveAll(outPath)
					//if err != nil {
					//	return err
					//}
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
		for _, output := range outputs {
			err = processOutput(chunkTask, output, DB)
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
					//err = os.RemoveAll(outPath)
					//if err != nil {
					//	return err
					//}
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
			for _, output := range outputs {
				err = processOutput(chunkTask, output, DB)
			}
			return nil

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
				//err = os.RemoveAll(outPath)
				//if err != nil {
				//	return err
				//}
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
			for _, output := range outputs {
				err = processOutput(chunkTask, output, DB)
			}
		}

		return nil
	}
	return nil
}

func processOutput(chunkTask model.ChunkTask, output Result, db *gorm.DB) error {
	if chunkTask.DeltaApiKey != "" && chunkTask.DeltaURL != "" {
		dealRequest := DealRequest{
			Cid:                output.PayloadCid,
			Miner:              output.Miner,
			Wallet:             WalletRequest{Address: chunkTask.DeltaWallet},
			PieceCommitment:    PieceCommitmentRequest{Piece: output.PieceCommitment.PieceCID, PaddedPieceSize: output.PieceCommitment.PaddedPieceSize},
			ConnectionMode:     chunkTask.ConnectionMode,
			RemoveUnsealedCopy: true,
			SkipIPNIAnnounce:   false,
			AutoRetry:          chunkTask.AutoRetry,
			DealVerifyState:    "verified",
		}

		if chunkTask.ConnectionMode == "e2e" {
			// load the source and make  http request to delta
			file := chunkTask.OutputDir + "/" + output.PieceCommitment.PieceCID + ".car"
			// open the file
			f, err := os.Open(file)
			if err != nil {

			}
			dealRequest.Size = int64(output.Size)

			// create a new file upload http request with optional extra params
			payload := &bytes.Buffer{}
			writer := multipart.NewWriter(payload)
			partFile, err := writer.CreateFormFile("data", output.PieceCommitment.PieceCID+".car")
			if err != nil {
				fmt.Println("CreateFormFile error: ", err)
				return nil
			}
			_, err = io.Copy(partFile, f)
			if err != nil {
				fmt.Println("Copy error: ", err)
				return nil
			}
			if partFile, err = writer.CreateFormField("metadata"); err != nil {
				fmt.Println("CreateFormField error: ", err)
				return nil
			}

			partMetadata := fmt.Sprintf(`{"auto_retry":true,"miner":"%s"}`, dealRequest.Miner)

			if _, err = partFile.Write([]byte(partMetadata)); err != nil {
				fmt.Println("Write error: ", err)
				return nil
			}
			if err = writer.Close(); err != nil {
				fmt.Println("Close error: ", err)
				return nil
			}

			req, err := http.NewRequest("POST",
				chunkTask.DeltaURL+"/api/v1/deal/end-to-end",
				payload)

			if err != nil {
				fmt.Println(err)
				return nil
			}

			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer "+chunkTask.DeltaApiKey)
			client := &http.Client{}
			var res *http.Response
			res, err = client.Do(req)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			if res.StatusCode == 200 {
				var dealE2EUploadResponse DealE2EUploadResponse
				body, err := io.ReadAll(res.Body)
				if err != nil {
					fmt.Println(err)
				}
				err = json.Unmarshal(body, &dealE2EUploadResponse)
				fmt.Println(dealE2EUploadResponse)
			}

		}

		if chunkTask.ConnectionMode == "import" {
			var buffer = &bytes.Buffer{}
			var dealRequestArr []DealRequest
			dealRequestArr = append(dealRequestArr, dealRequest)
			req, err := http.NewRequest("POST", chunkTask.DeltaURL+"/api/v1/deal/imports", buffer)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+chunkTask.DeltaApiKey)
			client := &http.Client{}
			var res *http.Response
			res, err = client.Do(req)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println(res.StatusCode)
			fmt.Println(res.Body)
			if res.StatusCode == 200 {
				var dealE2EUploadResponse DealE2EUploadResponse
				body, err := io.ReadAll(res.Body)
				if err != nil {
					fmt.Println(err)
				}
				err = json.Unmarshal(body, &dealE2EUploadResponse)
				fmt.Println(dealE2EUploadResponse)
			}
		}

	}
	return nil
}
