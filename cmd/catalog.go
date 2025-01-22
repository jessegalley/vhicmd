package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jessegalley/vhicmd/api"
	"github.com/jessegalley/vhicmd/internal/responseparser"
	"github.com/spf13/cobra"
)

var (
	flagInterface string // New flag to filter by interface
)

var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Fetch and display the OpenStack service catalog",
	Long:  "Fetches the service catalog from the OpenStack Identity API and displays the available services and their endpoints.",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := api.GetCatalog(vhiHost, authToken)
		if err != nil {
			return err
		}

		// Filter by interface type (default: public)
		filterInterface := strings.ToLower(flagInterface)
		if filterInterface == "" {
			filterInterface = "public" // Default to public
		}

		// Filter the catalog
		var filteredCatalog []responseparser.CatalogEntry
		for _, svc := range resp.Catalog {
			for _, ep := range svc.Endpoints {
				if strings.ToLower(ep.Interface) == filterInterface {
					filteredCatalog = append(filteredCatalog, responseparser.CatalogEntry{
						Type:      svc.Type,
						Name:      svc.Name,
						Interface: ep.Interface,
						Region:    ep.Region,
						URL:       ep.URL,
					})
				}
			}
		}

		// JSON output if flagJsonOutput is set
		if flagJsonOutput {
			b, _ := json.MarshalIndent(filteredCatalog, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		// Table output
		responseparser.PrintCatalogTable(filteredCatalog)
		return nil
	},
}

func init() {
	// Register the `catalog` command with the root
	rootCmd.AddCommand(catalogCmd)

	// JSON output flag
	catalogCmd.Flags().BoolVar(
		&flagJsonOutput,
		"json",
		false,
		"Output in JSON format (instead of a table).",
	)

	// Interface filter flag
	catalogCmd.Flags().StringVarP(
		&flagInterface,
		"interface",
		"i",
		"public",
		"Filter endpoints by interface type (public, private, admin).",
	)
}
