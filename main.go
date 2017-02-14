package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	hg "bitbucket.org/gohg/gohg"
)

var (
	login = flag.String("login", os.Getenv("HG_LOGIN"), "Mercurial login")
	password = flag.String("password", os.Getenv("HG_PASSWORD"), "Mercurial password")
	username = flag.String("username", os.Getenv("HG_USERNAME"), "Mercurial username")
)

func main() {
	flag.Parse()

	root := flag.Arg(0)
	if root == "" {
		root = "."
	}

	fmt.Printf("Start from: %s\n", root)

	folders := []string{}
	// recursive walk in dirs in current path
	filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if stat, err := os.Stat(fmt.Sprintf("%s/.hg", path)); err == nil && stat.IsDir() {
			folders = append(folders, path)
		}
		return nil
	})

	var wg sync.WaitGroup

	for _, f := range folders {
		wg.Add(1)

		go func(folder string) {
			defer wg.Done()

			client := hg.NewHgClient()
			err := client.Connect("", folder, nil, false)
			if err != nil {
				log.Fatal(err)
			}
			defer client.Disconnect()

			// check chages
			status, err := client.Status([]hg.HgOption{}, []string{""})
			if string(status) == "" {
				fmt.Printf("Nothing to commit %s\n", folder)
			} else {
				// add and remove changes
				_, err = client.AddRemove([]hg.HgOption{}, []string{""})
				if err == nil {
					fmt.Printf("Added %s\n", folder)
				} else {
					fmt.Printf("Add failed %s", err.Error())
				}

				// commit changes
				err = client.Commit([]hg.HgOption{hg.Message("commit changes"), hg.User(username)}, []string{""})
				if err == nil {
					fmt.Printf("Commited %s\n", folder)
				} else {
					fmt.Printf("Commit failed %s\n", err.Error())
				}
			}
			// check output changes
			hgcmd := []string{"out", "--config", "auth.x.prefix=*", "--config", fmt.Sprintf("auth.x.username=%s", login), "--config", fmt.Sprintf("auth.x.password=%s", password)}
			out, err := client.ExecCmd(hgcmd)
			// муть какая-то если нечего пушить, то
			if err == nil {
				// output exists try to push
				if out != nil {
					fmt.Printf("Output changeset exists %s\n", folder)
					hgcmd = []string{"push", "--config", "auth.x.prefix=*", "--config", fmt.Sprintf("auth.x.username=%s", login), "--config", fmt.Sprintf("auth.x.password=%s", password)}
					_, err = client.ExecCmd(hgcmd)
					if err == nil {
						fmt.Printf("Push done %s\n", folder)
					} else {
						fmt.Printf("Push failed %s\n", err.Error())
					}
				} else {
					fmt.Printf("Nothing to push %s\n", folder)
				}
			} else {
				fmt.Printf("Nothing to push %s\n", folder)
			}
		}(f)
	}
	wg.Wait()
}
