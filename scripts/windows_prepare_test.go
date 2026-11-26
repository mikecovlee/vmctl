package main

import (
	"encoding/xml"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestServerAnswerKeys(t *testing.T) {
	for _, key := range []string{"", "TVRH6-WHNXV-R9WG3-9XRFY-MY832", "invalid<&"} {
		t.Run(key, func(t *testing.T) {
			// Given a disposable preparation tree and lightweight media-tool substitutes.
			root := t.TempDir()
			for _, dir := range []string{"windows11", "windows-server2025", "bin"} {
				if err := os.Mkdir(filepath.Join(root, dir), 0700); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range []string{"windows11/prepare.sh", "windows11/bootstrap.ps1", "windows-server2025/prepare.sh", "windows-server2025/Autounattend.xml.tmpl"} {
				data, err := os.ReadFile(filepath.Join("../examples", name))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, name), data, 0700); err != nil {
					t.Fatal(err)
				}
			}
			tools := map[string]string{
				"7z":          `for arg do case "$arg" in -o*) dir=${arg#-o};; esac; done; mkdir -p "$dir/boot"; touch "$dir/boot/etfsboot.com"`,
				"openssl":     "printf '012345678901234567890123'",
				"ssh-keygen":  `while [ "$1" != -f ]; do shift; done; shift; touch "$1"; printf 'ssh-ed25519 test\n' > "$1.pub"`,
				"genisoimage": "exit 0",
			}
			for name, body := range tools {
				if err := os.WriteFile(filepath.Join(root, "bin", name), []byte("#!/bin/sh\n"+body+"\n"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			media := filepath.Join(root, "media")
			if err := os.WriteFile(media, nil, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", filepath.Join(root, "windows-server2025/prepare.sh"), media)
			cmd.Env = append(os.Environ(), "PATH="+filepath.Join(root, "bin")+":"+os.Getenv("PATH"), "OPENSSH_MSI="+media, "OPENSSH_SHA256=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "WINDOWS_IMAGE_INDEX=2", "WINDOWS_PRODUCT_KEY="+key)
			// When the actual generator renders its answer file.
			out, err := cmd.CombinedOutput()
			if strings.Contains(key, "<") {
				if err == nil {
					t.Fatalf("malformed key accepted: %s", out)
				}
				return
			}
			if err != nil {
				t.Fatalf("prepare: %v: %s", err, out)
			}
			data, err := os.ReadFile(filepath.Join(root, "windows-server2025/generated/tree/Autounattend.xml"))
			if err != nil {
				t.Fatal(err)
			}
			var answer struct {
				Settings []struct {
					Pass       string `xml:"pass,attr"`
					Components []struct {
						ProductKey string `xml:"ProductKey"`
						UserData   struct {
							ProductKey struct {
								Key string `xml:"Key"`
							} `xml:"ProductKey"`
						} `xml:"UserData"`
					} `xml:"component"`
				} `xml:"settings"`
			}
			if err := xml.Unmarshal(data, &answer); err != nil {
				t.Fatal(err)
			}
			// Then both setup phases receive the supplied key, or neither has an empty element.
			found := map[string]string{}
			for _, setting := range answer.Settings {
				for _, component := range setting.Components {
					if component.ProductKey != "" {
						found[setting.Pass] = component.ProductKey
					}
					if component.UserData.ProductKey.Key != "" {
						found[setting.Pass] = component.UserData.ProductKey.Key
					}
				}
			}
			if key == "" {
				if strings.Contains(string(data), "<ProductKey") {
					t.Fatal("empty product key emitted")
				}
			} else if found["windowsPE"] != key || found["specialize"] != key {
				t.Fatalf("phase keys: %v", found)
			}
		})
	}
}
