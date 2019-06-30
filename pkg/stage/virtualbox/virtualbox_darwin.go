// +build darwin

package virtualbox

func getShareDriveAndName() (string, string) {
	return "Users", "/Users"
}
