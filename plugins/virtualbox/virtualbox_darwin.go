// +build darwin

package main

func getSharedFolders() map[string]string {
	return map[string]string{"Users": "/Users"}
}
