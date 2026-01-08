package progress

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

import (
	"golang.org/x/sys/unix"
)

type Progressbar struct {
	currentStep int32
	totalSteps  int32
	finished    chan struct{}
	halted      chan struct{}
}

var spinnerFrames = []string{"|", "/", "-", "\\"}
var tickInterval = 500 * time.Millisecond

func New(totalSteps int) *Progressbar {
	return &Progressbar{
		totalSteps: int32(totalSteps),
		finished:   make(chan struct{}),
		halted:     make(chan struct{}),
	}
}

func (p *Progressbar) Run() func() {
	go func() {
		startTime := time.Now()
		interval := time.NewTicker(tickInterval)
		currentFrame := 0
		for {
			select {
			case <-p.finished:
				defer close(p.halted)
				fmt.Fprint(os.Stderr, "\n\r")
				return
			case <-interval.C:
				termWidth, err := fetchTerminalWidth()
				if err != nil {
					termWidth = 50
				}
				termWidth = termWidth * 2 / 3
				progress := float32(atomic.LoadInt32(&p.currentStep)) / float32(p.totalSteps)
				filledWidth := int(progress * float32(termWidth))
				fmt.Fprintf(
					os.Stderr,
					"(%s%s) [%d of %d] (~%0.0f/sec)\r",
					strings.Repeat("█", filledWidth),
					strings.Repeat("-", termWidth-filledWidth),
					atomic.LoadInt32(&p.currentStep),
					p.totalSteps,
					float32(atomic.LoadInt32(&p.currentStep))/float32(time.Since(startTime).Seconds()),
				)
				currentFrame = (currentFrame + 1) % len(spinnerFrames)
			}
		}
	}()
	return func() {
		p.finished <- struct{}{}
		<-p.halted
	}
}

func (p *Progressbar) Update(stepIncrement int32) {
	p.currentStep += stepIncrement
}

func fetchTerminalWidth() (int, error) {
	fileDesc := int(os.Stderr.Fd())
	termSize, err := unix.IoctlGetWinsize(fileDesc, unix.TIOCGWINSZ)
	if err != nil {
		return 0, err
	}
	return int(termSize.Col), nil
}
