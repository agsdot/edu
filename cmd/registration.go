package cmd

import (
	"fmt"

	"github.com/harrybrwn/edu/school/ucmerced/sched"
	"github.com/harrybrwn/errs"
	table "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newRegistrationCmd() *cobra.Command {
	var (
		term string = viper.GetString("registration.term")
		year int    = viper.GetInt("registration.year")
	)
	c := &cobra.Command{
		Use:   "registration",
		Short: "Get registration data",
		Long: `Use the 'registration' command to get information on class
registration information.`,
		Aliases: []string{"reg", "register"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var subj, num string
			if len(args) >= 1 {
				subj = args[0]
			}
			if len(args) >= 2 {
				num = args[1]
			}
			if year == 0 {
				return errs.New("no year given")
			}

			schedual, err := sched.BySubject(year, term, subj) // still works with an empty subj
			if err != nil {
				return err
			}

			tab := newTable(cmd.OutOrStderr())
			header := []string{"crn", "code", "title", "activity", "time", "seats open"}
			setTableHeader(tab, header)
			tab.SetAutoWrapText(false)
			for _, c := range schedual {
				if num != "" && c.Number != num {
					continue
				}
				tab.Append([]string{
					fmt.Sprintf("%d", c.CRN),
					c.Fullcode,
					c.Title,
					c.Activity,
					c.Time,
					fmt.Sprintf("%d", c.SeatsAvailible()),
				})
			}
			tab.Render()
			return nil
		},
	}
	flags := c.Flags()
	flags.StringVar(&term, "term", term, "specify the term (spring|summer|fall)")
	flags.IntVar(&year, "year", year, "specify the year for registration")
	return c
}

func setTableHeader(t *table.Table, header []string) {
	headercolors := make([]table.Colors, len(header))
	for i := range header {
		headercolors[i] = table.Colors{table.FgCyanColor}
	}
	t.SetHeader(header)
	t.SetHeaderColor(headercolors...)
}
