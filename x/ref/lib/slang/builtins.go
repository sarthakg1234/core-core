// Copyright 2021 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slang

import (
	"fmt"
	"io"
	"os"
	"sort"

	"v.io/v23/context"
	"v.io/x/lib/textutil"
)

func (scr *Script) SetStdout(stdout io.Writer) {
	scr.stdout = stdout
}

func (scr *Script) Context() *context.T {
	return scr.ctx
}

func (scr *Script) Printf(format string, args ...interface{}) {
	fmt.Fprintf(scr.Stdout(), format, args...)
}

func (scr *Script) Help(cmd string) error {
	v, ok := supportedVerbs[cmd]
	if !ok {
		return fmt.Errorf("unrecognised function: %q", cmd)
	}
	fmt.Fprintln(scr.Stdout(), v.String())
	wr := textutil.NewUTF8WrapWriter(scr.Stdout(), 70)
	wr.SetIndents("  ", "  ")
	wr.Write([]byte(v.help))
	return wr.Flush()
}

func (scr *Script) ListFunctions(tags ...string) {
	fns := RegisteredFunctions(tags...)
	for _, fn := range fns {
		if len(tags) == 0 {
			fmt.Fprintf(scr.Stdout(), "%s: %s\n", fn.Tag, fn.Function)
			continue
		}
		fmt.Fprintf(scr.Stdout(), "%s\n", fn.Function)
	}
}

func (scr *Script) ListTags() {
	fns := RegisteredFunctions()
	tags := map[string]bool{}
	for _, fn := range fns {
		tags[fn.Tag] = true
	}
	sorted := []string{}
	for k := range tags {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	for _, s := range sorted {
		fmt.Fprintln(scr.Stdout(), s)
	}
}

func (scr *Script) ExpandEnv(s string) string {
	if scr.envvars == nil {
		return os.ExpandEnv(s)
	}
	e := os.Expand(s, func(vn string) string {
		if v, ok := scr.envvars[vn]; ok {
			return v
		}
		return "${" + vn + "}"
	})
	return os.ExpandEnv(e)
}

func (scr *Script) Stdout() io.Writer {
	return scr.stdout
}

func printf(rt Runtime, format string, args ...interface{}) error {
	rt.Printf(format, args...)
	return nil
}

func sprintf(rt Runtime, format string, args ...interface{}) (string, error) {
	return fmt.Sprintf(format, args...), nil
}

func listFunctions(rt Runtime, tags ...string) error {
	rt.ListFunctions(tags...)
	return nil
}

func listTags(rt Runtime) error {
	rt.ListTags()
	return nil
}

func expandEnv(rt Runtime, s string) (string, error) {
	return rt.ExpandEnv(s), nil
}

func help(rt Runtime, cmd string) error {
	return rt.Help(cmd)
}

func init() {
	RegisterFunction(printf, "builtin", `equivalent to fmt.Printf`, "format", "args")
	RegisterFunction(sprintf, "builtin", `equivalent to fmt.Sprintf`, "format", "args")
	RegisterFunction(listFunctions, "builtin", `list available functions`, "tags")
	RegisterFunction(listTags, "builtin", `list available tags`)
	RegisterFunction(expandEnv, "builtin", `perform shell environment variable expansion`, "value")
	RegisterFunction(help, "builtin", "display information for a specific function", "name")
}
