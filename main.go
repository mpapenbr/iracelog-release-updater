package main

import (
	"log"
	"os"

	"github.com/google/go-github/v44/github"
	"github.com/ktrysmt/go-bitbucket"
	"github.com/mpapenbr/go-probot/probot"

	"github.com/mpapenbr/iracelog-release-updater/releaseupdater"
)

func main() {
	log.Printf("iracelog-release-updater version %s\n", releaseupdater.Version)
	config, err := releaseupdater.GetConfig("config.yml")
	if err != nil {
		log.Fatal("Could not read config")
	}

	probot.HandleEvent("ping", func(ctx *probot.Context) error {
		log.Printf("Ping event received\n")
		return nil
	})

	probot.HandleEvent("release", func(ctx *probot.Context) error {
		// Because we're listening for "release" we know the payload is a *github.ReleaseEvent
		//nolint:errcheck //ok
		event := ctx.Payload.(*github.ReleaseEvent)

		if *event.Action == "published" {
			c, err := bitbucket.NewBasicAuth(
				os.Getenv("BITBUCKET_APP_USER"),
				os.Getenv("BITBUCKET_APP_PASSWORD"))
			if err != nil {
				log.Printf("Could not create Bitbucket basic auth: %v\n", err)
				return nil
			}
			localContext := releaseupdater.Context{
				Config: config, ProbotCtx: ctx, BitbucketClient: c,
			}

			if event.Release.GetPrerelease() {
				log.Printf("Prereleases are not handled here repo=%s tag=%s\n",
					*event.GetRepo().Name, *event.Release.TagName)
				return nil
			}
			log.Printf("got release published from %s\n", *event.GetRepo().Name)
			releaseupdater.ProcessNewRelease(localContext, event)
		} else {
			log.Printf("I'm only interested in published releases. (was: %s)\n", *event.Action)
		}

		return nil
	})

	probot.Start()
}
