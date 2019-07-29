package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/google/go-github/github"
	"golang.org/x/crypto/ssh/terminal"
)

func FetchOrganizationRepositories(client *github.Client) ([]*github.Repository, error) {
//	client := github.NewClient(nil)
	repos, _, err := client.Repositories.ListByOrg(context.Background(), "leetestorg", nil)
	return repos, err
}

func BasicAuthentication() (*github.Client, string) {
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