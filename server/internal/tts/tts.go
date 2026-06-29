// Package tts synthesizes speech by shelling out to Piper (piper1-gpl). It is
// zero-config: it auto-discovers a voice model in the voices directory and the
// piper command. When nothing is found, Available() is false and callers fall
// back to client-side speech synthesis.
package tts

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Synthesizer runs the Piper CLI to turn Mandarin text into WAV audio.
type Synthesizer struct {
	bin   string   // executable, e.g. "piper" or "python3"
	args  []string // leading args before ours, e.g. ["-m","piper"]
	voice string   // path to the .onnx voice model
}

// New builds a synthesizer, auto-discovering anything not given explicitly:
//   - voice: explicit path if set & present, else the first *.onnx in voicesDir.
//   - binSpec "" / "auto": detect a working piper command; otherwise it's a
//     space-separated command (e.g. "python3 -m piper").
func New(binSpec, voice, voicesDir string) *Synthesizer {
	if voice == "" || !fileExists(voice) {
		voice = findVoice(voicesDir)
	}
	bin, args := resolveBin(binSpec)
	return &Synthesizer{bin: bin, args: args, voice: voice}
}

// findVoice returns the first .onnx model in dir, if any.
func findVoice(dir string) string {
	if dir == "" {
		return ""
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".onnx") {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}

// resolveBin returns the command + leading args to run Piper. An explicit spec
// is honored verbatim; "auto"/"" probes for a usable install.
func resolveBin(spec string) (string, []string) {
	if spec = strings.TrimSpace(spec); spec != "" && spec != "auto" {
		f := strings.Fields(spec)
		return f[0], f[1:]
	}
	// Prefer the console script; fall back to the python module form, verifying
	// the module is importable so Available() stays honest.
	if _, err := exec.LookPath("piper"); err == nil {
		return "piper", nil
	}
	for _, py := range []string{"python3", "python"} {
		if pythonHasPiper(py) {
			return py, []string{"-m", "piper"}
		}
	}
	return "piper", nil // not found; Available() will be false
}

func pythonHasPiper(py string) bool {
	if _, err := exec.LookPath(py); err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, py, "-c", "import piper").Run() == nil
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// Available reports whether synthesis can run (voice model + binary present).
func (s *Synthesizer) Available() bool {
	if s == nil || s.voice == "" || !fileExists(s.voice) {
		return false
	}
	_, err := exec.LookPath(s.bin)
	return err == nil
}

// Synth returns WAV audio for text. Newlines are collapsed so Piper produces a
// single utterance/file. lengthScale stretches phoneme duration: 1.0 is the
// voice's natural pace, >1 slower, <1 faster; values <=0 mean "use the default".
func (s *Synthesizer) Synth(ctx context.Context, text string, lengthScale float64) ([]byte, error) {
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return nil, fmt.Errorf("empty text")
	}

	out, err := os.CreateTemp("", "tts-*.wav")
	if err != nil {
		return nil, err
	}
	path := out.Name()
	out.Close()
	defer os.Remove(path)

	args := append(append([]string(nil), s.args...), "-m", s.voice, "-f", path)
	if lengthScale > 0 {
		args = append(args, "--length-scale", strconv.FormatFloat(lengthScale, 'f', 3, 64))
	}
	cmd := exec.CommandContext(ctx, s.bin, args...)
	cmd.Stdin = strings.NewReader(text)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("piper: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return os.ReadFile(path)
}
