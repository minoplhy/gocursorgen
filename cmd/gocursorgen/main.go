package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gocursorgen/internal/cursors"
	_ "gocursorgen/internal/image_decode/formats"
	"gocursorgen/internal/theme"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	var (
		inputFile    string
		outputDir    string
		prefix       string
		themeName    string
		writeConfigs bool
	)

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-V", "--version":
			fmt.Printf("gocursorgen: Version %s\n", createVersion())
			return 0

		case "-?", "--help":
			usage(args[0])
			return 0

		case "-n", "--name":
			i++
			if i >= len(args) {
				fmt.Fprintf(os.Stderr, "%s: %s requires an argument\n", args[0], args[i-1])
				usage(args[0])
				return 1
			}
			themeName = args[i]

		case "-c", "--configs":
			writeConfigs = true

		default:
			if inputFile == "" {
				inputFile = args[i]
			} else if outputDir == "" {
				outputDir = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "%s: unexpected argument %q\n", args[0], args[i])
				usage(args[0])
				return 1
			}
		}
	}

	if inputFile == "" {
		fmt.Fprintf(os.Stderr, "%s: input YAML file is required\n", args[0])
		usage(args[0])
		return 1
	}
	if outputDir == "" {
		fmt.Fprintf(os.Stderr, "%s: output directory is required\n", args[0])
		usage(args[0])
		return 1
	}

	if themeName == "" {
		themeName = filepath.Base(outputDir)
	}

	opts := cursors.Options{
		RetainFrames: writeConfigs,
		ThemeDir:     outputDir,
	}

	tf, err := theme.ParseFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", args[0], err)
		return 1
	}

	cursorsDir := filepath.Join(outputDir, "cursors")
	if err := os.MkdirAll(cursorsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "%s: cannot create output directory %q: %v\n", args[0], cursorsDir, err)
		return 1
	}

	// cursor name -> written .cursor path
	built := map[string]string{}

	for xcSymbol, cursorName := range tf.Theme {
		outPath := filepath.Join(cursorsDir, xcSymbol+"")

		entry, err := tf.CursorByName(cursorName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", args[0], err)
			return 1
		}
		entry.Options = opts

		// Optionally write text config alongside binary
		if writeConfigs {
			configPath := filepath.Join(cursorsDir, xcSymbol+".cursor")
			fmt.Printf("  config   %s.cursor\n", xcSymbol)
			if err := entry.WriteConfig(configPath, prefix); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", args[0], err)
				return 1
			}
		}

		// Already built — symlink instead of rewriting
		/* if existing, ok := built[cursorName]; ok {
			link := outPath
			_ = os.Remove(link)
			rel, err := filepath.Rel(cursorsDir, existing)
			if err != nil {
				rel = existing
			}
			if err := os.Symlink(rel, link); err != nil {
				fmt.Fprintf(os.Stderr, "%s: symlink %q -> %q: %v\n", args[0], link, rel, err)
				return 1
			}
			fmt.Printf("  symlink  %s -> %s\n", xcSymbol, cursorName)
			continue
		}*/ // Currently disable symlinking due to problem on leveling

		// Write binary X11 Cursor
		fmt.Printf("  writing  %s (%s)\n", xcSymbol, cursorName)
		if err := entry.WriteEntry(outPath, prefix); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", args[0], err)
			return 1
		}

		built[cursorName] = outPath
	}

	if err := writeIndexTheme(outputDir, themeName); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", args[0], err)
		return 1
	}

	fmt.Printf("\nbuilt theme %q in %s\n", themeName, outputDir)
	return 0
}

func writeIndexTheme(outputDir, themeName string) error {
	path := filepath.Join(outputDir, "index.theme")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create index.theme: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f,
		"[Icon Theme]\nName=%s\nComment=%s\n",
		themeName, themeName,
	)
	return err
}

func usage(progname string) {
	fmt.Fprintf(os.Stderr,
		"Usage: %s [-V|--version] [-?|--help] [-p|--prefix dir] [-n|--name name] [-c|--configs] <input.yaml> <output_dir>\n\n"+
			"  -p, --prefix dir      prefix prepended to image file paths\n"+
			"  -n, --name name       theme name written to index.theme (default: basename of output_dir)\n"+
			"  -c, --configs         also write xcursorgen text alongside each cursor\n"+
			"  input.yaml            theme definition file\n"+
			"  output_dir            directory to write cursors/ and index.theme into\n",
		progname,
	)
}

func createVersion() string {
	out := fmt.Sprintf("%s", version)
	if commit != "none" {
		out += fmt.Sprintf("-%s", commit)
	}
	if buildDate != "unknown" {
		out += fmt.Sprintf("-%s", buildDate)
	}
	return out
}
