package trace

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"asset-leasing-system/runtime/internal/model"
)

// Writer 负责将 Trace 写入文件系统
type Writer struct {
	tracesDir string
}

func NewWriter(tracesDir string) *Writer {
	return &Writer{tracesDir: tracesDir}
}

// Write 将 Trace 写入文件系统
func (w *Writer) Write(trace *model.Trace) (string, error) {
	// 生成 trace_id（如未设置）
	if trace.Identity.TraceID == "" {
		trace.Identity.TraceID = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%d", trace.Identity.PlanHash, time.Now().UnixNano()))))[:16]
	}

	// 构建目录路径：.traces/YYYY/MM/DD/
	now := trace.Context.Timestamp
	if now.IsZero() {
		now = time.Now()
	}
	dir := filepath.Join(w.tracesDir, now.Format("2006"), now.Format("01"), now.Format("02"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating trace directory: %w", err)
	}

	// 文件路径：trace_<plan_hash[:8]>.json
	filename := fmt.Sprintf("trace_%s.json", trace.Identity.TraceID)
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(trace, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling trace: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("writing trace: %w", err)
	}

	return path, nil
}

// Reader 负责从文件系统读取 Trace
type Reader struct {
	tracesDir string
}

func NewReader(tracesDir string) *Reader {
	return &Reader{tracesDir: tracesDir}
}

// ReadByID 按 trace_id 读取 Trace
func (r *Reader) ReadByID(traceID string) (*model.Trace, string, error) {
	// 搜索所有子目录
	var found *model.Trace
	var foundPath string
	err := filepath.Walk(r.tracesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		if filepath.Base(path) == "trace_"+traceID+".json" {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			var tr model.Trace
			if json.Unmarshal(data, &tr) == nil {
				found = &tr
				foundPath = path
			}
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("searching traces: %w", err)
	}
	if found == nil {
		return nil, "", fmt.Errorf("trace %q not found", traceID)
	}
	return found, foundPath, nil
}

// ReadLatest 读取最新的 Trace
func (r *Reader) ReadLatest() (*model.Trace, string, error) {
	var newest *model.Trace
	var newestPath string
	var newestModTime time.Time

	filepath.Walk(r.tracesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		if info.ModTime().After(newestModTime) {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			var tr model.Trace
			if json.Unmarshal(data, &tr) == nil {
				newest = &tr
				newestPath = path
				newestModTime = info.ModTime()
			}
		}
		return nil
	})

	if newest == nil {
		return nil, "", fmt.Errorf("no traces found")
	}
	return newest, newestPath, nil
}
