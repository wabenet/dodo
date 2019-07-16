package main

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/pkg/errors"
)

func CreateDiskImage(dest string, size int, in io.Reader) error {
	sizeBytes := int64(size) << 20

	vboxManage, err := exec.LookPath("VBoxManage")
	if err != nil {
		return errors.New("could not find VBoxManage command")
	}
	cmd := exec.Command(vboxManage, "convertfromraw", "stdin", dest, fmt.Sprintf("%d", sizeBytes), "--format", "VMDK")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	writtenBytes, err := io.Copy(stdin, in)
	if err != nil {
		return err
	}

	if left := sizeBytes - writtenBytes; left > 0 {
		if err := zeroFill(stdin, left); err != nil {
			return err
		}
	}

	if err := stdin.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}

func zeroFill(w io.Writer, n int64) error {
	const blocksize = 32 << 10
	zeros := make([]byte, blocksize)
	var k int
	var err error
	for n > 0 {
		if n > blocksize {
			k, err = w.Write(zeros)
		} else {
			k, err = w.Write(zeros[:n])
		}
		if err != nil {
			return err
		}
		n -= int64(k)
	}
	return nil
}
