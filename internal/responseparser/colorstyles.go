package responseparser

import (
	"github.com/gookit/color"
	"github.com/olekukonko/tablewriter"
)

// -------------------------------------------------------------------
// Style Helpers
// -------------------------------------------------------------------

// colorStyleBool returns a color-coded "TRUE" or "FALSE"
func colorStyleBool(value bool) string {
	if value {
		return color.Style{color.FgGreen, color.OpBold}.Render("TRUE")
	}
	return color.Style{color.FgRed, color.OpBold}.Render("FALSE")
}

// colorStyleStatus returns a color-coded status string
func colorStyleStatus(status string) string {
	switch status {
	case "ACTIVE":
		return color.Style{color.FgGreen, color.OpBold}.Render(status)
	case "ERROR":
		return color.Style{color.FgRed, color.OpBold}.Render(status)
	default:
		return color.Style{color.FgYellow, color.OpBold}.Render(status)
	}
}

// colorStyleVolAvailability returns a color-coded availability string
// Possible values:
// "available”, “error”, “creating”, “deleting”, “in-use”, “attaching”, “detaching”, “error_deleting” or “maintenance”.
func colorStyleVolAvailability(availability string) string {
	switch availability {
	case "available":
		return color.Style{color.FgGreen, color.OpBold}.Render(availability)
	case "in-use":
		return color.Style{color.FgMagenta, color.OpBold}.Render(availability)
	case "creating":
		return color.Style{color.FgYellow, color.OpBold}.Render(availability)
	case "deleting":
		return color.Style{color.FgYellow, color.OpBold}.Render(availability)
	case "attaching":
		return color.Style{color.FgYellow, color.OpBold}.Render(availability)
	case "detaching":
		return color.Style{color.FgYellow, color.OpBold}.Render(availability)
	case "error_deleting":
		return color.Style{color.FgRed, color.OpBold}.Render(availability)
	case "maintenance":
		return color.Style{color.FgYellow, color.OpBold}.Render(availability)
	default:
		return color.Style{color.FgRed, color.OpBold}.Render(availability)
	}
}

// applyTableStyle configures tablewriter styles
func applyTableStyle(table *tablewriter.Table) {
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("+")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")
	table.SetAutoWrapText(false)
	table.SetBorder(true)
}
