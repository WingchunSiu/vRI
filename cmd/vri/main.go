// Command vri is the vRI v0 CLI: a local evidence store for the
// personalized-benchmark loop. See docs/STORAGE.md for the contract.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/WingchunSiu/vRI/internal/app"
	"github.com/WingchunSiu/vRI/internal/store"
)

const usage = `vri - vRI evidence store (v0)

usage: vri <command> [flags]

commands:
  init                        create .vri store in the current directory
  import <path>               import source material -> source record + object
  evidence capture <path>     preserve raw evidence with provenance
  context query|open|materialize
                              retrieve task-specific evidence and experience
  experience propose|revise   retain revisable, evidence-backed experience
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
	case "evidence":
		err = cmdEvidence(os.Args[2:])
	case "context":
		err = cmdContext(os.Args[2:])
	case "experience":
		err = cmdExperience(os.Args[2:])
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

type stringListFlag []string

func (s *stringListFlag) String() string { return strings.Join(*s, ",") }
func (s *stringListFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
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

func cmdEvidence(args []string) error {
	if len(args) < 1 || args[0] != "capture" {
		return fmt.Errorf("usage: vri evidence capture [flags] <path>")
	}
	fs := flag.NewFlagSet("evidence capture", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "structured output")
	title := fs.String("title", "", "human-readable source title")
	kind := fs.String("kind", "artifact", "source kind (session|run|document|artifact)")
	sourceURI := fs.String("source-uri", "", "stable source URI (defaults to file URI)")
	completeness := fs.String("completeness", app.CompletenessComplete, "capture completeness (complete|partial|unknown)")
	actor := fs.String("actor", "agent", "who is capturing the evidence")
	var scopes stringListFlag
	fs.Var(&scopes, "scope", "scope in which the evidence is relevant (repeatable)")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: vri evidence capture [flags] <path>")
	}
	a, err := openApp(*jsonOut)
	if err != nil {
		return err
	}
	defer a.St.Close()
	_, err = a.EvidenceCapture(app.EvidenceCaptureOptions{
		Path: fs.Arg(0), Title: *title, Kind: *kind, SourceURI: *sourceURI,
		Completeness: *completeness, Scopes: scopes, Actor: *actor,
	})
	return err
}

func cmdContext(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: vri context query|open|materialize ...")
	}
	switch args[0] {
	case "query":
		fs := flag.NewFlagSet("context query", flag.ContinueOnError)
		jsonOut := fs.Bool("json", false, "structured output")
		recordType := fs.String("type", "", "limit results to one record type")
		limit := fs.Int("limit", 20, "maximum results")
		var scopes stringListFlag
		fs.Var(&scopes, "scope", "require an exact record scope (repeatable)")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() == 0 {
			return fmt.Errorf("usage: vri context query [flags] <query>")
		}
		a, err := openApp(*jsonOut)
		if err != nil {
			return err
		}
		defer a.St.Close()
		_, err = a.ContextQuery(strings.Join(fs.Args(), " "), *recordType, scopes, *limit)
		return err
	case "open":
		fs := flag.NewFlagSet("context open", flag.ContinueOnError)
		jsonOut := fs.Bool("json", false, "structured output")
		locator := fs.String("path", "", "relative file inside a captured directory")
		out := fs.String("out", "", "write or extract content to this path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: vri context open [flags] <ref>")
		}
		a, err := openApp(*jsonOut)
		if err != nil {
			return err
		}
		defer a.St.Close()
		_, err = a.ContextOpen(fs.Arg(0), *locator, *out)
		return err
	case "materialize":
		fs := flag.NewFlagSet("context materialize", flag.ContinueOnError)
		jsonOut := fs.Bool("json", false, "structured output")
		out := fs.String("out", "", "new destination directory")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() == 0 || *out == "" {
			return fmt.Errorf("usage: vri context materialize --out <dir> <ref...>")
		}
		a, err := openApp(*jsonOut)
		if err != nil {
			return err
		}
		defer a.St.Close()
		_, err = a.ContextMaterialize(fs.Args(), *out)
		return err
	default:
		return fmt.Errorf("usage: vri context query|open|materialize ...")
	}
}

func cmdExperience(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: vri experience propose|revise ...")
	}
	switch args[0] {
	case "propose":
		fs := flag.NewFlagSet("experience propose", flag.ContinueOnError)
		jsonOut := fs.Bool("json", false, "structured output")
		body := fs.String("body", "", "prose or code file containing the interpretation")
		title := fs.String("title", "", "short descriptive title")
		actor := fs.String("actor", "agent", "who is proposing the experience")
		var evidence, scopes stringListFlag
		fs.Var(&evidence, "evidence", "supporting record ID or object digest (repeatable)")
		fs.Var(&scopes, "scope", "scope in which the experience may apply (repeatable)")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 || *body == "" {
			return fmt.Errorf("usage: vri experience propose --body <file> --evidence <ref> [flags]")
		}
		a, err := openApp(*jsonOut)
		if err != nil {
			return err
		}
		defer a.St.Close()
		_, err = a.ExperiencePropose(app.ExperienceOptions{
			BodyPath: *body, Title: *title, Scopes: scopes, Evidence: evidence, Actor: *actor,
		})
		return err
	case "revise":
		fs := flag.NewFlagSet("experience revise", flag.ContinueOnError)
		jsonOut := fs.Bool("json", false, "structured output")
		body := fs.String("body", "", "replacement body file (omitted to inherit)")
		title := fs.String("title", "", "replacement title (omitted to inherit)")
		status := fs.String("status", "", "proposed|active|contested|superseded|retired")
		actor := fs.String("actor", "agent", "who is revising the experience")
		var evidence, scopes stringListFlag
		fs.Var(&evidence, "evidence", "replacement evidence reference set (repeatable)")
		fs.Var(&scopes, "scope", "replacement scope set (repeatable)")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return fmt.Errorf("usage: vri experience revise [flags] <id>")
		}
		a, err := openApp(*jsonOut)
		if err != nil {
			return err
		}
		defer a.St.Close()
		_, err = a.ExperienceRevise(fs.Arg(0), app.ExperienceOptions{
			BodyPath: *body, Title: *title, Status: *status, Scopes: scopes,
			Evidence: evidence, Actor: *actor,
		})
		return err
	default:
		return fmt.Errorf("usage: vri experience propose|revise ...")
	}
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
		typ := fs.String("type", "", "record type (see docs/STORAGE.md)")
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
