package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
)

func FetchOrganizations(username string) ([]*github.Organization, error) {
	client := github.NewClient(nil)
	orgs, _, err := client.Organizations.List(context.Background(), username, nil)
	return orgs, err
}

// ListMembersOptions specifies optional parameters to the
// OrganizationsService.ListMembers method.
func FetchOrganizationMembers(username string) ([]*github.User, error) {
	client := github.NewClient(nil)
	members, _, err := client.Organizations.ListMembers(context.Background(), "leetestorg", nil)
	return members, err
}

func FetchOrganizationRepositories(username string) ([]*github.Repository, error) {
	client := github.NewClient(nil)
	repos, _, err := client.Repositories.ListByOrg(context.Background(), "leetestorg", nil)
	return repos, err
}

/*
func FetchRepositoryLicense(repo Repository) (License, error) {
	repos, _, err := repo.RepositoriesService.License
	repos, _, err := client.Repositories.ListByOrg(context.Background(), "leetestorg", nil)
	return repos, err
}
*/

func main() {
	var username string
	fmt.Print("Enter GitHub username: ")
	fmt.Scanf("%s", &username)

	fmt.Printf("------- Organizations: %v\n", username)
	organizations, err := FetchOrganizations(username)
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
}