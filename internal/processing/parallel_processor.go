package processing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chrlesur/json-ld-converter/internal/config"
	"github.com/chrlesur/json-ld-converter/internal/jsonld"
	"github.com/chrlesur/json-ld-converter/internal/llm"
	"github.com/chrlesur/json-ld-converter/internal/logger"
)

type ParallelProcessor struct {
	config         *config.Config
	converter      *jsonld.Converter
	llmClient      llm.Client
	workerPool     chan struct{}
	taskQueue      chan Task
	resultQueue    chan Result
	wg             sync.WaitGroup
	activeWorkers  int32
	segmentResults []SegmentResult
	resultMutex    sync.Mutex
	maxRetries     int
	retryDelay     time.Duration
	failedTasks    chan Task
}

type Task struct {
	Segment      string
	Instructions string
	Index        int
	Attempts     int
}

type Result struct {
	JSONLD string
	Error  error
	Index  int
}

type SegmentResult struct {
	Index  int
	JSONLD string
}

func NewParallelProcessor(cfg *config.Config, converter *jsonld.Converter, llmClient llm.Client) *ParallelProcessor {
	return &ParallelProcessor{
		config:      cfg,
		converter:   converter,
		llmClient:   llmClient,
		workerPool:  make(chan struct{}, cfg.Conversion.NumWorkers),
		taskQueue:   make(chan Task, cfg.Conversion.QueueSize),
		resultQueue: make(chan Result, cfg.Conversion.QueueSize),
		maxRetries:  cfg.Conversion.MaxRetries,
		retryDelay:  time.Duration(cfg.Conversion.RetryDelayMs) * time.Millisecond,
		failedTasks: make(chan Task, cfg.Conversion.QueueSize),
	}
}

func (pp *ParallelProcessor) startWorkers(ctx context.Context) {
	for i := 0; i < pp.config.Conversion.NumWorkers; i++ {
		pp.wg.Add(1)
		go func(workerID int) {
			defer pp.wg.Done()
			logger.Info(fmt.Sprintf("Worker %d started", workerID))
			for {
				select {
				case task, ok := <-pp.taskQueue:
					if !ok {
						logger.Info(fmt.Sprintf("Worker %d finished", workerID))
						return
					}
					atomic.AddInt32(&pp.activeWorkers, 1)
					result := pp.processTask(ctx, task)
					atomic.AddInt32(&pp.activeWorkers, -1)
					pp.resultQueue <- result
				case task := <-pp.failedTasks:
					atomic.AddInt32(&pp.activeWorkers, 1)
					result := pp.processTask(ctx, task)
					atomic.AddInt32(&pp.activeWorkers, -1)
					pp.resultQueue <- result
				case <-ctx.Done():
					logger.Info(fmt.Sprintf("Worker %d stopped due to context cancellation", workerID))
					return
				}
			}
		}(i)
	}
}

func (pp *ParallelProcessor) processTask(ctx context.Context, task Task) Result {
	logger.Debug(fmt.Sprintf("Processing task: %v (Attempt: %d)", task, task.Attempts+1))

	enrichedSegment, err := pp.llmClient.EnrichSegment(ctx, task.Segment, task.Instructions)
	if err != nil {
		logger.Error(fmt.Sprintf("Error enriching segment with LLM: %v", err))
		return pp.handleTaskError(task, err)
	}

	jsonLD, err := pp.converter.Convert(enrichedSegment)
	if err != nil {
		logger.Error(fmt.Sprintf("Error converting to JSON-LD: %v", err))
		return pp.handleTaskError(task, err)
	}

	return Result{JSONLD: jsonLD, Index: task.Index}
}

func (pp *ParallelProcessor) AddTask(segment string, instructions string, index int) error {
	select {
	case pp.taskQueue <- Task{Segment: segment, Instructions: instructions, Index: index}:
		logger.Debug(fmt.Sprintf("Added task to queue: %s", segment[:20]))
		return nil
	default:
		return errors.New("task queue is full")
	}
}

func (pp *ParallelProcessor) GetResults() <-chan Result {
	return pp.resultQueue
}

func (pp *ParallelProcessor) ProcessSegments(ctx context.Context, segments []string, instructions string) (string, error) {
	pp.segmentResults = make([]SegmentResult, len(segments))
	successCount := 0

	for i, segment := range segments {
		if err := pp.AddTask(segment, instructions, i); err != nil {
			return "", fmt.Errorf("failed to add task: %w", err)
		}
	}

	pp.Start(ctx)
	defer pp.Stop()

	for successCount < len(segments) {
		select {
		case result := <-pp.GetResults():
			if result.Error != nil {
				logger.Warning(fmt.Sprintf("Error processing segment %d: %v", result.Index, result.Error))
				continue
			}
			pp.addSegmentResult(result)
			successCount++
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	return pp.reconcileResults()
}

func (pp *ParallelProcessor) Start(ctx context.Context) {
	go pp.startWorkers(ctx)
}

func (pp *ParallelProcessor) Stop() {
	close(pp.taskQueue)
	pp.wg.Wait()
	close(pp.resultQueue)
}

func (pp *ParallelProcessor) addSegmentResult(result Result) {
	pp.resultMutex.Lock()
	defer pp.resultMutex.Unlock()
	pp.segmentResults[result.Index] = SegmentResult{Index: result.Index, JSONLD: result.JSONLD}
}

func (pp *ParallelProcessor) reconcileResults() (string, error) {
	sort.Slice(pp.segmentResults, func(i, j int) bool {
		return pp.segmentResults[i].Index < pp.segmentResults[j].Index
	})

	var combinedResult struct {
		Context string        `json:"@context"`
		Graph   []interface{} `json:"@graph"`
	}
	combinedResult.Context = "https://schema.org"

	for _, segmentResult := range pp.segmentResults {
		var segmentData map[string]interface{}
		if err := json.Unmarshal([]byte(segmentResult.JSONLD), &segmentData); err != nil {
			return "", fmt.Errorf("error unmarshaling segment JSON-LD: %w", err)
		}

		if graph, ok := segmentData["@graph"].([]interface{}); ok {
			combinedResult.Graph = append(combinedResult.Graph, graph...)
		}
	}

	finalJSON, err := json.MarshalIndent(combinedResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling final JSON-LD: %w", err)
	}

	return string(finalJSON), nil
}

func (pp *ParallelProcessor) handleTaskError(task Task, err error) Result {
	if task.Attempts < pp.maxRetries {
		task.Attempts++
		time.Sleep(pp.retryDelay)
		pp.failedTasks <- task
		return Result{Error: fmt.Errorf("task rescheduled for retry: %w", err), Index: task.Index}
	}
	return Result{Error: fmt.Errorf("task failed after %d attempts: %w", task.Attempts+1, err), Index: task.Index}
}
