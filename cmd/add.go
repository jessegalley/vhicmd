package cmd

import (
	"fmt"

	"github.com/jessegalley/vhicmd/api"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add relationships between resources",
}

var addImageMemberCmd = &cobra.Command{
	Use:   "image-member <image> <project-id>",
	Short: "Grant project access to a shared image",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageID := args[0]
		projectID := args[1]

		imageURL, err := validateTokenEndpoint(tok, "image")
		if err != nil {
			return err
		}

		id, err := api.GetImageIDByName(imageURL, tok.Value, imageID)
		if err == nil {
			imageID = id
		}

		err = api.AddImageMember(imageURL, tok.Value, imageID, projectID)
		if err != nil {
			return err
		}

		fmt.Printf("Granted access to image %s for project %s\n", imageID, projectID)
		return nil
	},
}

func init() {
	addCmd.AddCommand(addImageMemberCmd)
	rootCmd.AddCommand(addCmd)
}
