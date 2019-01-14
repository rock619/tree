package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var (
	node          = []byte("├── ")
	lastNode      = []byte("└── ")
	trunk         = []byte("│   ")
	blank         = []byte("    ")
	arrow         = []byte(" -> ")
	newLine       = []byte("\n")
	pathSeparator = string(os.PathSeparator)
)

const (
	ExitCodeOK = iota
	ExitCodeNG
)

type Exec struct {
	out, errOut io.Writer
	level       int
	dirs, files int
}

type Tree struct {
	dir    string
	isLast []bool
}

func main() {
	w := bufio.NewWriterSize(os.Stdout, 512)
	ew := bufio.NewWriter(os.Stderr)
	exec := Exec{
		out:    w,
		errOut: ew,
	}
	ret := exec.Run(os.Args)
	w.Flush()
	ew.Flush()
	os.Exit(ret)
}

func (e *Exec) Run(args []string) int {
	flags := flag.NewFlagSet("tree", flag.ExitOnError)
	flags.SetOutput(e.errOut)
	flags.IntVar(&e.level, "L", 0, "Descend only level directories deep.")
	flags.Parse(args[1:])
	if flags.NArg() == 0 {
		e.run(".")
	} else {
		for _, r := range flags.Args() {
			e.run(r)
		}
	}
	dirExp := "directories"
	if e.dirs == 1 {
		dirExp = "directory"
	}
	fileExp := "files"
	if e.files == 1 {
		fileExp = "file"
	}
	fmt.Fprintf(e.out, "\n%d %s, %d %s\n", e.dirs, dirExp, e.files, fileExp)
	return ExitCodeOK
}

func (e *Exec) run(root string) {
	io.WriteString(e.out, root)
	e.out.Write(newLine)
	t := Tree{
		dir:    root,
		isLast: make([]bool, 0, 512),
	}
	e.work(&t)
}

func (e *Exec) work(t *Tree) {
	files, err := ioutil.ReadDir(t.dir)
	if err != nil {
		io.WriteString(e.errOut, err.Error())
	}
	t.isLast = append(t.isLast, false)
	for i, f := range files {
		if f.Name()[0] == '.' {
			continue
		}
		if i == len(files)-1 {
			t.isLast[len(t.isLast)-1] = true
		}
		e.writeLine(t, f)
		if !f.IsDir() {
			if f.Mode()&os.ModeSymlink == 0 {
				e.files++
				continue
			}
			l, err := os.Stat(t.dir + pathSeparator + f.Name())
			if err != nil {
				io.WriteString(e.errOut, err.Error())
			}
			if l.IsDir() {
				e.dirs++
			} else {
				e.files++
			}
			continue
		}
		e.dirs++
		if e.level != 0 && len(t.isLast) >= e.level {
			continue
		}
		t.dir += pathSeparator + f.Name()
		e.work(t)
		t.dir = t.dir[:len(t.dir)-len(f.Name())-1]
	}
	t.isLast = t.isLast[:len(t.isLast)-1]
}

func (e *Exec) writeLine(t *Tree, fi os.FileInfo) {
	for _, l := range t.isLast[:len(t.isLast)-1] {
		if l {
			e.out.Write(blank)
		} else {
			e.out.Write(trunk)
		}
	}
	if t.isLast[len(t.isLast)-1] {
		e.out.Write(lastNode)
	} else {
		e.out.Write(node)
	}
	io.WriteString(e.out, fi.Name())
	if fi.Mode()&os.ModeSymlink != 0 {
		dst, err := os.Readlink(t.dir + pathSeparator + fi.Name())
		if err != nil {
			io.WriteString(e.errOut, err.Error())
		}
		e.out.Write(arrow)
		io.WriteString(e.out, dst)
	}
	e.out.Write(newLine)
}
