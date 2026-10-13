package models

import "os"

type FSNode struct {
	Path       string      `json:"path"`
	Hash       string      `json:"hash"`
	Permission os.FileMode `json:"permission"`
}

func (n *FSNode) IsModified(other *FSNode) bool {
	return n.Hash != other.Hash || n.Permission != other.Permission
}
