package configurefs

import (
	"bazil.org/fuse/fs"
	"go.uber.org/zap"
)

// FS ist the Filesystem
type FS struct {
	Logger            *zap.Logger
	MountDirectory    string
	TemplateDirectory string
	VariablesFilename string
}

// Root return the file system root directory
func (f FS) Root() (fs.Node, error) {
	return NewDirectory(f.Logger, f.MountDirectory, f.TemplateDirectory, f.VariablesFilename), nil
}
