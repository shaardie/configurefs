package configurefs

import (
	"fmt"
	"time"

	"bazil.org/fuse"
	"golang.org/x/sys/unix"
)

// statToAttr is a helper function to create fill fuse.Attr from unix stat.
func statToAttr(stat *unix.Stat_t, attr *fuse.Attr) error {
	attr.Valid = time.Duration(0)
	attr.Inode = stat.Ino
	attr.Size = uint64(stat.Size)
	attr.Blocks = uint64(stat.Blocks)
	attr.Atime = time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
	attr.Mtime = time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)
	attr.Ctime = time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)
	attr.Mode = modeToFilemode(stat.Mode)
	attr.Nlink = uint32(stat.Nlink)
	attr.Uid = stat.Uid
	attr.Gid = stat.Gid
	attr.Rdev = uint32(stat.Rdev)
	attr.BlockSize = uint32(stat.Blksize)
	return nil
}

// unixStat is a helper function to make the unix.stat call a bit more golang
// conform.
func unixStat(name string) (*unix.Stat_t, error) {
	stat := &unix.Stat_t{}
	err := unix.Stat(name, stat)
	if err != nil {
		return stat, fmt.Errorf("failed to stat %v, %w", name, err)
	}
	return stat, nil
}
