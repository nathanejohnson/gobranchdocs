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
		gitPath         string
		pkgGoURL        string
		proxyGoURL      string
		dontOpenBrowser bool
	)
	fs := flag.NewFlagSet("gobranchdocs", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "%s [options] [path]\n", fs.Name())
		fs.PrintDefaults()
		fmt.Fprint(fs.Output(), "\nIf no path is specified, defaults to the current directory\n\n")
	}
	fs.StringVar(&pkgGoURL, "pkg-go-dev-url", "https://pkg.go.dev", "go doc url")
	fs.StringVar(&proxyGoURL, "proxy-go-url", "https://proxy.golang.org", "proxy url")
	fs.BoolVar(&dontOpenBrowser, "dont-open-browser", false, "disable opening browser url")

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
	log.Printf("head hash: %s", h)
	modName, err := getModuleNameFromGoMod(gitPath)
	if err != nil {
		log.Printf("err from go mod lookup: %s", err)
		os.Exit(1)
	}
	log.Printf("module name: %s", modName)

	u, err := generateURLFromModName(pkgGoURL, proxyGoURL, modName, h)
	if err != nil {
		log.Printf("error generating URL: %s", err)
		os.Exit(1)
	}
	log.Printf("got url: %s", u.String())
	if dontOpenBrowser {
		return
	}
	err = browser.OpenURL(u.String())
	if err != nil {
		log.Printf("error opening browser: %s", err)
		os.Exit(1)
	}
}

func getHeadSHA(gitPath string) (string, error) {
	repo, err := git.PlainOpenWithOptions(gitPath, &git.PlainOpenOptions{
		DetectDotGit:          true,
		EnableDotGitCommonDir: false,
	})
	if err != nil {
		return "", fmt.Errorf("error opening git repo: %w", err)

	}
	wt, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("error getting worktree: %w", err)
	}
	status, err := wt.Status()
	if err != nil {
		log.Printf("error getting status: %s", err)
		return "", err
	}

	log.Printf("status: %s", status)

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
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	version := v.Version
	u2, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u2.Path = path.Join(u2.Path, modName+"@"+version)
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
