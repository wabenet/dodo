// +build !windows,!darwin

package virtualbox

func getShareDriveAndName() (string, string) {
	return "hosthome", "/home"
}
