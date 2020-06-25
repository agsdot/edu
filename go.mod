module github.com/harrybrwn/edu

go 1.14

require (
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/gen2brain/beeep v0.0.0-20200526185328-e9c15c258e28
	github.com/harrybrwn/errs v0.0.2-0.20200523142445-e4279967174e
	github.com/harrybrwn/go-canvas v0.0.0-00010101000000-000000000000
	github.com/mitchellh/mapstructure v1.3.2
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

replace github.com/harrybrwn/go-canvas => ../../pkg/go-canvas
