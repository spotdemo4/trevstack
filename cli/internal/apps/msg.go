package apps

import "time"

type Msg struct {
	Text    string
	Time    time.Time
	Key     *string
	Loading *bool
	Success *bool

	App *App
}

type App struct {
	Name    string
	Color   string
	Loading *bool
}
