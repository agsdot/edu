package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/harrybrwn/edu/cmd/internal"
	"github.com/harrybrwn/edu/cmd/internal/config"
	"github.com/harrybrwn/edu/cmd/internal/files"
	"github.com/harrybrwn/edu/cmd/internal/opts"
	"github.com/harrybrwn/errs"
	"github.com/harrybrwn/go-canvas"
	"github.com/spf13/cobra"
)

// Conf is the global config
var Conf = &Config{}

// Config is the main configuration struct
type Config struct {
	Host         string `yaml:"host" default:"canvas.instructure.com"`
	Editor       string `yaml:"editor" env:"EDITOR"`
	BaseDir      string `yaml:"basedir" default:"$HOME/.edu/files"`
	Token        string `yaml:"token" env:"CANVAS_TOKEN"`
	TwilioNumber string `yaml:"twilio_number"`
	Twilio       struct {
		SID    string `yaml:"sid" env:"TWILIO_SID"`
		Token  string `yaml:"token" env:"TWILIO_TOKEN"`
		Number string `yaml:"number"`
	} `yaml:"twilio"`
	Notifications bool `yaml:"notifications" default:"true"`
	Registration  struct {
		Term string `yaml:"term"`
		Year int    `yaml:"year"`
	} `yaml:"registration"`
	Watch struct {
		Duration string `yaml:"duration" default:"12h"`
		CRNs     []int  `yaml:"crns"`
		Term     string `yaml:"term"`
		Year     int    `yaml:"year"`
		Files    bool   `yaml:"files"`
		Subject  string `yaml:"subject"`
	} `yaml:"watch"`
	Replacements       []files.Replacement            `yaml:"replacements"`
	CourseReplacements map[string][]files.Replacement `yaml:"course-replacements"`
}

// All returns all the commands.
func All(globals *opts.Global) []*cobra.Command {
	canvasCmd.AddCommand(
		canvasCommands(globals)...,
	)
	all := []*cobra.Command{
		newCoursesCmd(globals),
		newConfigCmd(),
		canvasCmd,
		newUpdateCmd(),
		newRegistrationCmd(globals),
		newTextCmd(),
	}
	if runtime.GOOS == "linux" {
		all = append(all, genServiceCmd())
	}
	return all
}

func newCoursesCmd(opts *opts.Global) *cobra.Command {
	var all bool
	c := &cobra.Command{
		Use:     "courses",
		Short:   "Show info on courses",
		Aliases: []string{"course", "crs"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err     error
				courses []*canvas.Course
			)
			courses, err = internal.GetCourses(all)
			if err != nil {
				return internal.HandleAuthErr(err)
			}
			tab := internal.NewTable(cmd.OutOrStderr())
			header := []string{"id", "name", "uuid", "code", "ends"}
			internal.SetTableHeader(tab, header, !opts.NoColor)
			for _, c := range courses {
				tab.Append([]string{fmt.Sprintf("%d", c.ID), c.Name, c.UUID, c.CourseCode, c.EndAt.Format("01/02/06")})
			}
			tab.Render()
			return nil
		},
	}
	flags := c.Flags()
	flags.BoolVarP(&all, "all", "a", all, "show all courses (defaults to only active courses)")
	return c
}

func newConfigCmd() *cobra.Command {
	var file, edit bool
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Manage configuration variables.",
		Aliases: []string{"conf"},
		RunE: func(cmd *cobra.Command, args []string) error {
			f := config.FileUsed()
			if file {
				fmt.Println(f)
				return nil
			}
			if edit {
				if f == "" {
					return errs.New("no config file found")
				}
				editor := config.GetString("editor")
				if editor == "" {
					return errs.New("no editor set (use $EDITOR or set it in the config)")
				}
				ex := exec.Command(editor, f)
				ex.Stdout, ex.Stderr, ex.Stdin = os.Stdout, os.Stderr, os.Stdin
				return ex.Run()
			}
			return cmd.Usage()
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use: "get", Short: "Get a config variable",
		Run: func(c *cobra.Command, args []string) {
			for _, arg := range args {
				c.Println(config.Get(arg))
			}
		}})
	cmd.Flags().BoolVarP(&edit, "edit", "e", false, "edit the config file")
	cmd.Flags().BoolVarP(&file, "file", "f", false, "print the config file path")
	return cmd
}
