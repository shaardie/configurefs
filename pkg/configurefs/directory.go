package configurefs

import (
	"context"
	"io/fs"
	"os"
	"path"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
)

// Directory represents a directory in the Filesystem
type Directory struct {
	logger      *zap.Logger
	mountDir    string
	templateDir string
	varsFile    string
}

// NewDirectory creates a new directory
func NewDirectory(logger *zap.Logger, mountDir, templateDir, varsFile string) Directory {
	logger = logger.With(
		zap.String("mountDir", mountDir),
		zap.String("templateDir", templateDir),
		zap.String("varsFile", varsFile),
	)
	logger.Debug("create directory")
	return Directory{
		logger:      logger,
		mountDir:    mountDir,
		templateDir: templateDir,
		varsFile:    varsFile,
	}
}

// Attr implements fs.Node for Directories.
// It uses the Attributes from the template directory.
func (d Directory) Attr(ctx context.Context, attr *fuse.Attr) error {
	d.logger.Debug("Attr")

	stat, err := unixStat(d.templateDir)
	if err != nil {
		return err
	}

	return statToAttr(stat, attr)
}

// Access implements fs.NodeAccesser for Directories.
// It uses the Attributes from the template directory.
func (d Directory) Access(ctx context.Context, req *fuse.AccessRequest) error {
	d.logger.Debug("Access")
	return unix.Access(d.templateDir, req.Mask)
}

// func (d Directory) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fusefs.Node, fusefs.Handle, error) {
// 	d.logger.Sugar().Errorw("Not implemented", "func", "Create")
// 	return nil, nil, syscall.EIO
// }

// Lookup implements fs.NodeStringLookuper for Directories.
// It uses the directory structure from the template directory.
func (d Directory) Lookup(ctx context.Context, name string) (fusefs.Node, error) {
	d.logger.Sugar().Debugw("Lookup", "name", name)
	fullname := path.Join(d.templateDir, name)
	stat, err := unixStat(fullname)
	if err != nil {
		return nil, err
	}

	if stat.Mode&unix.S_IFMT == syscall.S_IFDIR {
		return NewDirectory(
			d.logger,
			path.Join(d.mountDir, name),
			fullname,
			d.varsFile,
		), nil
	}
	return NewFile(
		d.logger,
		path.Join(d.mountDir, name),
		fullname,
		d.varsFile,
	), nil
}

// ReadDirAll implements fs.HandleReadDirAller for Directories.
// It uses the directory structure from the template directory.
func (d Directory) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	templateDirEntries, err := os.ReadDir(d.templateDir)
	if err != nil {
		return nil, err
	}

	dirents := make([]fuse.Dirent, 0, len(templateDirEntries))
	for _, e := range templateDirEntries {
		name := e.Name()
		stat, err := unixStat(path.Join(d.templateDir, name))
		if err != nil {
			return nil, err
		}

		dirent := fuse.Dirent{
			Inode: stat.Ino,
			Name:  name,
			Type:  modeToDirentType(stat.Mode),
		}
		dirents = append(dirents, dirent)
	}
	return dirents, nil
}

// modeToFilemode is a helper function to create os.Filemode from unix mode.
func modeToFilemode(mode uint32) os.FileMode {
	fm := fs.FileMode(mode & 0777)
	switch mode & unix.S_IFMT {
	case unix.S_IFSOCK:
		fm |= fs.ModeSocket
	case unix.S_IFLNK:
		fm |= fs.ModeSymlink
	case unix.S_IFBLK:
		fm |= fs.ModeDevice
	case unix.S_IFDIR:
		fm |= fs.ModeDir
	case unix.S_IFIFO:
		fm |= fs.ModeNamedPipe
	}
	if mode&unix.S_ISUID == unix.S_ISGID {
		fm |= fs.ModeSetuid
	}
	if mode&unix.S_ISGID == unix.S_ISGID {
		fm |= fs.ModeSetgid
	}
	if mode&unix.S_ISVTX == unix.S_ISVTX {
		fm |= fs.ModeSticky
	}
	return fm
}

// modeToDirentType is a helper function to create a fuse.DirentType from the unix
// mode.
func modeToDirentType(mode uint32) fuse.DirentType {
	switch mode & unix.S_IFMT {
	case unix.S_IFSOCK:
		return fuse.DT_Socket
	case unix.S_IFLNK:
		return fuse.DT_Link
	case unix.S_IFBLK:
		return fuse.DT_Block
	case unix.S_IFDIR:
		return fuse.DT_Dir
	case unix.S_IFIFO:
		return fuse.DT_Char
	default:
		return fuse.DT_Unknown
	}
}
