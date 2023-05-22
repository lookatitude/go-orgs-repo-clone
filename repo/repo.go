package repo

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/google/go-github/github"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"golang.org/x/oauth2"
	"archive/tar"
	"compress/gzip"
	"io"
	"path/filepath"
	"strings"
)

type RepoData struct {
	CloneURL string
	Name     string
}

func GetAllRepos(org string, token string) ([]*RepoData, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}
	var allRepos []*RepoData
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("HTTP request error: %s", resp.Status)
		}
		for _, repo := range repos {
			allRepos = append(allRepos, &RepoData{*repo.CloneURL, *repo.Name})
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func CloneAndCompressRepos(repos []*RepoData, token string, progress func()) error {
	var wg sync.WaitGroup
	for _, repo := range repos {
		wg.Add(1)
		go func(repo *RepoData) {
			defer wg.Done()
			log.Printf("Cloning repo: %s\n", repo.CloneURL)
			err := cloneRepo(repo.CloneURL, token)
			if err != nil {
				log.Println(err)
				return
			}
			err = compressRepo(repo.Name)
			if err != nil {
				log.Println(err)
				return
			}
			progress()
		}(repo)
	}
	wg.Wait()

	return nil
}

func cloneRepo(url string, token string) error {
	_, err := git.PlainClone("./"+url, false, &git.CloneOptions{
		URL: url,
		Auth: &http.BasicAuth{
			Username: "abc123", // this can be anything except an empty string
			Password: token,
		},
	})
	return err
}

func compressRepo(name string) error {
	// tar.gz the repo directory
	fw, err := os.Create(name + ".tar.gz")
	if err != nil {
		return err
	}
	defer fw.Close()

	gw := gzip.NewWriter(fw)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	if err := addFiles(tw, name, ""); err != nil {
		return err
	}

	// delete the repo directory
	if err := os.RemoveAll(name); err != nil {
		return err
	}

	return nil
}

func addFiles(tw *tar.Writer, basePath, baseInTar string) error {
	return filepath.Walk(basePath, func(fn string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		if baseInTar != "" {
			header.Name = filepath.Join(baseInTar, strings.TrimPrefix(fn, basePath))
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(fn)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})
}
