// +build !windows,!darwin

package configfiles

func getSystemDirectories(_ string) ([]string, error) {
	return []string{"/etc"}, nil
}
