// +build !windows,!darwin

package main

func getSharedFolders() map[string]string {
	return map[string]string{"hosthome": "/home"}
}
