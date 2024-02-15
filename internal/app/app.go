package app

import (
	"context"
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"oss.terrastruct.com/d2/d2graph"

	// "oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"

	// "oss.terrastruct.com/d2/d2oracle"
	clog "github.com/charmbracelet/log"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

type App struct {
	Ec2Client *ec2.Client

	D2CompileOptions *d2lib.CompileOptions
	D2RenderOpts     *d2svg.RenderOpts
}

func New() (*App, error) {
	app := App{}

	var awsConfig aws.Config
	localstack := flag.Bool("localstack", false, "")
	localstackHost := flag.String("localstack-host", "localhost:4566", "")
	verbose := flag.Bool("verbose", false, "")
	flag.Parse()
	if *verbose {
		clog.SetLevel(clog.DebugLevel)
	}
	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	app.Ec2Client = ec2.NewFromConfig(awsConfig, func(o *ec2.Options) {
		if *localstack {
			clog.Info("using localstack endpoints")
			o.BaseEndpoint = aws.String(fmt.Sprintf("http://%s", *localstackHost))
			o.Credentials = aws.AnonymousCredentials{}
		}
	})

	ruler, _ := textmeasure.NewRuler()

	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2elklayout.DefaultLayout, nil
	}
	app.D2CompileOptions = &d2lib.CompileOptions{
		LayoutResolver: layoutResolver,
		Ruler:          ruler,
	}
	app.D2RenderOpts = &d2svg.RenderOpts{
		Pad:     go2.Pointer(int64(5)),
		ThemeID: &d2themescatalog.GrapeSoda.ID,
	}

	return &app, nil

}
