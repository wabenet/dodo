// +build !windows,!darwin

package main

func getShareDriveAndName() (string, string) {
	return "hosthome", "/home"
}
