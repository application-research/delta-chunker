package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

func HexCmd() []*cli.Command {
	var carCommands []*cli.Command
	hexCmd := &cli.Command{
		Name:  "hex",
		Usage: "Generate the hexed version of a Delta generated wallet",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "wallet-dir",
				Usage: "Path to the wallet directory",
			},
		},
		Action: func(c *cli.Context) error {
			walletDir := c.String("wallet-dir")
			fmt.Println("wallet dir:", walletDir)
			// generate hex version of the wallet that can then be used on the `delta-wallet` field of the run config yaml.
			return nil
		},
	}
	carCommands = append(carCommands, hexCmd)

	return carCommands
}
