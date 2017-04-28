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

func main() {
	var (
		config = flag.String("config",
			fmt.Sprintf("%s%s.hgpusher.toml", os.Getenv("HOME"), string(os.PathSeparator)),
			"config path",
		)
	)

	flag.Parse()

	conf, err := NewConfigFromFile(*config)
	if err != nil {
		log.Fatalf("Error while loading config: %v", err)
	}

	root := flag.Arg(0)
	if root == "" {
		root = conf.Root
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
				log.Fatalf("%s -> %v", folder, err)
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
				err = client.Commit(
					[]hg.HgOption{hg.Message("commit changes"),
						hg.User(conf.Username)}, []string{""},
				)
				if err == nil {
					fmt.Printf("Commited %s\n", folder)
				} else {
					fmt.Printf("Commit failed %s\n", err.Error())
				}
			}
			// check output changes
			hgcmd := []string{
				"out",
				"--config", "auth.x.prefix=*",
				"--config", fmt.Sprintf("auth.x.username=%s", conf.Login),
				"--config", fmt.Sprintf("auth.x.password=%s", conf.Password),
			}
			out, err := client.ExecCmd(hgcmd)
			// муть какая-то если нечего пушить, то
			if err == nil {
				// output exists try to push
				if out != nil {
					fmt.Printf("Output changeset exists %s\n", folder)
					hgcmd = []string{
						"push",
						"--config", "auth.x.prefix=*",
						"--config", fmt.Sprintf("auth.x.username=%s", conf.Login),
						"--config", fmt.Sprintf("auth.x.password=%s", conf.Password),
					}
					_, err = client.ExecCmd(hgcmd)
					if err == nil {
						fmt.Printf("Push done %s\n", folder)
					} else {
						fmt.Printf("Push failed %s\n", err.Error())
					}
				} else {
					fmt.Printf("Nothing to push %s -> %s \n", folder, out)
				}
			}
		}(f)
	}
	wg.Wait()
}
