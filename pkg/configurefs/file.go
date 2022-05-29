package configurefs

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"text/template"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v3"
)

// File represents a file in the Filesystem
type File struct {
	Logger       *zap.Logger
	mountFile    string
	templateFile string
	varsFile     string
	generated    *[]byte
}

func NewFile(logger *zap.Logger, mountFile, templateFile, varsFile string) File {
	logger = logger.With(
		zap.String("mountFile", mountFile),
		zap.String("templateFile", templateFile),
	)
	logger.Debug("create file")
	return File{
		Logger:       logger,
		mountFile:    mountFile,
		templateFile: templateFile,
		varsFile:     varsFile,
	}
}

// Attr implements fs.Node for Files.
// It uses the Attributes from the template file,
// but overrides some stats to match the generated file.
func (f File) Attr(ctx context.Context, attr *fuse.Attr) error {
	f.Logger.Debug("Attr")

	if f.generated == nil {
		err := f.generate()
		if err != nil {
			return err
		}
	}

	stat, err := unixStat(f.templateFile)
	if err != nil {
		return err
	}

	err = statToAttr(stat, attr)
	if err != nil {
		return err
	}

	attr.Size = uint64(len(*f.generated))
	return nil
}

// Open implements the NodeOpener for files.
// It returns the generated file.
func (f File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	f.Logger.Sugar().Debugw("Open", "req", req)
	err := unix.Access(f.templateFile, uint32(req.Flags))
	if err != nil {
		return nil, err
	}

	if f.generated == nil {
		err := f.generate()
		if err != nil {
			return nil, err
		}
	}
	return fs.DataHandle(*f.generated), nil
}

// generate is a helper function to generate the file from the template and
// the variables.
// It stores the resulted generated file in the File object.
func (f *File) generate() error {
	f.Logger.Debug("generate")
	templateContent, err := ioutil.ReadFile(f.templateFile)
	if err != nil {
		return fmt.Errorf("failed to read template file, %w", err)
	}

	t, err := template.New(f.mountFile).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template, %w", err)
	}

	content, err := ioutil.ReadFile(f.varsFile)
	if err != nil {
		return fmt.Errorf("failed to read variable file, %w", err)
	}
	vars := make(map[string]interface{})
	err = yaml.Unmarshal(content, &vars)
	if err != nil {
		return fmt.Errorf("failed to parse variables, %w", err)
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, vars)
	if err != nil {
		return fmt.Errorf("failed to execute template, %w", err)
	}

	generated := buf.Bytes()
	f.generated = &generated
	return nil
}
