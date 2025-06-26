package analysis

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

type TmpFS struct {
	basePath string
}

func OpenTmpFs(name string) (*TmpFS, error) {
	basePath, err := os.MkdirTemp("", name)
	if err != nil {
		return nil, fmt.Errorf("mkdir temp: %w", err)
	}

	return &TmpFS{
		basePath: basePath,
	}, nil
}

func (fs *TmpFS) Close() {
	os.RemoveAll(fs.basePath)
}

func (fs *TmpFS) AddNewFile(uri, content string) (string, error) {
	// FIXME: use some sort of common base path so that each directory becomes a real directory in the shadow filesystem

	fullpath := fs.getFilepath(uri)
	if err := os.WriteFile(fullpath, []byte(content), 0o664); err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}

	return fullpath, nil
}

func (fs *TmpFS) OpenFile(uri string) (*os.File, error) {
	return fs.OpenFileFromName(fs.getFilename(uri))
}

func (fs *TmpFS) OpenFileFromName(filename string) (*os.File, error) {
	return fs.OpenFileFromPath(fs.joinPath(filename))
}

func (fs *TmpFS) OpenFileFromPath(fullpath string) (*os.File, error) {
	return os.Open(fullpath)
}

func (fs *TmpFS) getFilepath(uri string) string {
	return fs.joinPath(fs.getFilename(uri))
}

func (fs *TmpFS) getFilename(uri string) string {
	return base64.StdEncoding.EncodeToString([]byte(uri))
}

func (fs *TmpFS) joinPath(filename string) string {
	return filepath.Join(fs.basePath, filename)
}
