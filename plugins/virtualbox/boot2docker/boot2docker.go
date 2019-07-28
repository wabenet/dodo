package boot2docker

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	latestReleaseURL    = "https://api.github.com/repos/boot2docker/boot2docker/releases/latest"
	releasesDownloadURL = "https://github.com/boot2docker/boot2docker/releases/download"
)

func UpdateISOCache(cacheDir string) error {
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.Mkdir(cacheDir, 0700); err != nil {
			return err
		}
	}

	isoFile := filepath.Join(cacheDir, "boot2docker.iso")
	if _, err := os.Stat(isoFile); os.IsNotExist(err) {
		log.Info("boot2bocker ISO not found locally, downloading...")
		return downloadLatestISO(cacheDir)
	}

	current, err := getVersionInfo(isoFile)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("could not read boot2docker version, downloading...")
		return downloadLatestISO(cacheDir)
	}

	latest, err := getLatestRelease()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("could not get latest boot2docker release, skipping...")
		return nil
	}

	if current != latest {
		log.Info("new boot2docker release available, downloading...")
		return downloadLatestISO(cacheDir)
	}

	return nil
}

func getVersionInfo(file string) (string, error) {
	iso, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer iso.Close()

	bytes := make([]byte, 32)
	_, err = iso.ReadAt(bytes, int64(0x8028))
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(bytes))
	index := strings.Index(version, "-v")
	if index == -1 {
		return "", errors.New("could not find version information in file")
	}

	return version[index+1:], nil
}

func getLatestRelease() (string, error) {
	resp, err := http.Get(latestReleaseURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var t struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return "", err
	}
	if t.TagName == "" {
		return "", errors.New("could not get a version tag from the Github API")
	}

	return t.TagName, nil
}

func downloadLatestISO(cachePath string) error {
	file := "boot2docker.iso"
	path := filepath.Join(cachePath, file)

	tag, err := getLatestRelease()
	if err != nil {
		return err
	}

	isoURL := fmt.Sprintf("%s/%s/%s", releasesDownloadURL, tag, file)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"url": isoURL, "file": path}).Info("downloading file")
	src, err := http.Get(isoURL)
	if err != nil {
		return err
	}
	defer src.Body.Close()

	target, err := ioutil.TempFile(cachePath, fmt.Sprintf("%s.tmp", file))
	if err != nil {
		return err
	}

	defer func() {
		logger := log.WithFields(log.Fields{"file": target.Name()})
		if err := target.Close(); err != nil {
			logger.Warn("could not close file")
		}
		if _, err := os.Stat(target.Name()); err == nil {
			if err := os.Remove(target.Name()); err != nil {
				logger.Warn("could not remove file")
			}
		}
	}()

	if _, err := io.Copy(target, src.Body); err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			log.WithFields(log.Fields{"file": path}).Warn("could not remove file")
		}
	}

	return os.Rename(target.Name(), path)
}

func MakeDiskImage(publicSSHKeyPath string) (*bytes.Buffer, error) {
	// See https://github.com/boot2docker/boot2docker/blob/master/rootfs/rootfs/etc/rc.d/automount
	magicString := "boot2docker, please format-me"

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	file := &tar.Header{Name: magicString, Size: int64(len(magicString))}

	if err := tw.WriteHeader(file); err != nil {
		return nil, err
	}

	if _, err := tw.Write([]byte(magicString)); err != nil {
		return nil, err
	}

	file = &tar.Header{Name: ".ssh", Typeflag: tar.TypeDir, Mode: 0700}
	if err := tw.WriteHeader(file); err != nil {
		return nil, err
	}

	pubKey, err := ioutil.ReadFile(publicSSHKeyPath)
	if err != nil {
		return nil, err
	}

	file = &tar.Header{Name: ".ssh/authorized_keys", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return nil, err
	}

	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return nil, err
	}

	file = &tar.Header{Name: ".ssh/authorized_keys2", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return nil, err
	}

	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}
