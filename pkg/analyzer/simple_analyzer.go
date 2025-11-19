package analyzer

import (
	"fmt"
	"time"

	"go-ptop/pkg/jvm"
	"go-ptop/pkg/memory"
	"go-ptop/pkg/proc"
)

type simpleAnalyzer struct {
	memoryReader memory.MemoryReader
	jvmClient    jvm.JVMClient
	procReader   proc.ProcReader
}

// New returns a basic Analyzer backed by the provided readers/clients.
func New(memoryReader memory.MemoryReader, jvmClient jvm.JVMClient, procReader proc.ProcReader) Analyzer {
	return &simpleAnalyzer{
		memoryReader: memoryReader,
		jvmClient:    jvmClient,
		procReader:   procReader,
	}
}

func (a *simpleAnalyzer) Analyze(pid int32) (*AnalysisResult, error) {
	if a.memoryReader == nil || a.jvmClient == nil || a.procReader == nil {
		return nil, fmt.Errorf("analyzer not fully initialized")
	}

	segments, err := a.memoryReader.ReadSMaps(pid)
	if err != nil {
		return nil, fmt.Errorf("read smaps: %w", err)
	}

	if err := a.jvmClient.Connect(pid); err != nil {
		return nil, ErrAttachFailed{Reason: err}
	}
	defer a.jvmClient.Close()

	dump, err := a.jvmClient.GetThreadDump()
	if err != nil {
		return nil, ErrAttachFailed{Reason: err}
	}

	kernelThreads, err := a.procReader.Threads(pid)
	if err != nil {
		return nil, fmt.Errorf("read kernel threads: %w", err)
	}

	threadSegments, err := a.CorrelateThreads(dump.Threads, kernelThreads, segments)
	if err != nil {
		return nil, err
	}

	classified, err := a.ClassifyMemory(segments, nil)
	if err != nil {
		return nil, err
	}

	result := &AnalysisResult{
		PID:            pid,
		Timestamp:      time.Now(),
		Threads:        threadSegments,
		MemorySegments: segments,
		NMTReport:      nil,
	}

	var totalRSS, totalSize uint64
	for _, seg := range segments {
		totalRSS += seg.RSS
		totalSize += seg.Size
	}
	result.TotalRSS = totalRSS
	result.TotalVSize = totalSize

	categories := buildCategories(classified)
	result.JavaHeap = categoryOrDefault(categories, PurposeJavaHeap)
	result.Metaspace = categoryOrDefault(categories, PurposeMetaspace)
	result.CodeCache = categoryOrDefault(categories, PurposeCodeCache)
	result.ThreadStacks = categoryOrDefault(categories, PurposeThreadStack)
	result.DirectBuffers = categoryOrDefault(categories, PurposeDirectBuffer)
	result.NativeMemory = categoryOrDefault(categories, PurposeNativeLibrary)
	result.SharedLibraries = categoryOrDefault(categories, PurposeNativeLibrary)
	result.Other = categoryOrDefault(categories, PurposeUnknown)

	return result, nil
}

func (a *simpleAnalyzer) CorrelateThreads(javaThreads []jvm.JavaThread, kernelThreads []proc.KernelThread, segments []memory.ProcessMemorySegment) ([]ThreadMemorySegment, error) {
	javaIndex := make(map[int32]*jvm.JavaThread, len(javaThreads))
	for i := range javaThreads {
		jt := &javaThreads[i]
		if jt.OSID != 0 {
			javaIndex[jt.OSID] = jt
		}
	}

	results := make([]ThreadMemorySegment, 0, len(kernelThreads))
	for _, kt := range kernelThreads {
		var javaThread *jvm.JavaThread
		if jt, ok := javaIndex[kt.TID]; ok {
			javaThread = jt
		}

		seg, _ := findSegmentForAddress(segments, kt.StartStack)
		var ioStats proc.IOStats
		if a.procReader != nil {
			if stats, err := a.procReader.ThreadIO(kt.PID, kt.TID); err == nil && stats != nil {
				ioStats = *stats
			}
		}

		results = append(results, ThreadMemorySegment{
			JavaThread:   javaThread,
			KernelThread: kt,
			Segment:      seg,
			IOStats:      ioStats,
		})
	}

	return results, nil
}

func (a *simpleAnalyzer) ClassifyMemory(segments []memory.ProcessMemorySegment, report *jvm.NMTReport) ([]ClassifiedSegment, error) {
	classified := make([]ClassifiedSegment, 0, len(segments))
	for _, seg := range segments {
		purpose := inferPurpose(seg)
		classified = append(classified, ClassifiedSegment{
			Segment:    seg,
			Purpose:    purpose,
			Confidence: 0.5,
		})
	}
	return classified, nil
}

func buildCategories(classified []ClassifiedSegment) map[MemoryPurpose]MemoryCategory {
	categories := make(map[MemoryPurpose]MemoryCategory)
	for _, seg := range classified {
		cat := categories[seg.Purpose]
		if cat.Purpose == "" {
			cat.Purpose = seg.Purpose
		}
		cat.TotalSize += seg.Segment.Size
		cat.RSS += seg.Segment.RSS
		cat.Segments = append(cat.Segments, seg)
		categories[seg.Purpose] = cat
	}
	return categories
}

func categoryOrDefault(categories map[MemoryPurpose]MemoryCategory, purpose MemoryPurpose) MemoryCategory {
	if cat, ok := categories[purpose]; ok {
		return cat
	}
	return MemoryCategory{Purpose: purpose}
}

func inferPurpose(seg memory.ProcessMemorySegment) MemoryPurpose {
	switch seg.SegmentType {
	case memory.SegmentTypeHeap:
		return PurposeJavaHeap
	case memory.SegmentTypeStack:
		return PurposeThreadStack
	case memory.SegmentTypeMMap:
		return PurposeNativeLibrary
	case memory.SegmentTypeAnon:
		return PurposeDirectBuffer
	case memory.SegmentTypeVDSO, memory.SegmentTypeVVar:
		return PurposeNativeLibrary
	default:
		return PurposeUnknown
	}
}

func findSegmentForAddress(segments []memory.ProcessMemorySegment, addr uint64) (memory.ProcessMemorySegment, bool) {
	if addr == 0 {
		return memory.ProcessMemorySegment{}, false
	}
	for _, seg := range segments {
		if addr >= seg.StartAddr && addr < seg.EndAddr {
			return seg, true
		}
	}
	return memory.ProcessMemorySegment{}, false
}
