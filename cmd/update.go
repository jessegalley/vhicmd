package cmd

import (
	"fmt"
	"strings"

	"github.com/jessegalley/vhicmd/api"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"set"},
	Short:   "Update resource attributes",
}

var updateImageVisibilityCmd = &cobra.Command{
	Use:   "image-visibility <image> <visibility>",
	Short: "Update image visibility (public, private, shared, community)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageID := args[0]
		visibility := strings.ToLower(args[1])

		if visibility != "public" && visibility != "private" &&
			visibility != "shared" && visibility != "community" {
			return fmt.Errorf("visibility must be one of: public, private, shared, community")
		}

		imageURL, err := validateTokenEndpoint(tok, "image")
		if err != nil {
			return err
		}

		id, err := api.GetImageIDByName(imageURL, tok.Value, imageID)
		if err == nil {
			imageID = id
		}

		err = api.UpdateImageVisibility(imageURL, tok.Value, imageID, visibility)
		if err != nil {
			return err
		}

		fmt.Printf("Updated visibility of image %s to %s\n", imageID, visibility)
		return nil
	},
}

var updateImageMemberStatusCmd = &cobra.Command{
	Use:   "image-member-status <image> <member-id> <status>",
	Short: "Update image member status (accepted, rejected, pending)",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageID := args[0]
		memberID := args[1]
		status := strings.ToLower(args[2])

		if status != "accepted" && status != "rejected" && status != "pending" {
			return fmt.Errorf("status must be one of: accepted, rejected, pending")
		}

		imageURL, err := validateTokenEndpoint(tok, "image")
		if err != nil {
			return err
		}

		id, err := api.GetImageIDByName(imageURL, tok.Value, imageID)
		if err == nil {
			imageID = id
		}

		err = api.UpdateImageMemberStatus(imageURL, tok.Value, imageID, memberID, status)
		if err != nil {
			return err
		}

		fmt.Printf("Updated member %s status to %s for image %s\n", memberID, status, imageID)
		return nil
	},
}

func init() {
	updateCmd.AddCommand(updateImageVisibilityCmd)
	updateCmd.AddCommand(updateImageMemberStatusCmd)
	rootCmd.AddCommand(updateCmd)
}
