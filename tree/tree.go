package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
        "regexp"

	"github.com/fatih/color"
)

type ArgumentList struct {
    /* LISTING OPTIONS */
    aFlag           bool    // All files are printed.
    LFlag           *int    // Max display depth of the directory tree.
    dFlag           bool    // List directories only.
    IFlag           *string // Do not list those files that match the wild-card pattern.
    PFlag           *string // List only those files that match the wild-card patttern.
    fFlag           bool    // Prints the full path prefix for each file.

    /* FILE OPTIONS */
    sFlag           bool    // Print the size of each file in bytes along with the name.
    pFlag           bool    // Print the file type and permissions for each file (as per ls -l).
    QFlag           bool    // Quote the names of files in double quotes
}

type File struct {
    level   int
    name    string
    isDir   bool
    size    int64
    perm    string
    path    string
}

type Counters struct {
    dirCount    int
    filesCount  int
}

func main() {
    homedir, err := os.UserHomeDir()
    dn := fmt.Sprintf("%s/projects/commands/data", homedir)
    files, err := ReadDir(dn)
    if err != nil {
        fmt.Println(err)
    }

    args, err := parseArgs(os.Args)
    if err != nil {
        fmt.Println(err)
    }

    processCommand(files, *args)
}

func removePreSuffix(str string, pf string, sf string) string {
    return str
}

func parseArgs(argList []string) (*ArgumentList, error) {
    flags := &ArgumentList{}

    for i, a := range argList {
        switch a {
        case "-a":
            flags.aFlag = true
        case "-d":
            flags.dFlag = true
        case "-L":
            if i+1 >= len(argList) {
                return nil, errors.New("tree: expected [number] after -L flag")
            }
            num, e := strconv.ParseInt(argList[i+1], 10, 32)
            castNum := int(num)
            if e != nil {
                return nil, e
            }
            if castNum <= 0 {
                return nil, errors.New("tree: Invalid level, must be greater than 0.")
            }
            flags.LFlag = &castNum
        case "-s":
            flags.sFlag = true
        case "-p":
            flags.pFlag = true
        case "-P":
            if i+1 >= len(argList) {
                return nil, errors.New("tree: expected [pattern] after -P flag")
            }
            str := argList[i+1]
            if strings.HasPrefix(str, "\"") { str = strings.TrimPrefix(str, "\"") }
            if strings.HasSuffix(str, "\"") { str = strings.TrimSuffix(str, "\"") }
            flags.PFlag = &str
        case "-I":
            if i+1 >= len(argList) {
                return nil, errors.New("tree: expected [pattern] after -I flag")
            }
            str := argList[i+1]
            if strings.HasPrefix(str, "\"") { str = strings.TrimPrefix(str, "\"") }
            if strings.HasSuffix(str, "\"") { str = strings.TrimSuffix(str, "\"") }
            flags.IFlag = &str
        case "-Q":
            flags.QFlag = true
        case "-f":
            flags.fFlag = true
        }
    }

    return flags, nil
}

func processCommand(files []File, args ArgumentList) {
    counters := Counters{}

    colorFmt := color.New(color.FgBlue)

    for i := len(files)-1; i >= 0; i-- {
        f := files[i]

        if strings.HasPrefix(f.name, ".") && !args.aFlag {
            continue
        }

        if args.dFlag && !f.isDir {
            continue
        }

        if args.LFlag != nil && f.level > *args.LFlag {
            continue
        }

        if args.IFlag != nil {
            m, _ := regexp.Match(*args.IFlag, []byte(f.name))
            if m { continue }
        }

        if args.PFlag != nil {
            m, _ := regexp.Match(*args.PFlag, []byte(f.name))
            if !m { continue }
        }

        name := f.name

        if args.fFlag {
            name = f.path
        }

        if args.QFlag {
            name = fmt.Sprintf("\"%s\"", name)
        }

        if f.isDir {
            counters.dirCount += 1
        } else {
            counters.filesCount += 1
        }

        t := "%*s"
        indent := f.level

        if f.isDir {
            t = "|" + t
        }

        var v []interface{}
        v = append(v, indent + len(name), name)

        if args.sFlag {
            t += " [%d]"
            v = append(v, f.size)
        }

        if args.pFlag {
            t += " [%s]"
            v = append(v, f.perm)
        }

        var op string

        if f.isDir {
            op = colorFmt.Sprintf(t + "\n", v...)
        } else {
            op = fmt.Sprintf(t + "\n", v...)
        }

        fmt.Print(op)
    }

    printCounters(counters)
}

func printCounters(c Counters) {
    msg := ""
    if c.dirCount == 1 {
        msg += fmt.Sprint("1 directory, ")
    } else {
        msg += fmt.Sprintf("%d directories, ", c.dirCount)
    }
    if c.filesCount == 1 {
        msg += fmt.Sprint("1 file")
    } else {
        msg += fmt.Sprintf("%d files", c.filesCount)
    }

    fmt.Print("\n\n", msg)
}

func ReadDir(dirname string) ([]File, error) {
    fnames := make([]File, 0)
    err := readDir(dirname, &fnames, 1)
    return fnames, err
}

func readDir(dirname string, fnames *[]File, level int) error {
    entries, err := os.ReadDir(dirname)
    if err != nil {
        return err
    }

    for _, f := range entries {

        if f.IsDir() {
            readDir(path.Join(dirname, f.Name()), fnames, level + 1)
        }

        fi, _ := f.Info()

        file := File{
            level: level,
            name: f.Name(),
            isDir: f.IsDir(),
            size: fi.Size(),
            perm: fi.Mode().String(),
            path: path.Join(dirname, f.Name()),
        }
        *fnames = append(*fnames, file)
    }

    return nil
}
