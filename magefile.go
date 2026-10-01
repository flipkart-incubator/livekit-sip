// Copyright 2023 LiveKit, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build mage
// +build mage

package main

import (
	"context"
	"errors"
	"fmt"
	"go/build"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/livekit/mageutil"
)

var Default = Build

const (
	imageName = "livekit/sip"
)

var packages = []string{"pkg-config", "opus", "opusfile"}

type packageInfo struct {
	Dir string
}

func Bootstrap() error {
	brewPrefix, err := getBrewPrefix()
	if err != nil {
		return err
	}

	for _, plugin := range packages {
		if _, err := os.Stat(fmt.Sprintf("%s%s", brewPrefix, plugin)); err != nil {
			if err = run(fmt.Sprintf("brew install %s", plugin)); err != nil {
				return err
			}
		}
	}

	return nil
}

func Build() error {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}

	return run(fmt.Sprintf("go build -a -o %s/bin/sip ./cmd/livekit-sip", gopath))
}

// runs golangci-lint
func Lint() error {
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		return errors.New("golangci-lint is not installed, install instructions: https://golangci-lint.run/docs/welcome/install/")
	}
	return mageutil.Run(context.Background(), "golangci-lint run ./...")
}

func Test() error {
	return run("go test -v ./pkg/...")
}

func Integration() error {
	return run("go test -v ./test/integration/...")
}

func BuildDocker() error {
	return mageutil.Run(context.Background(),
		fmt.Sprintf("docker build -t %s:latest -f build/sip/Dockerfile .", imageName),
	)
}

func BuildDockerLinux() error {
	return mageutil.Run(context.Background(),
		fmt.Sprintf("docker build --platform linux/amd64 -t %s:latest -f build/sip/Dockerfile .", imageName),
	)
}

func SipClient() error {
	return run("go build -C ./test/sip-client/ ./...")
}

// Deb packages an existing linux binary into dist/livekit-sip_<ver>_<arch>.deb.
// Requires nfpm (https://nfpm.goreleaser.com). Does not build the binary.
//
//	SIP_BIN=bin/livekit-sip-linux-amd64 mage Deb
//	VERSION=1.2.3 NFPM_ARCH=amd64 SIP_BIN=bin/livekit-sip-linux-amd64 mage Deb
func Deb() error {
	if _, err := exec.LookPath("nfpm"); err != nil {
		return errors.New("nfpm is not installed: go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest")
	}
	bin := os.Getenv("SIP_BIN")
	if bin == "" {
		bin = "bin/livekit-sip-linux-amd64"
	}
	if _, err := os.Stat(bin); err != nil {
		return fmt.Errorf("binary not found (%s): set SIP_BIN to a linux livekit-sip binary", bin)
	}
	version := os.Getenv("VERSION")
	if version == "" {
		version = "0.0.0-dev"
	}
	arch := os.Getenv("NFPM_ARCH")
	if arch == "" {
		arch = "amd64"
	}
	staged := filepath.Join("packaging", "livekit-sip")
	if err := copyFile(bin, staged); err != nil {
		return err
	}
	raw, err := os.ReadFile(filepath.Join("packaging", "nfpm.yaml"))
	if err != nil {
		return err
	}
	cfg := strings.NewReplacer(
		"arch: amd64", "arch: "+arch,
		"version: 0.0.0-dev", "version: "+version,
	).Replace(string(raw))
	gen := filepath.Join("packaging", "nfpm.gen.yaml")
	if err := os.WriteFile(gen, []byte(cfg), 0o644); err != nil {
		return err
	}
	if err := os.MkdirAll("dist", 0o755); err != nil {
		return err
	}
	out, err := filepath.Abs(fmt.Sprintf("dist/livekit-sip_%s_%s.deb", version, arch))
	if err != nil {
		return err
	}
	cmd := exec.Command("nfpm", "pkg", "--packager", "deb", "--config", "nfpm.gen.yaml", "--target", out)
	cmd.Dir = "packaging"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// helpers

func getBrewPrefix() (string, error) {
	out, err := exec.Command("brew", "--prefix").Output()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/Cellar/", strings.TrimSpace(string(out))), nil
}

func run(commands ...string) error {
	for _, command := range commands {
		args := strings.Split(command, " ")
		if err := runArgs(args...); err != nil {
			return err
		}
	}
	return nil
}

func runArgs(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
