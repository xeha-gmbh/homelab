package shared

import (
	"bytes"
	"encoding/json"
	"github.com/spf13/cobra"
	"io"
	"os"
	"text/template"
)

func WithConfig(cmd *cobra.Command, opt *ExtraArgs) *printMessage {
	return &printMessage{cmd: cmd,opt:opt}
}

type printMessage struct {
	cmd 	*cobra.Command
	opt 	*ExtraArgs
}

func (p *printMessage) Info(templateText string, args map[string]interface{}) {
	p.print(p.cmd.OutOrStdout(), "INFO", templateText, args)
}

func (p *printMessage) Debug(templateText string, args map[string]interface{}) {
	if p.opt.Debug {
		p.print(p.cmd.OutOrStdout(), "DEBUG", templateText, args)
	}
}

func (p *printMessage) Error(exitCode int, templateText string, args map[string]interface{}) {
	p.print(p.cmd.OutOrStderr(), "ERROR", templateText, args)
	os.Exit(exitCode)
}

func (p *printMessage) print(w io.Writer, level, templateText string, args map[string]interface{}) {
	switch p.opt.OutputFormat {
	case OutputFormatJson:
		args["level"] = level
		args["message"] = p.getMessage(templateText, args)
		json.NewEncoder(w).Encode(args)
	default:
		w.Write([]byte("[" + level + "] " + p.getMessage(templateText, args) + "\n"))
	}
}

func (p *printMessage) getMessage(templateText string, args map[string]interface{}) string {
	buf := new(bytes.Buffer)
	t, err := template.New("template").Parse(templateText)
	if err != nil {
		panic(err)
	}
	err = t.Execute(buf, args)
	if err != nil {
		panic(err)
	}
	return buf.String()
}