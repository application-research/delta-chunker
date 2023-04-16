package cmd

import (
	"delta-chunker/model"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"os"
)

func CarChunkRunnerCmd() []*cli.Command {
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
			configFile := c.String("run-config")
			data, err := os.ReadFile(configFile)
			if err != nil {
				fmt.Println("Error reading config file:", err)
				return err
			}

			// parse the YAML data into Config struct
			var cfg model.Config
			err = yaml.Unmarshal(data, &cfg)
			if err != nil {
				fmt.Println("Error parsing YAML:", err)
				return err
			}

			// access the individual chunk tasks
			for _, task := range cfg.ChunkTasks {
				fmt.Printf("Task name: %s\n", task.Name)
				fmt.Printf("Source: %s\n", task.Source)
				fmt.Printf("Output directory: %s\n", task.OutputDir)
				fmt.Printf("Split size: %d\n", task.SplitSize)
				fmt.Printf("Connection mode: %s\n", task.ConnectionMode)
				fmt.Printf("Miner: %s\n", task.Miner)
				fmt.Printf("Delta URL: %s\n", task.DeltaURL)
				fmt.Printf("Delta token: %s\n", task.DeltaToken)
				fmt.Printf("Delta wallet: %s\n", task.DeltaWallet)
				fmt.Printf("Delta metadata request: %s\n", task.DeltaMetadataReq)

				// record each on the database
				
			}
			return nil
		},
	}
	carCommands = append(carCommands, carChunkerCmd)

	return carCommands
}
