package models

import (
	"path/filepath"
	"time"
)

const (
	ModeSetUID uint32 = 04000
	ModeSetGID uint32 = 02000
	ModeSticky uint32 = 01000
)

type FsNode struct {
	ID      int64 `json:"id"`
	AuditID int64 `json:"audit_id"`

	Path string     `json:"path"`
	Type FsNodeType `json:"type"`

	UID int `json:"uid"`
	GID int `json:"gid"`

	Owner string `json:"owner"`
	Group string `json:"group"`

	Mode uint32 `json:"mode"`

	Size int64 `json:"size"`

	Hash string `json:"hash,omitempty"`

	ModifiedAt time.Time `json:"modified_at"`

	LinkTarget string `json:"link_target,omitempty"`
}

func (f FsNode) IsFile() bool {
	return f.Type == FsFile
}

func (f FsNode) IsDirectory() bool {
	return f.Type == FsDirectory
}

func (f FsNode) IsSymlink() bool {
	return f.Type == FsSymlink
}

func (f FsNode) HasSUID() bool {
	return f.Mode&ModeSetUID != 0
}

func (f FsNode) HasSGID() bool {
	return f.Mode&ModeSetGID != 0
}

func (f FsNode) HasStickyBit() bool {
	return f.Mode&ModeSticky != 0
}

func (f FsNode) IsWorldWritable() bool {
	return f.Mode&0002 != 0
}

func (f FsNode) IsWorldReadable() bool {
	return f.Mode&0004 != 0
}

func (f FsNode) IsExecutable() bool {
	return f.Mode&0111 != 0
}

func (f FsNode) ParentPath() string {
	return filepath.Dir(f.Path)
}

func (f FsNode) Name() string {
	return filepath.Base(f.Path)
}

func (f FsNode) ContentChanged(other FsNode) bool {
	if f.Type != other.Type {
		return true
	}

	if f.Hash == "" || other.Hash == "" {
		return false
	}

	return f.Hash != other.Hash
}

func (f FsNode) PermissionsChanged(other FsNode) bool {
	return f.Mode != other.Mode ||
		f.UID != other.UID ||
		f.GID != other.GID
}
