package apps

import (
	"context"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spotdemo4/trevstack/cli/internal/utils"
)

type Proto struct {
	App     App
	c       chan Msg
	ctx     context.Context
	wg      *sync.WaitGroup
	watcher *fsnotify.Watcher

	dir     string
	rootDir string

	Cancel context.CancelFunc
	Wait   func()
}

func NewProto(dir string, rootDir string, c chan Msg) (*Proto, error) {

	// Create new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Add directory to watcher
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if slices.Contains(watcher.WatchList(), path) {
				return nil
			}

			err := watcher.Add(path)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Create new context
	ctx, cancel := context.WithCancel(context.Background())

	// Create wait group
	wg := sync.WaitGroup{}

	proto := Proto{
		App: App{
			Name:  "proto",
			Color: "#89dceb",
		},
		c:       c,
		ctx:     ctx,
		wg:      &wg,
		watcher: watcher,

		dir:     dir,
		rootDir: rootDir,

		Cancel: cancel,
		Wait:   wg.Wait,
	}

	// Start watching
	go proto.watch()

	return &proto, nil
}

func (p *Proto) msg(m Msg) {
	m.Time = time.Now()
	m.App = &p.App
	p.c <- m
}

func (p *Proto) watch() {
	p.wg.Add(1)
	defer p.wg.Done()
	defer p.watcher.Close()

	// Create new rate limit map
	rateLimit := make(map[string]time.Time)

	p.lint()

	p.msg(Msg{
		Text: "Watching for proto changes...",
	})

	for {
		select {
		case <-p.ctx.Done():
			return

		case event, ok := <-p.watcher.Events:
			if !ok {
				return
			}

			// Rate limit
			rl, ok := rateLimit[event.Name]
			if ok && time.Since(rl) < 1*time.Second {
				continue
			}
			rateLimit[event.Name] = time.Now()

			p.msg(Msg{
				Text: "File changed: " + strings.TrimPrefix(event.Name, p.dir),
			})

			ok, _ = p.lint()
			if !ok {
				continue
			}

			p.generate()

		case err, ok := <-p.watcher.Errors:
			if !ok {
				return
			}

			p.msg(Msg{
				Text:    err.Error(),
				Success: utils.BoolPointer(false),
			})
		}
	}
}

func (p *Proto) lint() (bool, error) {
	p.msg(Msg{
		Text:    "Linting",
		Loading: utils.BoolPointer(true),
		Key:     utils.StringPointer("lint"),
	})

	// Run buf lint
	cmd := exec.Command("buf", "lint")
	cmd.Dir = p.rootDir
	out, err := utils.Run(cmd)
	if err != nil {
		p.msg(Msg{
			Text:    err.Error(),
			Success: utils.BoolPointer(false),
		})
		return false, err
	}

	// Watch for output
	for line := range out {
		switch line := line.(type) {
		case utils.Stdout:
			p.msg(Msg{
				Text: string(line),
			})

		case utils.Stderr:
			p.msg(Msg{
				Text:    string(line),
				Success: utils.BoolPointer(false),
			})

		case utils.ExitCode:
			if line == 0 {
				p.msg(Msg{
					Text:    "Buf lint successful",
					Success: utils.BoolPointer(true),
					Loading: utils.BoolPointer(false),
					Key:     utils.StringPointer("lint"),
				})

				return true, nil
			}

			p.msg(Msg{
				Text:    fmt.Sprintf("Buf lint failed with exit code %d", out),
				Success: utils.BoolPointer(false),
				Loading: utils.BoolPointer(false),
				Key:     utils.StringPointer("lint"),
			})

			return false, fmt.Errorf("buf lint failed with exit code %d", line)
		}
	}

	return false, fmt.Errorf("buf lint failed")
}

func (p *Proto) generate() error {
	p.msg(Msg{
		Text:    "Generating proto files",
		Loading: utils.BoolPointer(true),
		Key:     utils.StringPointer("generate"),
	})

	// Run buf gen
	cmd := exec.Command("buf", "generate")
	cmd.Dir = p.rootDir
	out, err := utils.Run(cmd)
	if err != nil {
		p.msg(Msg{
			Text:    err.Error(),
			Success: utils.BoolPointer(false),
		})
		return err
	}

	// Watch for output
	for line := range out {
		switch line := line.(type) {
		case utils.Stdout:
			p.msg(Msg{
				Text: string(line),
			})

		case utils.Stderr:
			p.msg(Msg{
				Text:    string(line),
				Success: utils.BoolPointer(false),
			})

		case utils.ExitCode:
			if line == 0 {
				p.msg(Msg{
					Text:    "Buf generate successful",
					Success: utils.BoolPointer(true),
					Loading: utils.BoolPointer(false),
					Key:     utils.StringPointer("generate"),
				})

				return nil
			}

			p.msg(Msg{
				Text:    fmt.Sprintf("Buf generate failed with exit code %d", out),
				Success: utils.BoolPointer(false),
				Loading: utils.BoolPointer(false),
				Key:     utils.StringPointer("generate"),
			})

			return fmt.Errorf("generate failed with exit code %d", line)
		}
	}

	return fmt.Errorf("generate failed")
}
