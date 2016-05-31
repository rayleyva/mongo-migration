package main

import (
	"fmt"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/gizak/termui"
	"gopkg.in/mgo.v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "mongo-migration"
	app.Usage = "mongo-migration --from <mongo-host-origin> --collection-in <collection-to-migrate> --to <mongo-target> --collection-out <collection-destinity>"
	var collectionIn, collectionOut, fromDb, toDb, fromUrl, toUrl string

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "in",
			Value:       "input",
			Usage:       "collection input to migrate",
			Destination: &collectionIn,
		},
		cli.StringFlag{
			Name:        "out",
			Value:       "output",
			Usage:       "collection output name to be migrated",
			Destination: &collectionOut,
		},
		cli.StringFlag{
			Name:        "from",
			Value:       "mongodb://localhost:27017/example",
			Usage:       "mongo url from origin where is the collection to migrate",
			Destination: &fromUrl,
		},
		cli.StringFlag{
			Name:        "to",
			Value:       "mongodb://localhost:27017/example2",
			Usage:       "mongo url destination where will be collection to migrate",
			Destination: &toUrl,
		},
	}
	app.Action = func(c *cli.Context) {

		fmt.Println("get session from url: ", fromUrl)
		fromSession, err := getSession(fromUrl)
		if err != nil {
			fmt.Println("ops!: ", err)
			panic(err)
		}
		fmt.Println("get session to url: ", toUrl)
		toSession, err := getSession(toUrl)

		if err != nil {
			fmt.Println("ops!: ", err)
			panic(err)
		}
		fmt.Println("o")

		defer fromSession.Close()

		toSession.SetMode(mgo.Monotonic, true)
		fromSession.SetMode(mgo.Monotonic, true)

		from := InstanceInfo{Session: fromSession, Database: fromDb, CollectionName: collectionIn}
		to := InstanceInfo{Session: toSession, Database: toDb, CollectionName: collectionOut}

		err = termui.Init()
		if err != nil {
			panic(err)
		}
		defer termui.Close()

		strs := []string{
			"[q] [quit](fg-red)",
			"[d] [debug](switch to log in debug mode)",
			"[s] [start](fg-white,bg-green)"}

		ls := termui.NewList()
		ls.Items = strs
		ls.ItemFgColor = termui.ColorYellow
		ls.BorderLabel = "Press key to action"
		ls.Height = 7
		ls.Width = 25
		ls.Y = 0
		termui.Render(ls)
		started := false

		handleMigration := HandleMigration{false, false, false}

		termui.Handle("/sys/kbd/q", func(termui.Event) {
			handleMigration.Stop = true
			for !handleMigration.Stopped {
				fmt.Print(".")
				time.Sleep(1000 * time.Millisecond)
			}
			termui.StopLoop()
		})

		termui.Handle("/sys/kbd/s", func(termui.Event) {
			if !started {
				ImportCollection(&from, &to, &handleMigration)
				started = true
			}
		})
		termui.Handle("/sys/kbd/d", func(termui.Event) {
			handleMigration.LogMode = !handleMigration.LogMode
		})

		termui.Loop()
	}

	app.Run(os.Args)
}
func getSession(uri string) (*mgo.Session, error) {

	dialInfo, err := mgo.ParseURL(uri)

	if err != nil {
		fmt.Println("Failed to parse URI: ", uri, err)
		os.Exit(1)
	}

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		fmt.Println("Failed to connect: ", err)
		os.Exit(1)
	}

	return session, err
}
