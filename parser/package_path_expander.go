package parser

import (
	"os"
	"strings"

	"github.com/spf13/afero"
)

const MaxLevelInfinity = -1

type packagePathExpander struct {
	maxLevels int

	relativePackagePaths []string
}

func (p *packagePathExpander) appendRelPath(relPath string) {
	relPath = strings.Replace(relPath, "\\", "/", MaxLevelInfinity)
	p.relativePackagePaths = append(p.relativePackagePaths, relPath)
}

func (p *packagePathExpander) calcLevel(relPath string) (level int) {
	relPath = strings.Replace(relPath, "\\", "/", MaxLevelInfinity)
	relPath = strings.Trim(relPath, "/")
	level = len(strings.Split(relPath, "/"))
	return
}

func (p *packagePathExpander) WalkDir(baseFS afero.Fs) error {
	return afero.Walk(baseFS, "", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}
		if path == "" {
			p.appendRelPath(path)
			//Base path was already added at the begging of Walk
			return nil
		}

		level := p.calcLevel(path)
		if p.maxLevels != MaxLevelInfinity && level > p.maxLevels {
			return nil
		}

		p.appendRelPath(path)
		return nil
	})
}

func (p *packagePathExpander) PrefixedRelativePaths(prefix string) (paths []string) {
	paths = []string{}
	for _, path := range p.relativePackagePaths {
		paths = append(paths, strings.TrimRight(prefix, "/")+"/"+strings.TrimLeft(path, "/"))
	}
	return
}
