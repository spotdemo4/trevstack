package apps

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/spotdemo4/trevstack/cli/internal/utils"
)

type Node struct {
	App App
	c   chan Msg
	ctx context.Context
	wg  *sync.WaitGroup

	dir string

	Cancel context.CancelFunc
	Wait   func()
}

func NewNode(dir string, c chan Msg) *Node {

	// Create new context
	ctx, cancel := context.WithCancel(context.Background())

	// Create wait group
	wg := sync.WaitGroup{}

	node := Node{
		App: App{
			Name:  "node",
			Color: "#fab387",
		},
		c:   c,
		ctx: ctx,
		wg:  &wg,

		dir: dir,

		Cancel: cancel,
		Wait:   wg.Wait,
	}

	// Start watching
	go node.dev()

	return &node
}

func (n *Node) msg(m Msg) {
	m.Time = time.Now()
	m.App = &n.App
	n.c <- m
}

func (n *Node) dev() {
	n.wg.Add(1)
	defer n.wg.Done()

	// Create cmd
	cmd := exec.Command("npm", "run", "dev")
	cmd.Dir = n.dir

	// Stop cmd on exit
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		<-n.ctx.Done()

		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			cmd.Process.Kill() // If the process is not responding to the interrupt signal, kill it
		}
	}()

	// Start cmd
	out, err := utils.Run(cmd)
	if err != nil {
		n.msg(Msg{
			Text:    err.Error(),
			Success: utils.BoolPointer(false),
		})
		return
	}

	// Watch for output
	for line := range out {
		switch line := line.(type) {
		case utils.Stdout:
			n.msg(Msg{
				Text: string(line),
			})

		case utils.Stderr:
			n.msg(Msg{
				Text:    string(line),
				Success: utils.BoolPointer(false),
			})

		case utils.ExitCode:
			if line == 0 {
				n.msg(Msg{
					Text:    "Node stopped",
					Success: utils.BoolPointer(true),
				})
			} else {
				n.msg(Msg{
					Text:    fmt.Sprintf("Node failed with exit code %d", out),
					Success: utils.BoolPointer(false),
				})
			}
		}
	}
}
