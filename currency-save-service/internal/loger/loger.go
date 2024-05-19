package loger

import (
	"log/slog"
	"os"
)

type Logerr struct {
	Logerr *slog.Logger
}

func NewLogerr() *Logerr {
	logerr := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return &Logerr{
		Logerr: logerr,
	}
}
