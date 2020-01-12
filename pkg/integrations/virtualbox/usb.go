package virtualbox

import (
	"strconv"
)

type USBFilter struct {
	VMName    string
	Index     int
	Name      string
	VendorID  string
	ProductID string
}

func (filter *USBFilter) Create() error {
	_, err := vbm(
		"usbfilter", "add", strconv.Itoa(filter.Index),
		"--target", filter.VMName,
		"--name", filter.Name,
		"--vendorid", filter.VendorID,
		"--productid", filter.ProductID,
	)
	return err
}

func (filter *USBFilter) Remove() error {
	_, err := vbm(
		"usbfilter", "remove", strconv.Itoa(filter.Index),
		"--target", filter.VMName,
	)
	return err
}
