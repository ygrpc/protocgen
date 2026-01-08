package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestGeneratedCodeBuildsCShared(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	pkgDir := filepath.Dir(thisFile)
	repoRoot := filepath.Clean(filepath.Join(pkgDir, "../.."))

	if _, err := exec.LookPath("protoc"); err != nil {
		t.Skip("protoc not found")
	}
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found (cgo build requires a C compiler)")
	}

	cgoEnabledOut, err := exec.Command("go", "env", "CGO_ENABLED").Output()
	if err != nil {
		t.Skipf("go env CGO_ENABLED failed: %v", err)
	}
	if strings.TrimSpace(string(cgoEnabledOut)) != "1" {
		t.Skip("CGO_ENABLED != 1")
	}

	artifactsRoot := filepath.Join(repoRoot, "test", "artifacts")
	if err := os.MkdirAll(artifactsRoot, 0o755); err != nil {
		t.Fatalf("mkdir artifacts root: %v", err)
	}

	runDirName := fmt.Sprintf("run-%s-%d", time.Now().UTC().Format("20060102-150405"), os.Getpid())
	runDir := filepath.Join(artifactsRoot, runDirName)
	pluginDir := filepath.Join(runDir, "bin")
	genDir := filepath.Join(runDir, "gen")
	buildDir := filepath.Join(runDir, "build")

	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir plugin dir: %v", err)
	}
	if err := os.MkdirAll(genDir, 0o755); err != nil {
		t.Fatalf("mkdir gen dir: %v", err)
	}
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		t.Fatalf("mkdir build dir: %v", err)
	}

	pluginPath := filepath.Join(pluginDir, "protoc-gen-ygrpc-cgo")

	buildPlugin := exec.Command("go", "build", "-o", pluginPath, "./cmd/protoc-gen-ygrpc-cgo")
	buildPlugin.Dir = repoRoot
	buildPlugin.Env = os.Environ()
	if out, err := buildPlugin.CombinedOutput(); err != nil {
		t.Fatalf("build plugin failed: %v\n%s", err, out)
	}

	protoDir := filepath.Join(repoRoot, "test")
	protoFile := filepath.Join(protoDir, "test.proto")

	protocCmd := exec.Command(
		"protoc",
		"-I", protoDir,
		"--plugin=protoc-gen-ygrpc-cgo="+pluginPath,
		"--ygrpc-cgo_out="+genDir,
		protoFile,
	)
	protocCmd.Env = os.Environ()
	if out, err := protocCmd.CombinedOutput(); err != nil {
		t.Fatalf("protoc generation failed: %v\n%s", err, out)
	}

	generatedGo := filepath.Join(genDir, "ygrpc_cgo", "test.ygrpc.cgo.go")
	if _, err := os.Stat(generatedGo); err != nil {
		t.Fatalf("expected generated file missing: %v", err)
	}

	goMod := []byte("module cgogen_test\n\ngo 1.23.0\n")
	if err := os.WriteFile(filepath.Join(genDir, "go.mod"), goMod, 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	soPath := filepath.Join(buildDir, "libygrpc_test.so")
	buildShared := exec.Command("go", "build", "-buildmode=c-shared", "-o", soPath, "./ygrpc_cgo")
	buildShared.Dir = genDir
	buildShared.Env = os.Environ()
	if out, err := buildShared.CombinedOutput(); err != nil {
		t.Fatalf("buildmode=c-shared failed: %v\n%s", err, out)
	}

	headerPath := filepath.Join(buildDir, "libygrpc_test.h")
	headerBytes, err := os.ReadFile(headerPath)
	if err != nil {
		t.Fatalf("expected c-shared header missing: %v", err)
	}

	if !bytes.Contains(headerBytes, []byte("TestService_Ping")) {
		t.Fatalf("expected exported symbol not found in header: %s", headerPath)
	}
	if !bytes.Contains(headerBytes, []byte("PingResponsePtr")) {
		t.Fatalf("expected response triple not found in header: %s", headerPath)
	}
	if bytes.Contains(headerBytes, []byte("MsgErrorPtr")) {
		t.Fatalf("expected msg_error triple to be absent in header: %s", headerPath)
	}
	assertHeaderLineNotContains(t, headerBytes, "TestService_Ping(", "PingRequestFree")
	assertHeaderDoesNotContain(t, headerBytes, "TestService_PingOpt1(")
	assertHeaderLineContains(t, headerBytes, "TestService_PingOpt1_TakeReq(", "PingRequestOpt1Free")
	assertHeaderLineNotContains(t, headerBytes, "TestService_PingOpt3(", "PingRequestOpt3Free")
	assertHeaderLineContains(t, headerBytes, "TestService_PingOpt3_TakeReq(", "PingRequestOpt3Free")
	if !bytes.Contains(headerBytes, []byte("Ygrpc_GetErrorMsg")) {
		t.Fatalf("expected Ygrpc_GetErrorMsg to be present in header: %s", headerPath)
	}

	pruneArtifactDirs(t, artifactsRoot, 10)
}

func assertHeaderLineContains(t *testing.T, header []byte, mustContainInLine string, expectedSubstring string) {
	t.Helper()
	line, ok := findHeaderLineContaining(header, mustContainInLine)
	if !ok {
		t.Fatalf("expected header line containing %q", mustContainInLine)
	}
	if !strings.Contains(line, expectedSubstring) {
		t.Fatalf("expected header line containing %q to also contain %q, got: %s", mustContainInLine, expectedSubstring, line)
	}
}

func assertHeaderLineNotContains(t *testing.T, header []byte, mustContainInLine string, unexpectedSubstring string) {
	t.Helper()
	line, ok := findHeaderLineContaining(header, mustContainInLine)
	if !ok {
		t.Fatalf("expected header line containing %q", mustContainInLine)
	}
	if strings.Contains(line, unexpectedSubstring) {
		t.Fatalf("expected header line containing %q to NOT contain %q, got: %s", mustContainInLine, unexpectedSubstring, line)
	}
}

func assertHeaderDoesNotContain(t *testing.T, header []byte, unexpectedSubstring string) {
	t.Helper()
	_, ok := findHeaderLineContaining(header, unexpectedSubstring)
	if ok {
		t.Fatalf("expected header to NOT contain a line with %q", unexpectedSubstring)
	}
}

func findHeaderLineContaining(header []byte, substr string) (string, bool) {
	scanner := bufio.NewScanner(bytes.NewReader(header))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, substr) {
			return line, true
		}
	}
	return "", false
}

func pruneArtifactDirs(t *testing.T, artifactsRoot string, keep int) {
	entries, err := os.ReadDir(artifactsRoot)
	if err != nil {
		t.Fatalf("read artifacts root: %v", err)
	}

	runs := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "run-") {
			continue
		}
		runs = append(runs, filepath.Join(artifactsRoot, name))
	}

	sort.Strings(runs)
	if len(runs) <= keep {
		return
	}

	toDelete := runs[:len(runs)-keep]
	for _, dir := range toDelete {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("remove old artifact dir %s: %v", dir, err)
		}
	}
}
