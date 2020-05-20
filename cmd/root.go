package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/harrybrwn/errs"
	"github.com/harrybrwn/go-canvas"
	table "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version string

// Execute will execute the root comand on the cli
func Execute() (err error) {
	if err = viper.ReadInConfig(); err != nil {
		return err
	}

	root.AddCommand(
		newFilesCmd(),
		newConfigCmd(),
		newCoursesCmd(),
		newUpdateCmd(),
		completionCmd,
	)

	err = root.Execute()
	if err == nil {
		return nil
	}
	errorMessage(err)
	os.Exit(1)
	return nil
}

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$XDG_CONFIG_HOME/canvas")
	viper.AddConfigPath("$HOME/.config/canvas")
	viper.AddConfigPath("$HOME/.canvas")

	viper.SetEnvPrefix("edu")
	viper.BindEnv("host")
	viper.BindEnv("canvas_token", "CANVAS_TOKEN")
	viper.BindEnv("editor", "EDITOR")
}

var (
	root = &cobra.Command{
		Use:           "edu",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			host := viper.GetString("host")
			if host != "" {
				canvas.DefaultHost = host
			}
			token := viper.GetString("token")
			if token != "" {
				canvas.SetToken(os.ExpandEnv(token))
			} else {
				viper.Set("token", viper.GetString("canvas_token"))
				canvas.SetToken(os.ExpandEnv(viper.GetString("token")))
			}
			canvas.ConcurrentErrorHandler = errorHandler
		},
	}
	completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Print a completion script to stdout.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			root := cmd.Root()
			out := cmd.OutOrStdout()
			if len(args) == 0 {
				return errors.New("no shell type given")
			}
			switch args[0] {
			case "zsh":
				return root.GenZshCompletion(out)
			case "ps", "powershell":
				return root.GenPowerShellCompletion(out)
			case "bash":
				return root.GenBashCompletion(out)
			case "fish":
				return root.GenFishCompletion(out, false)
			}
			return errs.New("unknown shell type")
		},
		ValidArgs: []string{"zsh", "bash", "ps", "powershell", "fish"},
		Aliases:   []string{"comp"},
	}
)

func newFilesCmd() *cobra.Command {
	var (
		contentType string
		sortby      = []string{"created_at"}
	)
	c := &cobra.Command{
		Use:   "files",
		Short: "This is a garbage command lol.",
		RunE: func(cmd *cobra.Command, args []string) error {
			courses, err := canvas.ActiveCourses()
			if err != nil {
				return err
			}

			opts := []canvas.Option{canvas.SortOpt(sortby...)}
			if contentType != "" {
				opts = append(opts, canvas.ContentType(contentType))
			}
			for _, course := range courses {
				course.SetErrorHandler(errorHandler)
				files := course.Files(opts...)
				for f := range files {
					fmt.Println(f.CreatedAt, f.Size, f.Filename)
				}
			}
			return nil
		},
	}
	c.Flags().StringVarP(&contentType, "content-type", "c", "", "filter out files by content type (ex. application/pdf)")
	c.Flags().StringArrayVarP(&sortby, "sortyby", "s", sortby, "how the files should be sorted")
	return c
}

func newCoursesCmd() *cobra.Command {
	var (
		all bool
	)
	c := &cobra.Command{
		Use:     "courses",
		Short:   "Show info on courses",
		Aliases: []string{"course", "c"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				err     error
				courses []*canvas.Course
			)
			courses, err = canvas.ActiveCourses()
			if err != nil {
				return err
			}
			if all {
				completed, err := canvas.CompletedCourses()
				if err != nil {
					return err
				}
				courses = append(courses, completed...)
			}
			var namelen = 1
			for _, course := range courses {
				if len(course.Name) > namelen {
					namelen = len(course.Name)
				}
			}

			tab := table.NewWriter(os.Stdout)
			header := []string{"id", "name", "uuid", "code", "ends"}
			headercolors := make([]table.Colors, len(header))
			for i := range header {
				headercolors[i] = table.Colors{table.FgCyanColor}
			}
			tab.SetBorder(false)
			tab.SetColumnSeparator("")
			tab.SetAlignment(table.ALIGN_LEFT)
			tab.SetAutoFormatHeaders(false)
			tab.SetHeaderLine(false)
			tab.SetHeaderAlignment(table.ALIGN_LEFT)
			tab.SetHeader(header)
			tab.SetHeaderColor(headercolors...)
			for _, c := range courses {
				tab.Append([]string{fmt.Sprintf("%d", c.ID), c.Name, c.UUID, c.CourseCode, c.EndAt.Format("01/02/06")})
			}
			tab.Render()
			return nil
		},
	}
	c.Flags().BoolVarP(&all, "all", "a", all, "show all courses (defaults to only active courses)")
	return c
}

func newConfigCmd() *cobra.Command {
	var file, edit bool
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Manage configuration",
		Aliases: []string{"conf"},
		RunE: func(cmd *cobra.Command, args []string) error {
			f := viper.ConfigFileUsed()
			if file {
				cmd.Println(f)
				return nil
			}
			if edit {
				editor := viper.GetString("editor")
				ex := exec.Command(editor, f)
				ex.Stdout, ex.Stderr, ex.Stdin = os.Stdout, os.Stderr, os.Stdin
				return ex.Run()
			}
			if len(args) > 1 {
				if strings.ToLower(args[0]) == "get" {
					cmd.Println(viper.Get(args[0]))
					return nil
				}
			}
			return cmd.Usage()
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use: "get", Short: "Get a config variable",
		Run: func(c *cobra.Command, args []string) {
			for _, arg := range args {
				c.Println(viper.Get(arg))
			}
		}})
	cmd.Flags().BoolVarP(&edit, "edit", "e", false, "edit the config file")
	cmd.Flags().BoolVarP(&file, "file", "f", false, "print the config file path")
	return cmd
}

func mkdir(d string) error {
	if _, err := os.Stat(d); os.IsNotExist(err) {
		if err = os.MkdirAll(d, 0775); err != nil {
			return err
		}
	}
	return nil
}

func errorMessage(err error) {
	switch err.(type) {
	case *canvas.AuthError:
		fmt.Fprintf(os.Stderr, "Authentication Error: %v\n", err)
	case *canvas.Error:
		fmt.Fprintf(os.Stderr, "Canvas Error: %v\n", err)
	default:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func errorHandler(e error, stop chan int) {
	if e != nil {
		stop <- 1
		fmt.Println("Error: " + e.Error())
		os.Exit(1)
	}
}

func stop(msg string) {
	errmsg(msg)
	os.Exit(1)
}

func errmsg(msg interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", msg)
}
