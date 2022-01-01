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
    if strings.HasPrefix(str, pf) { str = strings.TrimPrefix(str, "\"") }
    if strings.HasSuffix(str, sf) { str = strings.TrimSuffix(str, "\"") }
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
            str := removePreSuffix(argList[i+1], "\"", "\"")
            flags.PFlag = &str
        case "-I":
            if i+1 >= len(argList) {
                return nil, errors.New("tree: expected [pattern] after -I flag")
            }
            str := removePreSuffix(argList[i+1], "\"", "\"")
            flags.IFlag = &str
        }
    }

    return flags, nil
}

type ArgumentList struct {
    aFlag           bool    // All files are printed.
    LFlag           *int    // Max display depth of the directory tree.
    dFlag           bool    // List directories only.
    sFlag           bool    // Print the size of each file in bytes along with the name.
    pFlag           bool    // Print the file type and permissions for each file (as per ls -l).
    IFlag           *string  // Do not list those files that match the wild-card pattern.
    PFlag           *string  // List only those files that match the wild-card patttern.
}

func processCommand(files []File, args ArgumentList) {
    dirCnt, fileCnt := 0, 0

    bFmt := color.New(color.FgBlue)

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
            if m {
                continue
            }
        }

        if args.PFlag != nil {
            m, _ := regexp.Match(*args.PFlag, []byte(f.name))
            if !m {
                continue
            }
        }

        if args.PFlag != nil && true {
            continue
        }

        if f.isDir {
            dirCnt += 1
        } else {
            fileCnt += 1
        }

        indent := ""
        for i:=0; i<f.level; i++ { indent += "  " }

        template := "%s-- %s"
        var v []interface{}
        v = append(v, indent, f.name)

        if args.sFlag {
            template += " [%d]"
            v = append(v, f.size)
        }

        if args.pFlag {
            template += " [%s]"
            v = append(v, f.perm)
        }

        if f.isDir {
            bFmt.Printf(template + "\n", v...)
        } else {
            fmt.Printf(template + "\n", v...)
        }
    }

    printStats(dirCnt, fileCnt)
}

func printStats(dirCount int, fileCount int) {
    msg := ""
    if dirCount == 1 {
        msg += fmt.Sprint("1 directory, ")
    } else {
        msg += fmt.Sprintf("%d directories, ", dirCount)
    }
    if fileCount == 1 {
        msg += fmt.Sprint("1 file")
    } else {
        msg += fmt.Sprintf("%d files", dirCount)
    }

    fmt.Print("\n\n", msg)
}

type File struct {
    level   int
    name    string
    isDir   bool
    size    int64
    perm    string
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
        }
        *fnames = append(*fnames, file)
    }

    return nil
}
