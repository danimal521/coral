package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// with go modules disabled

func main() {

	fAction := flag.String("action", "deploy", "action to take")
	fWhat := flag.String("what", "bigbang", "what action")
	fGHPAT := flag.String("ghpat", "", "GitHub personal access token")
	fRepoName := flag.String("repo", "", "GitHub repo name")
	flag.Parse()

	strAction := strings.ToLower(*fAction)
	strWhat := strings.ToLower(*fWhat)
	strGHPAT := *fGHPAT
	strGHRepoName := *fRepoName

	if strAction == "deploy" && strWhat == "bigbang" {
		fmt.Println("Deploying bigbang")

		if strGHPAT == "" || strGHRepoName == "" {
			fmt.Println("Please supply a GH PAT and GH Repo name")
			return
		}

		DeployBigBang(strGHPAT, strGHRepoName)
	}
}

func DeployBigBang(strGHPAT string, strGHRepoName string) {
	fmt.Println("Authenticate to GH")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strGHPAT},
	)

	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	strRepoName := strGHRepoName

	fmt.Println("Create repo:", strGHRepoName)
	repo := &github.Repository{
		Name:    github.String(strRepoName),
		Private: github.Bool(true),
	}

	repo, _, err := client.Repositories.Create(ctx, "", repo)
	if err != nil {
		fmt.Println(err)
	}

	AddFile(".sops.yaml", strRepoName, "danimal521", client)
	AddFile("base/bigbang-dev-cert.yaml", strRepoName, "danimal521", client)
	AddFile("base/configmap.yaml", strRepoName, "danimal521", client)
	AddFile("base/kustomization.yaml", strRepoName, "danimal521", client)

	AddFile("dev/bigbang.yaml", strRepoName, "danimal521", client)
	AddFile("dev/configmap.yaml", strRepoName, "danimal521", client)
	AddFile("dev/kustomization.yaml", strRepoName, "danimal521", client)
}

func ByteUrlToLines(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	buf := make([]byte, resp.ContentLength)

	totalLen, err := reader.Read(buf)

	fmt.Println("Fork: ", url)
	fmt.Println("data: ", totalLen)

	return buf, err
}

func AddFile(namewithpath string, repo string, owner string, client *github.Client) error {
	b, err := ByteUrlToLines("https://repo1.dso.mil/platform-one/big-bang/customers/template/-/raw/main/" + namewithpath)

	_, resp, err := client.Repositories.CreateFile(
		context.Background(),
		owner,
		repo,
		namewithpath,
		&github.RepositoryContentFileOptions{
			Content: b,
			Message: github.String("BigBang Template"),
			SHA:     nil,
		})

	fmt.Println("file written status code: ", resp.StatusCode)
	return err
}
