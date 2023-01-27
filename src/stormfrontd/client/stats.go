package client

import "syscall"

// Creating structure for DiskStatus
type DiskStatus struct {
	All  uint64 `json:"All" yaml:"All"`
	Used uint64 `json:"Used" yaml:"Used"`
	Free uint64 `json:"Free" yaml:"Free"`
}

// Function to get
// disk usage of path/disk
func DiskUsage(path string) (disk DiskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}
