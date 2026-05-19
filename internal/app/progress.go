package app

import (
	"fmt"
	"os"
	"sync"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type workerProgress interface {
	StartTask(workerID int, dbName, collection string, total int64)
	Advance(workerID int, n int64)
	CompleteTask(workerID int)
	Wait()
}

func newWorkerProgress(enabled bool, workerCount int) workerProgress {
	if !enabled {
		return noopProgress{}
	}

	p := mpb.New(mpb.WithOutput(os.Stdout), mpb.WithWidth(40))
	barStyle := mpb.BarStyle().
		Lbound("▕").
		Filler("█").
		Tip("▌", "▊", "█").
		Padding("░").
		Rbound("▏").
		FillerMeta(colorGreen).
		TipMeta(colorGreen).
		PaddingMeta(colorDim)
	workers := make([]workerProgressState, workerCount)
	for i := 0; i < workerCount; i++ {
		workers[i] = workerProgressState{
			label: formatIdleWorkerLabel(i),
		}
		state := &workers[i]
		bar := p.New(0, barStyle,
			mpb.PrependDecorators(
				decor.Any(func(_ decor.Statistics) string {
					state.mu.Lock()
					label := state.label
					state.mu.Unlock()
					return fmt.Sprintf("%s%-36s%s |", ansiBrightWhite, label, ansiReset)
				}, decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				decor.Any(func(_ decor.Statistics) string {
					state.mu.Lock()
					copied := state.copied
					total := state.total
					state.mu.Unlock()

					percent := int64(0)
					if total > 0 {
						percent = copied * 100 / total
					}
					return fmt.Sprintf("%s%3d%% (%6d/%6d)%s", ansiDim, percent, copied, total, ansiReset)
				}, decor.WCSyncSpaceR),
			),
		)
		workers[i].bar = bar
		workers[i].bar.SetCurrent(0)
		workers[i].bar.SetTotal(1, false)
	}

	return &mpbWorkerProgress{progress: p, workers: workers}
}

type noopProgress struct{}

func (noopProgress) StartTask(workerID int, dbName, collection string, total int64) {}
func (noopProgress) Advance(workerID int, n int64)                                  {}
func (noopProgress) CompleteTask(workerID int)                                      {}
func (noopProgress) Wait()                                                          {}

type mpbWorkerProgress struct {
	progress *mpb.Progress
	workers  []workerProgressState
}

type workerProgressState struct {
	mu     sync.Mutex
	label  string
	copied int64
	total  int64
	bar    *mpb.Bar
}

func (p *mpbWorkerProgress) StartTask(workerID int, dbName, collection string, total int64) {
	if workerID < 0 || workerID >= len(p.workers) {
		return
	}

	worker := &p.workers[workerID]
	worker.mu.Lock()
	worker.label = formatTaskWorkerLabel(workerID, dbName, collection)
	worker.copied = 0
	worker.total = total
	worker.mu.Unlock()

	taskTotal := total
	if taskTotal <= 0 {
		taskTotal = 1
	}

	worker.bar.SetCurrent(0)
	worker.bar.SetTotal(taskTotal, false)
}

func (p *mpbWorkerProgress) Advance(workerID int, n int64) {
	if workerID < 0 || workerID >= len(p.workers) || n <= 0 {
		return
	}

	worker := &p.workers[workerID]
	worker.mu.Lock()
	worker.copied += n
	if worker.total > 0 && worker.copied > worker.total {
		worker.copied = worker.total
	}
	worker.mu.Unlock()
	worker.bar.IncrInt64(n)
}

func (p *mpbWorkerProgress) CompleteTask(workerID int) {
	if workerID < 0 || workerID >= len(p.workers) {
		return
	}

	worker := &p.workers[workerID]
	worker.mu.Lock()
	worker.label = formatIdleWorkerLabel(workerID)
	worker.copied = 0
	worker.total = 0
	worker.mu.Unlock()

	worker.bar.SetCurrent(0)
	worker.bar.SetTotal(1, false)
}

func (p *mpbWorkerProgress) Wait() {
	for i := range p.workers {
		p.workers[i].bar.SetTotal(1, true)
	}
	p.progress.Wait()
}

func formatTaskWorkerLabel(workerID int, dbName, collection string) string {
	return fmt.Sprintf("%d: %s.%s", workerID, dbName, collection)
}

func formatIdleWorkerLabel(workerID int) string {
	return fmt.Sprintf("%d: idle", workerID)
}

func colorGreen(s string) string {
	return ansiGreen + s + ansiReset
}

func colorDim(s string) string {
	return ansiDim + s + ansiReset
}
