package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"go.uber.org/zap"

	"github.com/shaardie/configurefs/pkg/configurefs"
)

var (
	templateDirectory string
	mountDirectory    string
	variablesFilename string
	debug             bool
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.StringVar(&templateDirectory, "template-dir", "", "Template directory containing the templates")
	flag.StringVar(&mountDirectory, "mount-dir", "", "Mount directory where the fusefs is actual mounted")
	flag.StringVar(&variablesFilename, "variable-file", "", "Variable file to fill the templates")
	flag.Parse()
}

func mainWithReturnCode() int {
	// Create logger
	loggerCfg := zap.NewProductionConfig()
	if debug {
		loggerCfg.Level.SetLevel(zap.DebugLevel)
	}
	logger, err := loggerCfg.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger, %v\n", err)
		return 1
	}
	logger = logger.With(
		zap.Any("config", map[string]interface{}{
			"bool":              debug,
			"templateDirectory": templateDirectory,
			"mountDirectory":    mountDirectory,
			"variablesFilename": variablesFilename,
		}),
	)

	// Mount the filesystem
	c, err := fuse.Mount(
		mountDirectory,
		fuse.FSName("configurefs"),
		fuse.Subtype("configurefs"),
		fuse.ReadOnly(),
	)
	if err != nil {
		logger.Sugar().Errorw(
			"Failed to mount fuse filesystem",
			"error", err,
			"mountDirectory", mountDirectory)
	}
	defer func() {
		err := c.Close()
		if err != nil {
			logger.Sugar().Errorw("failed to close fuse connection", "error", err)
			return
		}
	}()

	// Try to umount on shutdown
	s := make(chan os.Signal)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	var sError error
	go func() {
		<-s
		err := fuse.Unmount(mountDirectory)
		if err != nil {
			logger.Sugar().Errorw("failed to umount during teardown filesystem", "error", err)
			sError = err
			return
		}
	}()

	// Serve filesystem
	if err := fs.New(c, &fs.Config{
		Debug: func(msg interface{}) {
			logger.Sugar().Debugw("fuse info", "fuse msg", msg)
		},
	}).Serve(configurefs.FS{
		Logger:            logger,
		MountDirectory:    mountDirectory,
		TemplateDirectory: templateDirectory,
		VariablesFilename: variablesFilename,
	}); err != nil || sError != nil {
		logger.Sugar().Errorw("failed to stop serving filesystem", "error", err)
		return 1
	}

	logger.Info("fuse filesystem umounted and stopped")
	return 0
}

func main() {
	os.Exit(mainWithReturnCode())
}
