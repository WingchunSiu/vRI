// Command vri is the vRI v0 CLI: a local evidence store for the
// personalized-benchmark loop. See docs/CORE_DESIGN.md for the contract.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/WingchunSiu/vRI/internal/app"
	"github.com/WingchunSiu/vRI/internal/store"
)

const usage = `vri - vRI evidence store (v0)

usage: vri <command> [flags]

commands:
  init                        create .vri store in the current directory
  import <path>               import source material -> source record + object
  object put|get|list         content-addressed files and directories
  record put|get|list         typed JSON records
  ingest harbor-job <dir>     ingest a Harbor job dir -> harbor_job record
  runs <manifest-file-or-id>  reward matrix across jobs for a manifest
  review                      rebuild the review packet from the store
  release <manifest-file>     validate and release a benchmark manifest

all commands that produce structured output accept --json.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "init":
		err = cmdInit(os.Args[2:])
	case "import":
		err = cmdImport(os.Args[2:])
	case "object":
		err = cmdObject(os.Args[2:])
	case "record":
		err = cmdRecord(os.Args[2:])
	case "ingest":
		err = cmdIngest(os.Args[2:])
	case "runs":
		err = cmdRuns(os.Args[2:])
	case "review":
		err = cmdReview(os.Args[2:])
	case "release":
		err = cmdRelease(os.Args[2:])
	case "-h", "--help", "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "vri: unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "vri:", err)
		os.Exit(1)
	}
}

func cwd() (string, error) {
	return os.Getwd()
}

func openApp(jsonOut bool) (*app.App, error) {
	dir, err := cwd()
	if err != nil {
		return nil, err
	}
	return app.Open(dir, os.Stdout, jsonOut)
}

func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "structured output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir, err := cwd()
	if err != nil {
		return err
	}
	st, err := store.Init(dir)
	if err != nil {
		return err
	}
	defer st.Close()
	if *jsonOut {
		fmt.Printf(`{"store": %q}`+"\n", st.Dir)
		return nil
	}
	fmt.Printf("initialized empty vRI store in %s\n", st.Dir)
	return nil
}

func cmdImport(args []string) error {
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "structured output")
	kind := fs.String("kind", store.KindSourceExcerpt, "object kind (source_excerpt|file)")
	actor := fs.String("actor", "user", "who is writing the record")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: vri import [--kind k] <path>")
	}
	a, err := openApp(*jsonOut)
	if err != nil {
		return err
	}
	defer a.St.Close()
	_, err = a.Import(fs.Arg(0), *kind, *actor)
	return err
}

func cmdObject(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: vri object put|get|list ...")
	}
	switch args[0] {
	case "put":
		fs := flag.NewFlagSet("object put", flag.ContinueOnError)
		jsonOut := fs.Bool("json", false, "structured output")
		kind := fs.String("kind", "", "object kind (default: file, or task_package for directories)")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: vri object put [--kind k] <path>")
		}
		a, err := openApp(*jsonOut)
		if err != nil {
			return err
		}
		defer a.St.Close()
		_, err = a.ObjectPut(fs.Arg(0), *kind)
		return err
	case "get":
		fs := flag.NewFlagSet("object get", flag.ContinueOnError)
		out := fs.String("out", "", "output path (required for directory objects)")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: vri object get [--out path] <digest>")
		}
		a, err := openApp(false)
		if err != nil {
			return err
		}
		defer a.St.Close()
		return a.ObjectGet(fs.Arg(0), *out)
	case "list":
		fs := flag.NewFlagSet("object list", flag.ContinueOnError)
		jsonOut := fs.Bool("json", false, "structured output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		a, err := openApp(*jsonOut)
		if err != nil {
			return err
		}
		defer a.St.Close()
		_, err = a.ObjectList()
		return err
	default:
		return fmt.Errorf("usage: vri object put|get|list ...")
	}
}

func cmdRecord(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: vri record put|get|list ...")
	}
	switch args[0] {
	case "put":
		fs := flag.NewFlagSet("record put", flag.ContinueOnError)
		typ := fs.String("type", "", "record type (see docs/CORE_DESIGN.md)")
		file := fs.String("file", "", "read payload from file")
		data := fs.String("json", "", "payload as inline JSON")
		actor := fs.String("actor", "user", "who is writing the record")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		var payload []byte
		var err error
		switch {
		case *file != "" && *data != "":
			return fmt.Errorf("pass --file or --json, not both")
		case *file != "":
			payload, err = os.ReadFile(*file)
			if err != nil {
				return err
			}
		case *data != "":
			payload = []byte(*data)
		default:
			return fmt.Errorf("record put requires --file <path> or --json '<payload>'")
		}
		a, err := openApp(false)
		if err != nil {
			return err
		}
		defer a.St.Close()
		_, err = a.RecordPut(*typ, payload, *actor)
		return err
	case "get":
		fs := flag.NewFlagSet("record get", flag.ContinueOnError)
		jsonOut := fs.Bool("json", false, "structured output")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: vri record get <id>")
		}
		a, err := openApp(*jsonOut)
		if err != nil {
			return err
		}
		defer a.St.Close()
		_, err = a.RecordGet(fs.Arg(0))
		return err
	case "list":
		fs := flag.NewFlagSet("record list", flag.ContinueOnError)
		jsonOut := fs.Bool("json", false, "structured output")
		typ := fs.String("type", "", "filter by record type")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		a, err := openApp(*jsonOut)
		if err != nil {
			return err
		}
		defer a.St.Close()
		_, err = a.RecordList(*typ)
		return err
	default:
		return fmt.Errorf("usage: vri record put|get|list ...")
	}
}

func cmdIngest(args []string) error {
	if len(args) < 1 || args[0] != "harbor-job" {
		return fmt.Errorf("usage: vri ingest harbor-job <dir>")
	}
	fs := flag.NewFlagSet("ingest harbor-job", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "structured output")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: vri ingest harbor-job <dir>")
	}
	a, err := openApp(*jsonOut)
	if err != nil {
		return err
	}
	defer a.St.Close()
	_, err = a.IngestHarborJob(fs.Arg(0))
	return err
}

func cmdRuns(args []string) error {
	fs := flag.NewFlagSet("runs", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "structured output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: vri runs <manifest-file-or-id>")
	}
	a, err := openApp(*jsonOut)
	if err != nil {
		return err
	}
	defer a.St.Close()
	_, err = a.Runs(fs.Arg(0))
	return err
}

func cmdReview(args []string) error {
	fs := flag.NewFlagSet("review", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "structured output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	a, err := openApp(*jsonOut)
	if err != nil {
		return err
	}
	defer a.St.Close()
	_, err = a.Review()
	return err
}

func cmdRelease(args []string) error {
	fs := flag.NewFlagSet("release", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "structured output")
	reviewer := fs.String("reviewer", "user", "who approved the release")
	outcome := fs.String("outcome", "approve", "review outcome (approve|reject|more-evidence)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: vri release [--reviewer name] <manifest-file>")
	}
	a, err := openApp(*jsonOut)
	if err != nil {
		return err
	}
	defer a.St.Close()
	_, err = a.Release(fs.Arg(0), *reviewer, *outcome)
	return err
}
