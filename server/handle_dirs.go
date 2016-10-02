package server

import (
	"github.com/jehiah/go-strftime"
	"path"
	"strings"
	"fmt"
	"os"
	"io"
)

func (c *clientHandler) absPath(p string) string {
	path := c.Path()

	if strings.HasPrefix(p, "/") {
		path = p
	} else {
		if path != "/" {
			path += "/"
		}
		path += p
	}

	return path
}

func (c *clientHandler) handleCWD() {
	if c.param == ".." {
		c.handleCDUP()
		return
	}

	path := c.absPath(c.param)

	// TODO: Find something smarter, this is obviously quite limitating...
	if path == "/debug" {
		c.writeMessage(250, "Debug activated !")
		c.debug = true
		return
	}

	if err := c.driver.ChangeDirectory(c, path); err == nil {
		c.SetPath(path)
		c.writeMessage(250, fmt.Sprintf("CD worked on %s", path))
	} else {
		c.writeMessage(550, fmt.Sprintf("CD issue: %s", err.Error()))
	}
}

func (c *clientHandler) handleMKD() {
	path := c.absPath(c.param)
	if err := c.driver.MakeDirectory(c, path); err == nil {
		c.writeMessage(250, fmt.Sprintf("Created dir %s", path))
	} else {
		c.writeMessage(550, fmt.Sprintf("Could not create %s : %s", path, err.Error()))
	}
}

func (c *clientHandler) handleRMD() {
	path := c.absPath(c.param)
	if err := c.driver.DeleteFile(c, path); err == nil {
		c.writeMessage(250, fmt.Sprintf("Deleted dir %s", path))
	} else {
		c.writeMessage(550, fmt.Sprintf("Could not delete dir %s : %s", path, err.Error()))
	}
}

func (c *clientHandler) handleCDUP() {
	parent, _ := path.Split(c.Path())
	if parent != "/" && strings.HasSuffix(parent, "/") {
		parent = parent[0:len(parent)-1]
	}
	if err := c.driver.ChangeDirectory(c, parent); err == nil {
		c.SetPath(parent)
		c.writeMessage(250, fmt.Sprintf("CDUP worked on %s", parent))
	} else {
		c.writeMessage(550, fmt.Sprintf("CDUP issue: %s", err.Error()))
	}
}

func (c *clientHandler) handlePWD() {
	c.writeMessage(257, "\"" + c.Path() + "\" is the current directory")
}

func (c *clientHandler) handleLIST() {
	if files, err := c.driver.ListFiles(c); err == nil {
		if tr, err := c.TransferOpen(); err == nil {
			defer c.TransferClose()
			c.dirList(tr, files)
		}
	} else {
		c.writeMessage(500, "Could not list - "+err.Error())
	}
}

func (c *clientHandler) dirList(w io.Writer, files []os.FileInfo) error {
	for _, file := range files {
		fmt.Fprint(w, file.Mode().String())
		fmt.Fprintf(w, " 1 %s %s ", "ftp", "ftp")
		fmt.Fprintf(w, "%12d", file.Size())
		fmt.Fprintf(w, strftime.Format(" %b %d %H:%M ", file.ModTime()))
		fmt.Fprintf(w, "%s\r\n", file.Name())
	}
	fmt.Fprint(w, "\r\n")
	return nil
}