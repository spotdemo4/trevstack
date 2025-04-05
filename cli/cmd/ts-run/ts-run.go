package main

import (
	"fmt"
	"log"
	"os"

	"github.com/boyter/gocodewalker"
)

type env struct {
	DBType string
	DBUser string
	DBPass string
	DBHost string
	DBPort string
	DBName string

	RootDir  string
	NodeDir  string
	ProtoDir string
}

func main() {
	// Get pwd
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Current path: %s\n", path)

	findApps(path)
	return

	// c := make(chan apps.Msg, 10)

	// // Create protobuf watcher
	// proto, err := apps.NewProto(env.ProtoDir, env.RootDir, c)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Create node watcher
	// node := apps.NewNode(env.NodeDir, c)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// apps := []*apps.App{
	// 	&proto.App,
	// 	&node.App,
	// }

	// // Start tea
	// p := tea.NewProgram(
	// 	models.NewRunner(c, apps),
	// 	tea.WithAltScreen(),
	// 	tea.WithMouseCellMotion(),
	// )
	// if _, err := p.Run(); err != nil {
	// 	fmt.Printf("Alas, there's been an error: %v", err)
	// }

	// // Cancel watchers
	// proto.Cancel()
	// proto.Wait()

	// node.Cancel()
	// node.Wait()

	// close(c)
}

func findApps(path string) {
	fileListQueue := make(chan *gocodewalker.File, 100)
	fileWalker := gocodewalker.NewFileWalker(path, fileListQueue)

	errorHandler := func(e error) bool {
		fmt.Println("ERR", e.Error())
		return true
	}
	fileWalker.SetErrorHandler(errorHandler)

	go fileWalker.Start()

	for f := range fileListQueue {
		fmt.Printf("%s, %s\n", f.Filename, f.Location)
	}
}
