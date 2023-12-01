package progress

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// WorkerProgress represents the progress for a specific worker.
type WorkerProgress struct {
	workerID int
	total    int64
	current  float64
}

// ProgressBar represents a progress bar for multiple workers.
type ProgressBar struct {
	mu           sync.Mutex
	workerMap    map[int]*WorkerProgress
	updateDelay  time.Duration
	maxWorkerLen int
}

// NewProgressBar creates a new progress bar.
func NewProgressBar(updateDelay time.Duration) *ProgressBar {
	return &ProgressBar{
		workerMap:   make(map[int]*WorkerProgress),
		updateDelay: updateDelay,
	}
}

// StartWorker starts a worker to update the progress.
func (p *ProgressBar) StartWorker(workerID int, total int64, done <-chan struct{}, progress <-chan float64) {
	workerProgress := &WorkerProgress{
		workerID: workerID,
	}
	p.mu.Lock()
	p.workerMap[workerID] = workerProgress
	// if len(workerID) > p.maxWorkerLen {
	// 	p.maxWorkerLen = len(workerID)
	// }
	p.mu.Unlock()

	go func() {
		for {
			select {
			case <-done:
				return
			case val := <-progress:
				fmt.Println(val)
				p.updateProgress(workerID, val, total)
			}
		}
	}()
}

func (p *ProgressBar) updateProgress(workerID int, value float64, total int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	workerProgress, ok := p.workerMap[workerID]
	if !ok {
		return
	}
	if total != workerProgress.total {
		workerProgress.total = total
	}

	workerProgress.current += value
	p.printProgress()
	time.Sleep(p.updateDelay)
}

func (p *ProgressBar) printProgress() {
	// Clear the current line
	fmt.Print("\r")
	workerKeys := make([]int, 0, len(p.workerMap))
	for key := range p.workerMap {
		workerKeys = append(workerKeys, key)
	}
	// Sort the keys
	sort.Ints(workerKeys)

	for _, key := range workerKeys {
		workerProgress := p.workerMap[key]
		progress := float64(workerProgress.current) / float64(workerProgress.total) * 100
		progressStr := fmt.Sprintf("%d: %.2f%%", workerProgress.workerID, progress)
		fmt.Print(progressStr + "  ")
	}
	// Flush the output
	fmt.Print("\r")
}

// Complete marks the progress as complete for a worker.
func (p *ProgressBar) Complete(workerID int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.workerMap, workerID)
	p.printProgress()
	fmt.Printf("\n%d: Complete\n", workerID)
}

// WaitForCompletion waits for all workers to complete.
func (p *ProgressBar) WaitForCompletion() {
	for len(p.workerMap) > 0 {
		p.printProgress()
		time.Sleep(p.updateDelay)
	}

	// Move to the next line
	fmt.Println()
}
