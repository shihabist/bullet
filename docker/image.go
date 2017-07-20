package docker

import (
	"fmt"
	"os"
	"strings"

	"github.com/FurqanSoftware/bullet/spec"
	"github.com/FurqanSoftware/bullet/ssh"
)

type Image struct {
	Application spec.Application
	Program     spec.Program
	ID          string
	Repository  string
}

type GetImageOptions struct {
	DockerPath string
}

func GetImage(c *ssh.Client, app spec.Application, prog spec.Program, options GetImageOptions) (*Image, error) {
	name := fmt.Sprintf("%s_%s", app.Identifier, prog.Key)

	img := Image{}
	b, err := c.Output(fmt.Sprintf("%s ps -a --format '{{.ID}}\t{{.Repository}}'", options.DockerPath))
	if err != nil {
		return nil, err
	}
	s := string(b)

	lines := strings.Split(s, "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if parts[1] != name {
			continue
		}
		img = Image{
			Application: app,
			Program:     prog,
			ID:          parts[0],
			Repository:  parts[1],
		}
		break
	}
	return &img, nil
}

type BuildImageOptions struct {
	DockerPath string
}

func BuildImage(c *ssh.Client, app spec.Application, prog spec.Program, options BuildImageOptions) error {
	f, err := os.Open(prog.Container.Dockerfile)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	appDir := fmt.Sprintf("/opt/%s", app.Identifier)

	err = c.Push(fmt.Sprintf("%s/Dockerfile.%s", appDir, prog.Key), 0644, fi.Size(), f)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%s_%s", app.Identifier, prog.Key)
	return c.Run(fmt.Sprintf("docker build -t %s -f %s/Dockerfile.%s %s", name, appDir, prog.Key, appDir))
}
