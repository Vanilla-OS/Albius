# Prometheus

Prometheus is a simple and accessible library for pulling and mounting container 
images. It is designed to be used as a dependency in [ABRoot](https://github.com/vanilla-os/abroot) 
and [Albius](https://github.com/vanilla-os/albius).

## Build dependencies

- `libbtrfs-dev`
- `libdevmapper-dev`
- `libgpgme-dev`

## Usage

```go
package main

import (
	"fmt"

	"github.com/docker/docker/pkg/reexec"
	"github.com/vanilla-os/prometheus"
)

func main() {
	if reexec.Init() { // needed for subprocesses
		return
	}

	pmt, err := prometheus.NewPrometheus("storage", "overlay")
	if err != nil {
		panic(err)
	}

	manifest, err := pmt.PullImage(
		"registry.vanillaos.org/vanillaos/desktop:main", // image name
		"vos-desktop", // stored image name
	)
	if err != nil {
		panic(err)
	}

    fmt.Printf("Image pulled with digest %s\n", manifest.Config.Digest)

	image, err := pmt.GetImageByDigest(manifest.Config.Digest)
	if err != nil {
		panic(err)
	}

	mountPoint, err := pmt.MountImage(image.TopLayer)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Image mounted at %s\n", mountPoint)

    if err := pmt.UnmountImage(mountPoint); err != nil {
        panic(err)
    }
}
```

## License

This project is based on some of the [containers](https://github.com/containers)
libraries, which are licensed under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).

Prometheus is distributed under the [GPLv3](https://www.gnu.org/licenses/gpl-3.0.en.html)
license.


## Why the name Prometheus?

Prometheus was the Titan of Greek mythology who stole fire from the gods to 
give it to humans, symbolizing the transmission of knowledge and technology. 
The Prometheus package provides a simple and accessible solution for pulling 
and mounting container images, making it easier to interact with OCI images 
in other projects.
