package oci

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/kennygrant/sanitize"
)

type OCIManifest struct {
	SchemaVersion int `json:"schemaVersion"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int    `json:"size"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int    `json:"size"`
	} `json:"layers"`
}

// Digest returns the sha256 digest of an OCI image
// as a string
func Digest(image string) (string, error) {
	options := []crane.Option{}
	return crane.Digest(image, options...)
}

// DigestPath returns the sha256 digest of an OCI image
// as a string that has been sanitized for safety as
// a file name
func DigestPath(image string) (string, error) {
	digest, err := Digest(image)
	if err != nil {
		return "", err
	}
	return sanitize.Name(digest), nil
}

// Manifest returns the container manifest of
// an OCI image
func Manifest(image string) (OCIManifest, error) {
	options := []crane.Option{}
	bb, err := crane.Manifest(image, options...)
	if err != nil {
		return OCIManifest{}, err
	}
	var mani OCIManifest
	err = json.Unmarshal(bb, &mani)
	return mani, nil
}

func Write(image, basePath string) error {
	options := []crane.Option{}
	manifest, err := Manifest(image)
	if err != nil {
		return err
	}
	// make the tars
	for n, layer := range manifest.Layers {
		src := fmt.Sprintf("%s@%s", image, layer.Digest)
		layer, err := crane.PullLayer(src, options...)
		if err != nil {
			return fmt.Errorf("pulling layer %s: %w", src, err)
		}
		blob, err := layer.Uncompressed()
		if err != nil {
			return fmt.Errorf("fetching blob %s: %w", src, err)
		}
		tarPath := filepath.Join(basePath, fmt.Sprintf("%d.tar", n))
		blobTar, err := os.Create(tarPath)
		defer blobTar.Close()
		if err != nil {
			return fmt.Errorf("creating tar : %w", err)
		}
		_, err = io.Copy(blobTar, blob)
		if err != nil {
			return fmt.Errorf("writing tar : %w", err)
		}

	}
	// extract the tars
	for n := range manifest.Layers {
		tarPath := filepath.Join(basePath, fmt.Sprintf("%d.tar", n))

		cmd := exec.Command("tar", "xf", tarPath)
		cmd.Dir = basePath // hack
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("extracting tar : %w", err)
		}
		err = os.Remove(tarPath)
		if err != nil {
			return fmt.Errorf("removing tar : %w", err)
		}
	}
	return nil
}

// Save creates a tar file in `basePath` with the
// sanitized `image` sha256 as the file name.
// This is probably not what you want to extract an
// image
func Save(image, basePath string) error {
	options := []crane.Option{}

	imageMap := map[string]v1.Image{}
	o := crane.GetOptions(options...)

	ref, err := name.ParseReference(image, o.Name...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %w", image, err)
	}

	rmt, err := remote.Get(ref, o.Remote...)
	if err != nil {
		return err
	}

	img, err := rmt.Image()
	if err != nil {
		return err
	}

	imageMap[image] = img
	dp, err := DigestPath(image)
	if err != nil {
		return err
	}
	imagePath := filepath.Join(basePath, dp)
	return crane.MultiSave(imageMap, imagePath)
}
