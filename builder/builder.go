package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Compile(code []byte) ([]byte, error) {

	buildDir, err := ioutil.TempDir("", "wrapper")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(buildDir)

	funcDir := filepath.Join(buildDir, "src", "function")

	os.MkdirAll(funcDir, 0777)

	err = ioutil.WriteFile(filepath.Join(funcDir, "function.go"), code, 0666)
	if err != nil {
		return nil, err
	}

	env := os.Environ()
	for i, v := range env {
		if strings.HasPrefix(v, "GOPATH=") {
			env[i] = v + string(os.PathListSeparator) + buildDir
			break
		}
	}

	cmd := exec.Command("go", "get", "github.com/andyleap/gofaas/wrapper")
	cmd.Dir = buildDir
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("go get errors: %s", out)
		}
		return nil, fmt.Errorf("error building go source: %v", err)
	}

	cmd = exec.Command("go", "build", "-o", filepath.Join(buildDir, "binary.out"), "-tags", "wrap", "github.com/andyleap/gofaas/wrapper")
	cmd.Dir = buildDir
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			errs := strings.Replace(string(out), filepath.Join(funcDir, "function.go"), "function.go", -1)
			return nil, fmt.Errorf("go build errors: %s", errs)
		}
		return nil, fmt.Errorf("error building go source: %v", err)
	}

	binary, err := ioutil.ReadFile(filepath.Join(buildDir, "binary.out"))
	if err != nil {
		return nil, err
	}

	return binary, nil
}
