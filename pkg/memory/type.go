package memory

import "errors"

// MemoryReader defines operations for retrieving process memory data.
type MemoryReader interface {
	// ReadSMaps returns the parsed /proc/<pid>/smaps entries.
	ReadSMaps(pid int32) ([]ProcessMemorySegment, error)

	// ReadPageMap returns page-level metadata for a virtual address range.
	ReadPageMap(pid int32, virtAddr, length uint64) ([]PageInfo, error)

	// ReadNUMAMaps returns NUMA node allocation info.
	ReadNUMAMaps(pid int32) ([]NUMASegment, error)
}

// ProcessMemorySegment mirrors a single smaps entry.
type ProcessMemorySegment struct {
	StartAddr   uint64
	EndAddr     uint64
	Permissions string
	Offset      uint64
	Device      string
	Inode       uint64
	Path        string

	Size         uint64 // Kilobytes
	RSS          uint64 // Kilobytes
	PSS          uint64 // Kilobytes
	SharedClean  uint64
	SharedDirty  uint64
	PrivateClean uint64
	PrivateDirty uint64
	Referenced   uint64
	Anonymous    uint64
	Swap         uint64

	SegmentType SegmentType
}

// SegmentType categorizes a mapping.
type SegmentType string

const (
	SegmentTypeHeap    SegmentType = "heap"
	SegmentTypeStack   SegmentType = "stack"
	SegmentTypeMMap    SegmentType = "mmap"
	SegmentTypeVDSO    SegmentType = "vdso"
	SegmentTypeVVar    SegmentType = "vvar"
	SegmentTypeAnon    SegmentType = "anonymous"
	SegmentTypeUnknown SegmentType = "unknown"
)

// PageInfo represents a single pagemap entry.
type PageInfo struct {
	VirtualAddr uint64
	PhysicalPFN uint64
	Present     bool
	Swapped     bool
	FileBacked  bool
	Flags       uint64
}

// NUMASegment captures NUMA allocations.
type NUMASegment struct {
	AddressRange string
	NodePages    map[int]int64
}

var ErrUnsupported = errors.New("memory reader not supported on this platform")
