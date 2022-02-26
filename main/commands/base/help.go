package base

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

func Help(w io.Writer, args []string) {
	cmd := RootCommand
Args:
	for i, arg := range args {
		for _, sub := range cmd.Commands {
			if sub.Name() == arg {
				cmd = sub
				continue Args
			}
		}
		helpSuccess := CommandEnv.Exec + " help"
		if i > 0 {
			helpSuccess += " " + strings.Join(args[:i], " ")
		}
		fmt.Fprintf(os.Stderr, "%s help %s: unknown help topic. Run '%s'.\n", CommandEnv.Exec, strings.Join(args, " "), helpSuccess)
		SetExitStatus(2)
		Exit()
	}
	if len(cmd.Commands) > 0 {
		PrintUsage(os.Stdout, cmd)
	} else {
		buildCommandText(cmd)
		tmpl(os.Stdout, helpTemplate, makeTmplData(cmd))
	}
}

var usageTemplate = `{{.Long | trim}}

Usage:

	{{.UsageLine}} <command> [arguments]

The commands are:
{{range .Commands}}{{if and (ne .Short "") (or (.Runnable) .Commands)}}
	{{.Name | width $.CommandsWidth}} {{.Short}}{{end}}{{end}}

Use "{{.Exec}} help{{with .LongName}} {{.}}{{end}} <command>" for more information about a command.
{{if eq (.UsageLine) (.Exec)}}
Additional help topics:
{{range .Commands}}{{if and (not .Runnable) (not .Commands)}}
	{{.Name | width $.CommandsWidth}} {{.Short}}{{end}}{{end}}

Use "{{.Exec}} help{{with .LongName}} {{.}}{{end}} <topic>" for more information about that topic.
{{end}}
`

var helpTemplate = `{{if .Runnable}}usage: {{.UsageLine}}

{{end}}{{.Long | trim}}
`

type errWriter struct {
	w   io.Writer
	err error
}

func (w *errWriter) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	if err != nil {
		w.err = err
	}
	return n, err
}

func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace, "capitalize": capitalize, "width": width})
	template.Must(t.Parse(text))
	ew := &errWriter{w: w}
	err := t.Execute(ew, data)
	if ew.err != nil {
		if strings.Contains(ew.err.Error(), "pipe") {
			SetExitStatus(1)
			Exit()
		}
		Fatalf("writing output: %v", ew.err)
	}
	if err != nil {
		panic(err)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[n:]
}

func width(width int, value string) string {
	format := fmt.Sprintf("%%-%ds", width)
	return fmt.Sprintf(format, value)
}

func PrintUsage(w io.Writer, cmd *Command) {
	buildCommandText(cmd)
	bw := bufio.NewWriter(w)
	tmpl(bw, usageTemplate, makeTmplData(cmd))
	bw.Flush()
}

func buildCommandText(cmd *Command) {
	data := makeTmplData(cmd)
	cmd.UsageLine = buildText(cmd.UsageLine, data)
	cmd.Long = buildText(cmd.Long, data)
}

func buildText(text string, data interface{}) string {
	buf := bytes.NewBuffer([]byte{})
	text = strings.ReplaceAll(text, "\t", "    ")
	tmpl(buf, text, data)
	return buf.String()
}

type tmplData struct {
	*Command
	*CommandEnvHolder
}

func makeTmplData(cmd *Command) tmplData {
	width := 12
	for _, c := range cmd.Commands {
		l := len(c.Name())
		if width < l {
			width = l
		}
	}
	CommandEnv.CommandsWidth = width
	return tmplData{
		Command:          cmd,
		CommandEnvHolder: &CommandEnv,
	}
}
