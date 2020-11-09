package cmd

import (
	"fmt"

	"github.com/dfuse-io/solana-go"
	"github.com/dfuse-io/solana-go/token"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

var splGetMintCmd = &cobra.Command{
	Use:   "get-mint {mint_addr}",
	Short: "Retrieves mint information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		mintAddr, err := solana.PublicKeyFromBase58(args[0])
		if err != nil {
			return fmt.Errorf("decoding mint addr: %w", err)
		}

		client := getClient()

		acct, err := client.GetAccountInfo(ctx, mintAddr)
		if err != nil {
			return fmt.Errorf("couldn't get account data: %w", err)
		}

		mint, err := token.DecodeMint(acct.Value.Data)
		if err != nil {
			return fmt.Errorf("unable to retrieve int information: %w", err)
		}

		if !mint.IsInitialized {
			fmt.Println("Uninitialized mint. Data length", len(acct.Value.Data))
			return nil
		}

		var out []string

		out = append(out, fmt.Sprintf("Data length | %d", len(acct.Value.Data)))

		if mint.MintAuthorityOption != 0 {
			out = append(out, fmt.Sprintf("Mint Authority | %s", mint.MintAuthority))
		} else {
			out = append(out, "No mint authority")
		}

		out = append(out, fmt.Sprintf("Supply | %d", mint.Supply))
		out = append(out, fmt.Sprintf("Decimals | %d", mint.Decimals))

		if mint.FreezeAuthorityOption != 0 {
			out = append(out, fmt.Sprintf("Freeze Authority | %s", mint.FreezeAuthority))
		} else {
			out = append(out, "No freeze authority")
		}

		fmt.Println(columnize.Format(out, nil))

		return nil
	},
}

func init() {
	splCmd.AddCommand(splGetMintCmd)
}
