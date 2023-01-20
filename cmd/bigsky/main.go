package main

import (
	"context"

	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli/v2"
	"github.com/whyrusleeping/gosky/api"
	"github.com/whyrusleeping/gosky/bgs"
	"github.com/whyrusleeping/gosky/carstore"
	cliutil "github.com/whyrusleeping/gosky/cmd/gosky/util"
	"github.com/whyrusleeping/gosky/events"
	"github.com/whyrusleeping/gosky/indexer"
	"github.com/whyrusleeping/gosky/notifs"
	"github.com/whyrusleeping/gosky/repomgr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"gorm.io/plugin/opentelemetry/tracing"
)

var log = logging.Logger("bigsky")

func init() {
	logging.SetAllLoggers(logging.LevelDebug)
}

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name: "jaeger",
		},
		&cli.StringFlag{
			Name:  "db",
			Value: "sqlite=bgs.db",
		},
		&cli.StringFlag{
			Name:  "carstoredb",
			Value: "sqlite=carstore.db",
		},
		&cli.StringFlag{
			Name:  "carstore",
			Value: "bgscarstore",
		},
		&cli.BoolFlag{
			Name: "dbtracing",
		},
		&cli.StringFlag{
			Name:  "plc",
			Usage: "hostname of the plc server",
			Value: "https://plc.directory",
		},
	}

	app.Action = func(cctx *cli.Context) error {

		if cctx.Bool("jaeger") {
			url := "http://localhost:14268/api/traces"
			exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
			if err != nil {
				return err
			}
			tp := tracesdk.NewTracerProvider(
				// Always be sure to batch in production.
				tracesdk.WithBatcher(exp),
				// Record information about this application in a Resource.
				tracesdk.WithResource(resource.NewWithAttributes(
					semconv.SchemaURL,
					semconv.ServiceNameKey.String("bgs"),
					attribute.String("environment", "test"),
					attribute.Int64("ID", 1),
				)),
			)

			otel.SetTracerProvider(tp)
		}

		dbstr := cctx.String("db")

		db, err := cliutil.SetupDatabase(dbstr)
		if err != nil {
			return err
		}

		if cctx.Bool("dbtracing") {
			if err := db.Use(tracing.NewPlugin()); err != nil {
				return err
			}
		}

		cardb, err := cliutil.SetupDatabase(cctx.String("carstoredb"))
		if err != nil {
			return err
		}

		csdir := cctx.String("carstore")
		cstore, err := carstore.NewCarStore(cardb, csdir)
		if err != nil {
			return err
		}

		repoman := repomgr.NewRepoManager(db, cstore)

		evtman := events.NewEventManager()

		go evtman.Run()

		// not necessary to generate notifications, should probably make the
		// indexer just take optional callbacks for notification stuff
		notifman := notifs.NewNotificationManager(db, repoman.GetRecord)

		didr := &api.PLCServer{Host: cctx.String("plc")}

		ix, err := indexer.NewIndexer(db, notifman, evtman, didr)
		if err != nil {
			return err
		}

		repoman.SetEventHandler(func(ctx context.Context, evt *repomgr.RepoEvent) {
			if err := ix.HandleRepoEvent(ctx, evt); err != nil {
				log.Errorw("failed to handle repo event", "err", err)
			}
		})

		bgs := bgs.NewBGS(db, ix, repoman, evtman, didr)

		return bgs.Start(":2470")
	}

	app.RunAndExitOnError()
}
