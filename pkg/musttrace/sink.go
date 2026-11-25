package musttrace

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sdcio/data-server/pkg/tree/ops/validation"
	"github.com/sdcio/yang-parser/xpath"
)

const maxSlugLen = 64

// Options configures which must-statement evaluations FileSink records and
// where trace files are written.
//
// Filters selects evaluations when empty or when any element is "*" or matches
// as a substring of Path, KeylessPath, or Statement (empty filter strings are
// ignored). FailingOnly retains failed, compile_error, runtime_error, and
// stack_underflow_ignored outcomes while discarding passed and skipped ones.
type Options struct {
	Dir         string
	Filters     []string
	FailingOnly bool
}

// Summary reports the finalized output of a FileSink after Close.
type Summary struct {
	Dir     string
	Written int
}

// IndexEntry summarizes one per-evaluation JSONL file in index.jsonl.
type IndexEntry struct {
	Seq  uint64 `json:"seq"`
	File string `json:"file"`
	validation.MustTraceMeta
	validation.MustTraceOutcome
	EventCount int `json:"event_count"`
}

type headerLine struct {
	Kind string                   `json:"kind"`
	Meta validation.MustTraceMeta `json:"meta"`
}

type endLine struct {
	Kind    string                      `json:"kind"`
	Outcome validation.MustTraceOutcome `json:"outcome"`
}

// FileSink writes one JSONL execution trace per retained must evaluation plus
// a sorted index.jsonl manifest. Directories are created lazily on write or
// Close so pre-validation failures leave disk untouched.
type FileSink struct {
	opts    Options
	nextSeq atomic.Uint64

	mu       sync.Mutex
	entries  []IndexEntry
	firstErr error
}

// NewFileSink validates Options and returns a FileSink ready for concurrent
// validation tracing without creating directories until output is flushed.
func NewFileSink(opts Options) (*FileSink, error) {
	if opts.Dir == "" {
		return nil, fmt.Errorf("trace directory must not be empty")
	}
	return &FileSink{opts: opts}, nil
}

// BeginMustTrace returns an isolated per-evaluation tracer when meta satisfies
// Options.Filters, or nil to skip tracing for that evaluation.
func (s *FileSink) BeginMustTrace(meta validation.MustTraceMeta) validation.MustEvalTracer {
	if !matchesFilters(meta, s.opts.Filters) {
		return nil
	}
	return &fileEvalTracer{
		sink: s,
		meta: meta,
	}
}

// Close creates Options.Dir if needed, writes index.jsonl ordered by sequence
// number, and returns any file error encountered during tracing.
func (s *FileSink) Close() (Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.firstErr != nil {
		return Summary{Dir: s.opts.Dir, Written: len(s.entries)}, s.firstErr
	}

	slices.SortFunc(s.entries, func(a, b IndexEntry) int {
		if a.Seq < b.Seq {
			return -1
		}
		if a.Seq > b.Seq {
			return 1
		}
		return 0
	})

	indexPath := filepath.Join(s.opts.Dir, "index.jsonl")
	if err := writeJSONLFile(indexPath, func(enc *json.Encoder) error {
		for _, entry := range s.entries {
			if err := enc.Encode(entry); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return Summary{Dir: s.opts.Dir, Written: len(s.entries)}, err
	}

	return Summary{
		Dir:     s.opts.Dir,
		Written: len(s.entries),
	}, nil
}

func (s *FileSink) recordEval(meta validation.MustTraceMeta, events []xpath.TraceEvent, outcome validation.MustTraceOutcome) {
	if s.opts.FailingOnly && !isFailingOrErrored(outcome.Status) {
		return
	}

	seq := s.nextSeq.Add(1)
	fileName := fmt.Sprintf("%04d-%s-%s.jsonl", seq, sanitizePathSlug(meta.Path), shortExprHash(meta.Statement))
	fullPath := filepath.Join(s.opts.Dir, fileName)

	err := writeJSONLFile(fullPath, func(enc *json.Encoder) error {
		if err := enc.Encode(headerLine{Kind: "must_header", Meta: meta}); err != nil {
			return err
		}
		for _, ev := range events {
			if err := enc.Encode(ev); err != nil {
				return err
			}
		}
		return enc.Encode(endLine{Kind: "must_end", Outcome: outcome})
	})

	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		if s.firstErr == nil {
			s.firstErr = err
		}
		return
	}
	s.entries = append(s.entries, IndexEntry{
		Seq:              seq,
		File:             fileName,
		MustTraceMeta:    meta,
		MustTraceOutcome: outcome,
		EventCount:       len(events),
	})
}

type fileEvalTracer struct {
	sink   *FileSink
	meta   validation.MustTraceMeta
	events []xpath.TraceEvent
}

func (t *fileEvalTracer) TraceEvent(ev xpath.TraceEvent) {
	t.events = append(t.events, ev)
}

func (t *fileEvalTracer) End(outcome validation.MustTraceOutcome) {
	t.sink.recordEval(t.meta, t.events, outcome)
}

func matchesFilters(meta validation.MustTraceMeta, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, f := range filters {
		if f == "*" {
			return true
		}
		if f == "" {
			continue
		}
		if strings.Contains(meta.Path, f) ||
			strings.Contains(meta.KeylessPath, f) ||
			strings.Contains(meta.Statement, f) {
			return true
		}
	}
	return false
}

func isFailingOrErrored(status validation.MustTraceStatus) bool {
	switch status {
	case validation.MustTraceStatusFailed,
		validation.MustTraceStatusCompileError,
		validation.MustTraceStatusRuntimeError,
		validation.MustTraceStatusStackUnderflow:
		return true
	default:
		return false
	}
}

func sanitizePathSlug(p string) string {
	trimmed := strings.Trim(p, "/")
	if trimmed == "" {
		return "root"
	}
	var b strings.Builder
	b.Grow(min(len(trimmed), maxSlugLen))
	lastUnderscore := false
	for _, r := range trimmed {
		isSafe := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '.'
		if isSafe {
			b.WriteRune(r)
			lastUnderscore = false
		} else if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
		if b.Len() >= maxSlugLen {
			break
		}
	}
	slug := strings.Trim(b.String(), "_")
	if slug == "" {
		return "entry"
	}
	return slug
}

func shortExprHash(expr string) string {
	sum := sha256.Sum256([]byte(expr))
	return hex.EncodeToString(sum[:4])
}

func writeJSONLFile(path string, writeRecords func(enc *json.Encoder) error) (retErr error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create trace directory %s: %w", filepath.Dir(path), err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); retErr == nil && closeErr != nil {
			retErr = closeErr
		}
	}()

	bw := bufio.NewWriter(f)
	enc := json.NewEncoder(bw)
	enc.SetEscapeHTML(false)
	if err := writeRecords(enc); err != nil {
		return err
	}
	return bw.Flush()
}
