package tk

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	fileEmoji  = "\U0001f4c4"
	dirEmoji   = "\U0001f4c1"
	file       = "file"
	dir        = "dir"
	resetColor = "\x1b[0m"
)

var lineColors = []string{
	"\x1b[31m", // red
	"\x1b[32m", // green
	"\x1b[33m", // yellow
}

var nameColors = map[string]string{
	file: "\x1b[35m", // magenta
	dir:  "\x1b[36m", // cyan
}

// Node ..
type node struct {
	root  *node
	nodes []node
	depth int
	last  bool
	path  string
	total map[string]int
	info  os.FileInfo
}

func (n *node) buildTree(flags map[string]interface{}) error {
	entries, err := ioutil.ReadDir(n.path)
	if err != nil {
		return err
	}

	hasAll := flags["all"].(bool)
	hasWin := flags["win"].(bool)
	entries = exceptHiddens(hasAll, hasWin, n.path, entries)

	if subStr := flags["find"].(string); subStr != "" {
		entries = search(entries, subStr)
	}

	dirs := justDirs(entries)
	n.total[dir] = n.total[dir] + len(dirs)
	n.total[file] = n.total[file] + len(entries) - len(dirs)

	if justDir := flags["dir"].(bool); justDir {
		entries = dirs
	}

	for i, e := range entries {
		last := false
		if i+1 == len(entries) {
			last = true
		}
		_n := node{n, nil, n.depth + 1, last, fmt.Sprintf("%s%s", appendSeperator(n.path), e.Name()), n.total, e}
		if maxDepth := flags["level"].(int); e.Mode().IsDir() {
			if !(maxDepth != 0 && n.depth+1 >= maxDepth) {
				if err = _n.buildTree(flags); err != nil {
					return err
				}
			}
		}

		if !flags["trim"].(bool) || (!_n.info.IsDir() || len(_n.nodes) > 0) {
			n.nodes = append(n.nodes, _n)
		} else if last && len(n.nodes) > 0 {
			n.nodes[len(n.nodes)-1].last = true
		}
	}
	return nil
}

func (n node) draw(wr io.Writer, flags map[string]interface{}) {
	n.print(wr, flags)
	for _, _n := range n.nodes {
		if _n.nodes != nil {
			_n.draw(wr, flags)
		} else {
			_n.print(wr, flags)
		}
	}
}

func (n node) print(wr io.Writer, flags map[string]interface{}) {
	line := ""
	hasColor := flags["color"].(bool)
	outputFile := flags["output"].(string)
	if n.root != nil {
		for _n := n.root; _n.root != nil; _n = _n.root {
			prefix := fmt.Sprintf("%s%s", "│", strings.Repeat(" ", 3))
			if _n.last {
				prefix = strings.Repeat(" ", 4)
			}
			if hasColor && outputFile == "" {
				prefix = colorize(lineColors[_n.depth%len(lineColors)], prefix)
			}
			line = fmt.Sprintf("%s%s", prefix, line)
		}

		suffix := "├── "
		if n.last {
			suffix = "└── "
		}

		if flags["emoji"].(bool) && outputFile == "" {
			e := fileEmoji
			if n.info.IsDir() {
				e = dirEmoji
			}
			suffix = fmt.Sprintf("%s%s%s", suffix, e, strings.Repeat(" ", 2))
		}

		if hasColor && outputFile == "" {
			suffix = colorize(lineColors[n.depth%len(lineColors)], suffix)
		}
		line = fmt.Sprintf("%s%s", line, suffix)
	}

	name := filepath.Base(n.path)
	if flags["path"].(bool) || n.root == nil {
		name = n.path
	}

	nameColor := nameColors[file]
	if n.info.IsDir() {
		name = appendSeperator(name)
		nameColor = nameColors[dir]
	}

	meta := ""
	if flags["mode"].(bool) {
		meta = fmt.Sprintf("%v  %v", meta, n.info.Mode())
	}

	if flags["size"].(bool) {
		if flags["verbose"].(bool) {
			meta = fmt.Sprintf("%v  %v", meta, n.info.Size())
		} else {
			meta = fmt.Sprintf("%v  %v", meta, formatSize(n.info.Size()))
		}
	}

	if flags["time"].(bool) {
		meta = fmt.Sprintf("%v  %v", meta, n.info.ModTime().Format("2-Jan-06 15:04"))
	}

	if meta != "" {
		if strings.HasPrefix(meta, strings.Repeat(" ", 2)) {
			meta = meta[2:]
		}
		name = fmt.Sprintf("[%v] %v", meta, name)
	}

	if hasColor && outputFile == "" {
		name = colorize(nameColor, name)
	}

	fmt.Fprintf(wr, "%s%s\n", line, name)
}

// Tree ..
type tree struct {
	root  node
	flags map[string]interface{}
}

// DrawTree .. draws a tree map
func DrawTree(flags map[string]interface{}) {
	if runtime.GOOS == "windows" && flags["emoji"].(bool) {
		fmt.Println("no emoji support on windows!")
		os.Exit(1)
	}
	rootPath := flags["root"].(string)
	info, err := os.Stat(rootPath)
	if os.IsNotExist(err) {
		fmt.Println("no such directory:", rootPath)
		os.Exit(1)
	}
	tree := tree{node{nil, nil, 0, false, rootPath, map[string]int{dir: 0, file: 0}, info}, flags}
	tree.draw()

	os.Exit(0)
}

func (t tree) draw() {
	if err := t.root.buildTree(t.flags); err != nil {
		fmt.Println(err)
	}

	if outputFile := t.flags["output"].(string); outputFile != "" {
		buf := new(bytes.Buffer)
		t.root.draw(buf, t.flags)
		if err := writeToFile(buf.String(), outputFile); err != nil {
			fmt.Println(err)
		}
	} else {
		t.root.draw(os.Stdout, t.flags)
	}

	if t.flags["number"].(bool) {
		fmt.Printf("\ntotal directories: %v, total files: %v\n", t.root.total[dir], t.root.total[file])
	}
}

func isHidden(path, fileName string) bool {
	// if runtime.GOOS == "windows" {
	// 	pointer, err := syscall.UTF16PtrFromString(filepath.Join(path, fileName))
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	attributes, err := syscall.GetFileAttributes(pointer)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0
	// }
	return false
}

func search(files []os.FileInfo, subStr string) []os.FileInfo {
	result := []os.FileInfo{}
	for _, f := range files {
		if f.IsDir() || strings.Contains(f.Name(), subStr) {
			result = append(result, f)
		}
	}
	return result
}

func justDirs(files []os.FileInfo) []os.FileInfo {
	dirs := []os.FileInfo{}
	for _, f := range files {
		if f.IsDir() {
			dirs = append(dirs, f)
		}
	}
	return dirs
}

func appendSeperator(path string) string {
	if !strings.HasSuffix(path, string(os.PathSeparator)) {
		return fmt.Sprintf("%s%s", path, string(os.PathSeparator))
	}
	return path
}

func formatSize(s int64) string {
	GB := int64(1024 * 1024 * 1024)
	MB := int64(1024 * 1024)
	KB := int64(1024)
	unit := "B"
	amount := 0.
	if s > GB {
		amount = math.Round(float64(s/GB)*100) / 100
		unit = "G"
	}
	if s > MB {
		amount = math.Round(float64(s/MB)*100) / 100
		unit = "M"
	}
	if s > KB {
		amount = math.Round(float64(s/KB)*100) / 100
		unit = "K"
	}
	return fmt.Sprintf("%v%v", amount, unit)
}

func exceptHiddens(all, win bool, path string, files []os.FileInfo) []os.FileInfo {
	result := []os.FileInfo{}
	for _, file := range files {
		if !all && strings.HasPrefix(file.Name(), ".") {
			continue
		}
		if !win && isHidden(path, file.Name()) {
			continue
		}
		result = append(result, file)
	}
	return result
}

func writeToFile(output, outputFile string) (err error) {
	file, err := os.Create(outputFile)
	if err != nil {
		return
	}
	defer file.Close()

	if _, err = io.WriteString(file, output); err != nil {
		return
	}
	if err = file.Sync(); err != nil {
		return
	}
	return
}

func colorize(color, str string) string {
	return fmt.Sprintf("%s%s%s", color, str, resetColor)
}
