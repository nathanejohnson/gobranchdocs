package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/pkg/browser"

	"golang.org/x/mod/modfile"
)



func main() {

	var (
		gitPath    string
		pkgGoURL   string
		proxyGoURL string
	)
	fs := flag.NewFlagSet("gobranchdocs", flag.ExitOnError)
	fs.StringVar(&pkgGoURL, "pkg-go-dev-url", "https://pkg.go.dev", "pkg.go.dev url")
	fs.StringVar(&proxyGoURL, "proxy-go-url", "https://proxy.golang.org", "proxy.golang.org url")

	fs.Parse(os.Args[1:])

	args := fs.Args()
	if len(args) == 0 {
		gitPath = "."
	} else {
		gitPath = args[0]
	}

	h, err := getHeadSHA(gitPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Printf("head hash: %s\n", h)
	modName, err := getModuleNameFromGoMod(gitPath)
	if err != nil {
		fmt.Printf("err from go mod lookup: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("module name: %s\n", modName)

	u, err := generateURLFromModName(pkgGoURL, proxyGoURL, modName, h)
	if err != nil {
		fmt.Printf("error generating URL: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("got url: %s\n", u.String())
	err = browser.OpenURL(u.String())
	if err != nil {
		fmt.Printf("error opening browser: %s\n", err)
		os.Exit(1)
	}
}

func getHeadSHA(gitPath string) (string, error) {
	repo, err := git.PlainOpenWithOptions(gitPath, &git.PlainOpenOptions{
		DetectDotGit:          true,
		EnableDotGitCommonDir: false,
	})
	if err != nil {
		log.Printf("error opening git repo %s: %s", gitPath, err)
		os.Exit(1)
	}
	wt, err := repo.Worktree()
	if err != nil {
		log.Printf("error fetching work tree: %s", err)
		os.Exit(1)
	}
	status, err := wt.Status()
	if err != nil {
		log.Printf("error getting status: %s", err)
		return "", err
	}

	fmt.Printf("status: %s\n", status)

	l, err := repo.Log(&git.LogOptions{})

	if err != nil {
		return "", fmt.Errorf("error getting log: %w", err)
	}
	c, err := l.Next()
	if err != nil {
		err = fmt.Errorf("error getting log: %w", err)
		return "", err
	}
	l.Close()
	h := c.Hash
	return h.String(), nil
}

func generateURLFromModName(baseURL, proxyGoURL, modName, headSHA string) (*url.URL, error) {
	u, err := url.Parse(proxyGoURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, modName, "@v", headSHA+".info")
	log.Printf("fetching %s", u)
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	v := struct {
		Time    string
		Version string
	}{}
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&v)
	if err != nil {
		return nil, err
	}
	version := v.Version
	u2, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u2.Path = path.Join(u2.Path, modName + "@" + version)
	return u2, nil
}

func getModuleNameFromGoMod(dir string) (string, error) {
	goModPath := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	f, err := modfile.ParseLax("go.mod", data, nil)
	if err != nil {
		return "", err
	}
	return f.Module.Mod.String(), nil
}
