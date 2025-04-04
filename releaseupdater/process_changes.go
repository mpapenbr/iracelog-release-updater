package releaseupdater

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v44/github"
	"github.com/ktrysmt/go-bitbucket"
)

func ProcessNewRelease(ctx Context, release *github.ReleaseEvent) {
	fmt.Printf("Incoming release event from %s\n", *release.Repo.FullName)
	for i := range ctx.Config.Actions {
		action := ctx.Config.Actions[i]
		if action.From == *release.Repo.Name {
			commitComponent := release.Repo.Name
			if action.Component != "" {
				commitComponent = &action.Component
			}
			for j := range action.Update {
				update := action.Update[j]
				log.Printf("%v\n", update)
				repoOwner := *release.Repo.Owner.Login
				replacer := func(content string) (string, string) {
					return ReplaceVersionString(content, update.Regex, *release.Release.TagName),
						fmt.Sprintf(
							"pkg: Bump %s to %s",
							*commitComponent,
							*release.Release.TagName,
						)
				}
				switch strings.ToLower(update.RepoType) {
				case "bitbucket":
					handleBitbucket(ctx, update, replacer)
				default:
					handleGithub(ctx, repoOwner, update, replacer)
				}
			}
		}
	}
}

//nolint:whitespace //can't make all linters happy
func handleGithub(
	ctx Context, repoOwner string,
	update Update,
	processContent func(content string) (string, string),
) {
	targetBranch := getBranch(update)
	for _, toUpdateFile := range update.Files {
		content, _, _, err := ctx.ProbotCtx.GitHub.Repositories.GetContents(
			context.Background(),
			repoOwner,
			update.Repo,
			toUpdateFile,
			&github.RepositoryContentGetOptions{})
		if err != nil {
			log.Printf("error reading source: %+v\n", err)
			continue
		}
		fileContent, _ := b64.StdEncoding.DecodeString(*content.Content)

		log.Printf("RegEx: <%s>\n", update.Regex)

		newVersion, message := processContent(string(fileContent))
		if newVersion != string(fileContent) {
			fmt.Printf("Updating file %s\n", toUpdateFile)

			_, _, err = ctx.ProbotCtx.GitHub.Repositories.UpdateFile(
				context.Background(),
				repoOwner,
				update.Repo,
				toUpdateFile,
				&github.RepositoryContentFileOptions{
					Content: []byte(newVersion),
					Message: github.String(message),
					SHA:     github.String(*content.SHA),
					Branch:  &targetBranch,
				})
			if err != nil {
				log.Printf("error updating %s: %+v\n", toUpdateFile, err)
				continue
			}
		} else {
			log.Println("No changes detected")
		}
	}
}

//nolint:whitespace,funlen //can't make all linters happy
func handleBitbucket(
	ctx Context,
	update Update,
	processContent func(content string) (string, string),
) {
	targetBranch := getBranch(update)
	for _, toUpdateFile := range update.Files {
		fileContent, err := ctx.BitbucketClient.Repositories.Repository.GetFileBlob(
			&bitbucket.RepositoryBlobOptions{
				Owner:    "mpapenbr",
				RepoSlug: update.Repo, Path: toUpdateFile, Ref: targetBranch,
			})
		if err != nil {
			log.Printf("error reading source: %+v\n", err)
			continue
		}

		newVersion, message := processContent(fileContent.String())
		//nolint:nestif //checked
		if newVersion != fileContent.String() {
			fmt.Printf("Updating file %s\n", toUpdateFile)

			f, _ := os.CreateTemp("", "bbupload")
			if _, err = f.WriteString(newVersion); err != nil {
				log.Printf("could not write temp upload file: %+v\n", err)
				f.Close()
				continue
			}
			f.Close()

			if err = ctx.BitbucketClient.Repositories.Repository.WriteFileBlob(
				&bitbucket.RepositoryBlobWriteOptions{
					Owner:    "mpapenbr",
					RepoSlug: update.Repo,
					Files:    []bitbucket.File{{Path: f.Name(), Name: toUpdateFile}},
					Branch:   targetBranch,
					Message:  message,
				}); err != nil {
				log.Printf("error upload file to bitbucket: %+v\n", err)
			}
			log.Printf("Deleting temp file %s\n", f.Name())
			if err = os.Remove(f.Name()); err != nil {
				log.Printf("Error deleting temp file %s: %v\n", f.Name(), err)
			}
		} else {
			log.Println("No changes detected")
		}
	}
}

func getBranch(update Update) string {
	if update.Branch != "" {
		return update.Branch
	} else {
		return "main"
	}
}
