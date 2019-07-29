package main

import (
	"bufio"
	"context"
	"fmt"
	"flag"
	"errors"
	"io/ioutil"
	"time"
	"log"

	"os"
	"strings"
	"syscall"

	"github.com/google/go-github/github"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	// owner
	sourceOwner   = flag.String("source-owner", "leetestorg", "Name of the owner (user or org) of the repo to create the commit in.")
	// repo
	sourceRepo    = flag.String("source-repo", "Test1", "Name of repo to create the commit in.")
	// Message
	commitMessage = flag.String("commit-message", "Add License to project", "Content of the commit message.")
	// branch name
	commitBranch  = flag.String("commit-branch", "license", "Name of branch to create the commit in. If it does not already exists, it will be created using the `base-branch` parameter")
	baseBranch    = flag.String("base-branch", "master", "Name of branch to create the `commit-branch` from.")
	// owner
	prRepoOwner   = flag.String("merge-repo-owner", "leetestorg", "Name of the owner (user or org) of the repo to create the PR against. If not specified, the value of the `-source-owner` flag will be used.")
	// pull request repo
	prRepo        = flag.String("merge-repo", "Test1", "Name of repo to create the PR against. If not specified, the value of the `-source-repo` flag will be used.")
	prBranch      = flag.String("merge-branch", "master", "Name of branch to create the PR against (the one you want to merge your branch in via the PR).")
	// title
	prSubject     = flag.String("pr-title", "Pull request title", "Title of the pull request. If not specified, no pull request will be created.")
	// description
	prDescription = flag.String("pr-text", "Description of pull request", "Text to put in the description of the pull request.")
	// source file
	sourceFiles   = flag.String("files", "LICENSE", `Comma-separated list of files to commit and their location.
The local file is separated by its target location by a semi-colon.
If the file should be in the same location with the same name, you can just put the file name and omit the repetition.
Example: README.md,main.go:github/examples/commitpr/main.go`)
	// author name
	authorName  = flag.String("author-name", "leebrucec", "Name of the author of the commit.")
	// author email
	authorEmail = flag.String("author-email", "leebrucec@gmail.com", "Email of the author of the commit.")
)


var ctx = context.Background()

//*
// getRef returns the commit branch reference object if it exists or creates it
// from the base branch before returning it.
func getRef(client *github.Client) (ref *github.Reference, err error) {
	if ref, _, err = client.Git.GetRef(ctx, *sourceOwner, *sourceRepo, "refs/heads/"+*commitBranch); err == nil {
		return ref, nil
	}

	// We consider that an error means the branch has not been found and needs to
	// be created.
	if *commitBranch == *baseBranch {
		return nil, errors.New("The commit branch does not exist but `-base-branch` is the same as `-commit-branch`")
	}

	if *baseBranch == "" {
		return nil, errors.New("The `-base-branch` should not be set to an empty string when the branch specified by `-commit-branch` does not exists")
	}

	var baseRef *github.Reference
	if baseRef, _, err = client.Git.GetRef(ctx, *sourceOwner, *sourceRepo, "refs/heads/"+*baseBranch); err != nil {
		return nil, err
	}
	newRef := &github.Reference{Ref: github.String("refs/heads/" + *commitBranch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = client.Git.CreateRef(ctx, *sourceOwner, *sourceRepo, newRef)
	return ref, err
}

// getTree generates the tree to commit based on the given files and the commit
// of the ref you got in getRef.
func getTree(client *github.Client, ref *github.Reference) (tree *github.Tree, err error) {
	// Create a tree with what to commit.
	entries := []github.TreeEntry{}

	// Load each file into the tree.
	for _, fileArg := range strings.Split(*sourceFiles, ",") {
		file, content, err := getFileContent(fileArg)
		if err != nil {
			return nil, err
		}
		entries = append(entries, github.TreeEntry{Path: github.String(file), Type: github.String("blob"), Content: github.String(string(content)), Mode: github.String("100644")})
	}

	tree, _, err = client.Git.CreateTree(ctx, *sourceOwner, *sourceRepo, *ref.Object.SHA, entries)
	return tree, err
}

// getFileContent loads the local content of a file and return the target name
// of the file in the target repository and its contents.
func getFileContent(fileArg string) (targetName string, b []byte, err error) {
	var localFile string
	files := strings.Split(fileArg, ":")
	switch {
	case len(files) < 1:
		return "", nil, errors.New("empty `-files` parameter")
	case len(files) == 1:
		localFile = files[0]
		targetName = files[0]
	default:
		localFile = files[0]
		targetName = files[1]
	}

	b, err = ioutil.ReadFile(localFile)
	return targetName, b, err
}

// createCommit creates the commit in the given reference using the given tree.
func pushCommit(client *github.Client, ref *github.Reference, tree *github.Tree) (err error) {
	// Get the parent commit to attach the commit to.
	parent, _, err := client.Repositories.GetCommit(ctx, *sourceOwner, *sourceRepo, *ref.Object.SHA)
	if err != nil {
		return err
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA

	// Create the commit using the tree.
	date := time.Now()
	author := &github.CommitAuthor{Date: &date, Name: authorName, Email: authorEmail}
	commit := &github.Commit{Author: author, Message: commitMessage, Tree: tree, Parents: []github.Commit{*parent.Commit}}
	newCommit, _, err := client.Git.CreateCommit(ctx, *sourceOwner, *sourceRepo, commit)
	if err != nil {
		return err
	}

	// Attach the commit to the master branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = client.Git.UpdateRef(ctx, *sourceOwner, *sourceRepo, ref, false)
	return err
}

func createPR(client *github.Client) (err error) {
	if *prSubject == "" {
		return errors.New("missing `-pr-title` flag; skipping PR creation")
	}

	if *prRepoOwner != "" && *prRepoOwner != *sourceOwner {
		*commitBranch = fmt.Sprintf("%s:%s", *sourceOwner, *commitBranch)
	} else {
		prRepoOwner = sourceOwner
	}

	if *prRepo == "" {
		prRepo = sourceRepo
	}

	newPR := &github.NewPullRequest{
		Title:               prSubject,
		Head:                commitBranch,
		Base:                prBranch,
		Body:                prDescription,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(ctx, *prRepoOwner, *prRepo, newPR)
	if err != nil {
		return err
	}

	fmt.Printf("PR created: %s\n", pr.GetHTMLURL())
	return nil
}

//*/

func FetchOrganizationRepositories(client *github.Client) ([]*github.Repository, error) {
//	client := github.NewClient(nil)
	repos, _, err := client.Repositories.ListByOrg(context.Background(), "leetestorg", nil)
	return repos, err
}

func BasicAuthentication() ( *github.Client, string) {
	r := bufio.NewReader(os.Stdin)
	fmt.Print("GitHub Username: ")
	username, _ := r.ReadString('\n')

	fmt.Print("GitHub Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	password := string(bytePassword)

	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}

	client := github.NewClient(tp.Client())

	return client, username
}

/*
func FetchRepositoryLicense(repo Repository) (License, error) {
	repos, _, err := repo.RepositoriesService.License
	repos, _, err := client.Repositories.ListByOrg(context.Background(), "leetestorg", nil)
	return repos, err
}
*/

func main() {
	if *sourceOwner == "" || *sourceRepo == "" || *commitBranch == "" || *sourceFiles == "" || *authorName == "" || *authorEmail == "" {
		log.Fatal("You need to specify a non-empty value for the flags `-source-owner`, `-source-repo`, `-commit-branch`, `-files`, `-author-name` and `-author-email`")
	}


	client, username := BasicAuthentication();

	fmt.Printf("username = %v\n", username)

//	if err != nil {
//		fmt.Printf("\nerror: %v\n", err)
//		return
//	}

	repos, err := FetchOrganizationRepositories(client)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	} else {
		for i, repo := range repos {
			fmt.Printf("%v. %v\n", i+1, repo.GetName())
			license := repo.GetLicense()
			if license != nil {
				fmt.Printf("\tName = %v\n", license.GetName())
				fmt.Printf("\tURL =  %v\n", license.GetURL())
				fmt.Printf("\tKey =  %v\n", license.GetKey())
			}
		}
	}



	ref, err := getRef(client)
	if err != nil {
		log.Fatalf("Unable to get/create the commit reference: %s\n", err)
	}
	if ref == nil {
		log.Fatalf("No error where returned but the reference is nil")
	}

	tree, err := getTree(client, ref)
	if err != nil {
		log.Fatalf("Unable to create the tree based on the provided files: %s\n", err)
	}

	if err := pushCommit(client, ref, tree); err != nil {
		log.Fatalf("Unable to create the commit: %s\n", err)
	}

	if err := createPR(client); err != nil {
		log.Fatalf("Error while creating the pull request: %s", err)
	}



	// fmt.Printf("\n%v\n", github.Stringify(user))

	//username := user.GetName()

/*
	fmt.Printf("------- Organizations: %v\n", username)
	organizations, err := FetchOrganizations(client, username)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	} else {
		for i, organization := range organizations {
			fmt.Printf("%v. %v\n", i+1, organization.GetLogin())
		}
	}
	fmt.Printf("------ Members: %v\n", username)
	members, err := FetchOrganizationMembers(username)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	} else {
		for i, member := range members {
			fmt.Printf("%v. %v\n", i+1, member.GetLogin())
		}
	}
	fmt.Printf("------ repositories: %v\n", username)
	repos, err := FetchOrganizationRepositories(username)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	} else {
		for i, repo := range repos {
			fmt.Printf("%v. %v\n", i+1, repo.GetName())
			license := repo.GetLicense()
			if license != nil {
				fmt.Printf("\tName = %v\n", license.GetName())
				fmt.Printf("\tURL =  %v\n", license.GetURL())
				fmt.Printf("\tKey =  %v\n", license.GetKey())
			}
		}
	}
	*/
}