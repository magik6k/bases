package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"encoding/base64"
	"encoding/hex"
	"github.com/jbenet/go-base58"
	"github.com/whyrusleeping/base32"

	mbase "github.com/multiformats/go-multibase"
)

type hexReader struct {
	src io.Reader
}

func (r *hexReader) Read(b []byte) (int, error) {
	buf := make([]byte, len(b)*2)
	n, err := r.src.Read(buf)
	if err != nil {
		return 0, err
	}

	buf = buf[:n]
	if len(buf) > 2 && bytes.Equal([]byte{'0', 'x'}, buf[:2]) {
		buf = buf[2:]
	}
	out1 := bytes.TrimSpace(buf)
	out2 := bytes.Replace(out1, []byte{' '}, nil, -1)
	return hex.Decode(b, out2)
}

type hexWriter struct {
	dst io.Writer
}

func (w *hexWriter) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	w.dst.Write([]byte(hex.EncodeToString(b)))
	return len(b), nil
}

func (w *hexWriter) Close() error {
	return nil
}

type b58Reader struct {
	src io.Reader
}

func (r *b58Reader) Read(b []byte) (int, error) {
	buf := make([]byte, len(b)*2)
	n, err := r.src.Read(buf)
	if err != nil {
		return 0, err
	}

	buf = buf[:n]
	out := base58.Decode(string(buf))
	if len(out) == 0 {
		return 0, fmt.Errorf("invalid base58 input")
	}
	outlen := copy(b, out)
	return outlen, nil
}

type b58Writer struct {
	w io.Writer
}

func (w *b58Writer) Close() error { return nil }
func (w *b58Writer) Write(b []byte) (int, error) {
	encoded := base58.Encode(b)
	_, err := w.w.Write([]byte(encoded))
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "must pass two arguments, source encoding and target encoding")
		os.Exit(1)
	}

	var r io.Reader
	switch os.Args[1] {
	case "b64", "base64":
		r = base64.NewDecoder(base64.RawStdEncoding, os.Stdin)
	case "hex", "b16":
		r = &hexReader{os.Stdin}
	case "b32", "base32":
		r = base32.NewDecoder(base32.RawStdEncoding, os.Stdin)
	case "b58", "base58":
		r = &b58Reader{os.Stdin}
	case "raw", "bin":
		r = os.Stdin
	case "mbase":
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "reading data failed: ", err)
			os.Exit(1)
		}
		_, out, err := mbase.Decode(string(bytes.TrimSpace(data)))
		if err != nil {
			fmt.Fprintln(os.Stderr, "decoding data failed: ", err)
			os.Exit(1)
		}
		r = bytes.NewReader(out)
	default:
		fmt.Fprintln(os.Stderr, "unrecognized input encoding")
		os.Exit(1)
	}

	var w io.WriteCloser
	switch os.Args[2] {
	case "b64", "base64":
		w = base64.NewEncoder(base64.RawStdEncoding, os.Stdout)
	case "hex", "b16":
		w = &hexWriter{os.Stdout}
	case "b32", "base32":
		w = base32.NewEncoder(base32.RawStdEncoding, os.Stdout)
	case "b58", "base58":
		w = &b58Writer{os.Stdout}
	case "raw", "bin":
		w = os.Stdout
	default:
		fmt.Fprintln(os.Stderr, "unrecognized output encoding")
		os.Exit(1)
	}

	defer w.Close()
	_, err := io.Copy(w, r)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
