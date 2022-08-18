package main

import (
	"log"
	"os"

	"github.com/ktrysmt/go-bitbucket"
	"github.com/mpapenbr/iracelog-release-updater/releaseupdater"

	"github.com/google/go-github/v44/github"
	"github.com/mpapenbr/go-probot/probot"
)

func main() {

	log.Printf("iracelog-release-updater version %s\n", releaseupdater.Version)
	config, err := releaseupdater.GetConfig("config.yml")
	if err != nil {
		log.Fatal("Could not read config")
	}

	probot.HandleEvent("ping", func(ctx *probot.Context) error {
		log.Printf("Ping event recieved\n")
		return nil
	})

	probot.HandleEvent("release", func(ctx *probot.Context) error {

		// Because we're listening for "release" we know the payload is a *github.ReleaseEvent
		event := ctx.Payload.(*github.ReleaseEvent)
		if *event.Action == "published" {

			c := bitbucket.Client(*bitbucket.NewBasicAuth(os.Getenv("BITBUCKET_APP_USER"), os.Getenv("BITBUCKET_APP_PASSWORD")))
			localContext := releaseupdater.Context{Config: config, ProbotCtx: ctx, BitbucketClient: &c}
			// log.Printf("got release published %+v\n", event)
			log.Printf("got release published from %s\n", *event.GetRepo().Name)
			releaseupdater.ProcessNewRelease(localContext, event)

		} else {
			log.Printf("I'm only interested in published releases. (was: %s)\n", *event.Action)
		}

		return nil
	})

	probot.Start()
}
